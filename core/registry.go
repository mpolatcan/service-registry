/* Written by Mutlu Polatcan
   23.07.2018

	Simple, publisher-subscriber based service registry for service discovery and dynamic configuration update
 */

package core

import (
	"sync"
	"time"
	"net/http"
	"strconv"
	"log"
	"io/ioutil"
	"net/url"
	"encoding/json"
	"fmt"
	"strings"
	"os"
	"bytes"
)

type Registry struct {
	Services map[string]*List

	Observers map[string]*List

	HealthCheckerStatus map[string]bool

	FailureCounts map[string]int

	Mutex *sync.Mutex

	FailureThreshold int

	HealthCheckInterval string

	InitialDelay string
}

func (registry *Registry) StartRegistry()  {
	registry.Services = make(map[string]*List)
	registry.Observers = make(map[string]*List)
	registry.HealthCheckerStatus = make(map[string]bool)
	registry.FailureCounts = make(map[string]int)
	registry.Mutex = &sync.Mutex{}
	registry.FailureThreshold, _ = strconv.Atoi(os.Getenv("SR_HEALTHCHECK_FAILURE_THRESHOLD"))
	registry.HealthCheckInterval = os.Getenv("SR_HEALTHCHECK_INTERVAL")
	registry.InitialDelay = os.Getenv("SR_HEALTHCHECK_INITIAL_DELAY")

	http.HandleFunc("/register", registry.RegisterHandler)
	http.HandleFunc("/services", registry.ServicesHandler)
	log.Fatal(http.ListenAndServe(":" + os.Getenv("SR_PORT"), nil))
}

func (registry *Registry) GetServices(serviceName string) *List {
	return registry.Services[serviceName]
}

func (registry *Registry) DeleteServiceGroup(serviceName string) {
	if registry.GetServiceCount(serviceName) == 0 {
		delete(registry.Services, serviceName)
	}
}

func (registry *Registry) GetServiceCount(serviceName string) int {
	return registry.GetServices(serviceName).Len()
}

func (registry *Registry) AllocateServiceList(serviceName string) {
	if registry.Services[serviceName] == nil {
		registry.Services[serviceName] = new(List).Init()
	}
}

func (registry *Registry) AddService(serviceName string, service Service) {
	registry.Services[serviceName].Add(service)
}

func (registry *Registry) RemoveService(serviceName string, service Service) {
	registry.Services[serviceName].Remove(service)
}

func (registry *Registry) GetObservers(serviceName string) *List {
	return registry.Observers[serviceName]
}

func (registry *Registry) DeleteObserverGroup(serviceName string) {
	delete(registry.Observers, serviceName)
}

func (registry *Registry) GetObserverCount(serviceName string) int {
	return registry.GetObservers(serviceName).Len()
}

func (registry *Registry) AllocateObserverList(serviceName string) {
	if registry.Observers[serviceName] == nil {
		registry.Observers[serviceName] = new(List).Init()
	}
}

func (registry *Registry) AddObserver(serviceName string, observer Observer) {
	registry.Observers[serviceName].Add(observer)
}

func (registry *Registry) RemoveObserver(serviceName string, observer Observer) {
	registry.Observers[serviceName].Remove(observer)
}

func (registry *Registry) ChangeHealthCheckerStatus(serviceName string, status bool) {
	registry.HealthCheckerStatus[serviceName] = status
}

func (registry *Registry) IsHealthCheckerActivated(serviceName string) bool {
	return registry.HealthCheckerStatus[serviceName]
}

func (registry *Registry) GetFailureCount(hostname string) int {
	return registry.FailureCounts[hostname]
}

func (registry *Registry) IncrementFailureCount(hostname string) {
	registry.FailureCounts[hostname]++
}

func (registry *Registry) ResetFailureCount(hostname string) {
	registry.FailureCounts[hostname] = 0
}

func (registry *Registry) DeleteFailureCount(hostname string) {
	delete(registry.FailureCounts, hostname)
}

func (registry *Registry) Wait(durationStr string) {
	duration, err := time.ParseDuration(durationStr)

	if err != nil {
		panic(err)
	} else {
		time.Sleep(duration)
	}
}

func (registry *Registry) StartHealthChecker(service Service) {
	if !registry.IsHealthCheckerActivated(service.ServiceName) {
		registry.ChangeHealthCheckerStatus(service.ServiceName, true)

		go func() {
			registry.Wait(registry.InitialDelay)

			for ; registry.GetServiceCount(service.ServiceName) > 0; {

				for _, service := range registry.GetServices(service.ServiceName).Values() {
					serviceValue := service.(Service)

					response, err := http.Get(fmt.Sprintf("http://%s:%d%s",serviceValue.ServiceHostname,serviceValue.ServicePort,serviceValue.ServiceHeartbeatEndpoint))

					if err != nil {
						registry.IncrementFailureCount(serviceValue.ServiceHostname)

						if registry.GetFailureCount(serviceValue.ServiceHostname) > 0 {
							log.Printf("Retrying %d times for %s-%s\n\n", registry.GetFailureCount(serviceValue.ServiceHostname), serviceValue.ServiceName, serviceValue.ServiceHostname)
						}

						log.Println(err)

						if registry.GetFailureCount(serviceValue.ServiceHostname) >= registry.FailureThreshold {
							log.Printf("Node %s is dead! Removing node %s from registry...\n\n", serviceValue.ServiceHostname, serviceValue.ServiceHostname)
							registry.RemoveService(serviceValue.ServiceName, serviceValue)
							registry.DeleteFailureCount(serviceValue.ServiceHostname)
							registry.UpdateObservers(serviceValue.ServiceName)
						}
					} else {
						status := response.StatusCode

						if status == 204 {
							registry.ResetFailureCount(serviceValue.ServiceHostname)
							log.Printf("Status Code: %d -> Node %s is alive!\n\n", status, serviceValue.ServiceHostname)
						} else {
							log.Printf("Status Code: %d -> Node %s is dead!\n\n", status, serviceValue.ServiceHostname)
							log.Printf("Removing node %s from registry...\n\n", serviceValue.ServiceHostname)
							registry.RemoveService(serviceValue.ServiceName, serviceValue)
							registry.DeleteFailureCount(serviceValue.ServiceHostname)
							registry.UpdateObservers(serviceValue.ServiceName)
						}
					}
				}

				registry.Wait(registry.HealthCheckInterval)
			}

			log.Printf("Deactivating healthchecker of %s\n\n", service.ServiceName)
			registry.DeleteServiceGroup(service.ServiceName)
			registry.ChangeHealthCheckerStatus(service.ServiceName, false)
		}()
	}
}

func (registry *Registry) UpdateObservers(serviceName string) {
	if registry.GetObservers(serviceName) != nil {
		go func() {
			if registry.GetObservers(serviceName) != nil && registry.GetServices(serviceName) != nil {
				for _, observer := range registry.GetObservers(serviceName).Values() {
					// Retry forever
					for {
						servicesJSON, errJSON := json.Marshal(registry.GetServices(serviceName).Values()); if errJSON != nil { log.Println(errJSON) }

						observerValue := observer.(Observer)

						_, err := http.Post(fmt.Sprintf("http://%s:%d%s", observerValue.ObserverHostname, observerValue.ObserverPort, observerValue.ObserverUpdateEndpoint), "application/text", bytes.NewBuffer(servicesJSON))

						if err == nil {
							break
						}
					}
				}
			}
		}()
	}
}

func (registry *Registry) PrintInfo(err error, message string) {
	if err != nil {
		log.Println(err)
	} else {
		log.Println(message)
	}
}

func (registry *Registry) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	queryValues, queryErr := url.ParseQuery(r.URL.RawQuery)

	registry.PrintInfo(queryErr, "Query is valid!")

	if err != nil {
		log.Println(err)
	}  else {
		if queryValues.Get("type") == "service" {
			service := Service{}
			err := json.Unmarshal(data,&service)

			if err != nil {
				log.Println(err)
			}

			registry.Mutex.Lock()
			// --------------------- CRITICAL SECTION ------------------
			registry.PrintInfo(err, fmt.Sprintf("New service has registered -> %+v\n", service))

			registry.AllocateServiceList(service.ServiceName)

			registry.AddService(service.ServiceName, service)

			registry.ResetFailureCount(service.ServiceHostname)

			registry.StartHealthChecker(service)

			registry.UpdateObservers(service.ServiceName)
			// --------------------------------------------------------------
			registry.Mutex.Unlock()
		} else if queryValues.Get("type") == "observer" {
			observer := Observer{}
			err := json.Unmarshal(data,&observer)

			if err != nil {
				log.Println(err)
			}

			registry.Mutex.Lock()
			// --------------------- CRITICAL SECTION ------------------
			registry.PrintInfo(err, fmt.Sprintf("New observer has registered -> %+v\n", observer))

			observedServices := strings.Split(observer.ObservedServices, ",")

			for _, service := range observedServices {
				registry.AllocateObserverList(service)

				registry.AddObserver(service, observer)

				registry.UpdateObservers(service)
			}
			// --------------------------------------------------------------
			registry.Mutex.Unlock()
		}
	}
}

func (registry *Registry) ServicesHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		log.Println(err)
	}  else {
		queryParamService := queryValues.Get("service")

		if queryParamService == "" {
			fmt.Fprintf(w, "service parameter required on query!")
		} else {
			requestedServices := strings.Split(queryParamService, ",")
			foundedServicesMap := make(map[string][]interface{})

			var success = true

			for _, requestedService := range requestedServices {
				foundedServices := registry.GetServices(requestedService)

				if foundedServices == nil {
					success = false
					fmt.Fprintf(w, "There is no service named \"%s\"", requestedService)
					break
				} else {
					foundedServicesMap[requestedService] = registry.GetServices(requestedService).Values()
				}
			}

			if success {
				foundedServicesJSON, err := json.Marshal(foundedServicesMap)

				registry.PrintInfo(err, fmt.Sprintf("Founded services -> %+v", string(foundedServicesJSON)))

				fmt.Fprintf(w, "%s", string(foundedServicesJSON))
			}
		}
	}
}

func (registry *Registry) PingHandler(w http.ResponseWriter, r *http.Request) {
	// If registry is alive , send
	w.WriteHeader(200)
}