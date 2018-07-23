package core

import (
	"container/list"
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"net/url"
	"io/ioutil"
	"fmt"
	"strconv"
	"time"
	"os"
	"bytes"
	"sync"
)

// TODO Count failures
// TODO Updating observers

type Registry struct {
	Services map[string]*list.List

	Observers map[string]*list.List

	HealthcheckerChannels map[string]chan bool

	FailureCounts map[string]int

	Mutex *sync.Mutex
}

func (registry *Registry) InitRegistry()  {
	registry.Services = make(map[string]*list.List)
	registry.Observers = make(map[string]*list.List)
	registry.HealthcheckerChannels = make(map[string]chan bool)
	registry.FailureCounts = make(map[string]int)
	registry.Mutex = &sync.Mutex{}
}

func (registry *Registry) GetServices(serviceName string) *list.List {
	return registry.Services[serviceName]
}

func (registry *Registry) GetServiceCount(serviceName string) int {
	return registry.GetServices(serviceName).Len()
}

func (registry *Registry) AllocateServiceList(serviceName string) {
	if registry.Services[serviceName] == nil {
		registry.Services[serviceName] = list.New()
	}
}

func (registry *Registry) AddService(serviceName string, service Service) {
	registry.Services[serviceName].PushBack(service)
}

func (registry *Registry) RemoveService(serviceName string, service *list.Element) {
	registry.Services[serviceName].Remove(service)
}

func (registry *Registry) GetObservers(serviceName string) *list.List {
	return registry.Observers[serviceName]
}

func (registry *Registry) GetObserverCount(serviceName string) int {
	return registry.GetObservers(serviceName).Len()
}

func (registry *Registry) AllocateObserverList(serviceName string) {
	if registry.Observers[serviceName] == nil {
		registry.Observers[serviceName] = list.New()
	}
}

func (registry *Registry) AddObserver(serviceName string, observer Observer) {
	registry.Observers[serviceName].PushBack(observer)
}

func (registry *Registry) RemoveObserver(serviceName string, observer *list.Element) {
	registry.Observers[serviceName].Remove(observer)
}

func (registry *Registry) GetHealthcheckerChannel(serviceName string) chan bool {
	return registry.HealthcheckerChannels[serviceName]
}

func (registry *Registry) CreateHealthcheckerChannel(serviceName string) {
	if registry.HealthcheckerChannels[serviceName] == nil {
		registry.HealthcheckerChannels[serviceName] = make(chan bool)
	}
}

func (registry *Registry) Wait(durationStr string) {
	duration, err := time.ParseDuration(durationStr)

	if err != nil {
		panic(err)
	} else {
		time.Sleep(duration)
	}
}

func (registry *Registry) CreateHealthchecker(service Service) {
	if registry.GetHealthcheckerChannel(service.ServiceName) == nil {
		registry.CreateHealthcheckerChannel(service.ServiceName)

		go func() {
			for {
				ok := <- registry.GetHealthcheckerChannel(service.ServiceName)

				if !ok {
					log.Printf("Deactivating healthchecker of %s\n", service.ServiceName)
					break
				}

				for service := registry.GetServices(service.ServiceName).Front(); service != nil; service = service.Next() {
					serviceValue := service.Value.(Service)

					response, err := http.Get("http://" + serviceValue.ServiceHostname + ":" + strconv.Itoa(serviceValue.ServicePort) + serviceValue.ServiceHeartbeatEndpoint)

					if err != nil {
						log.Println(err)

						// TODO Change retry healthcheck
						go func() {
							registry.RetryHealthcheck(service)
						}()
					} else {
						status := response.StatusCode

						if status == 204 {
							log.Printf("Status Code: %d -> Node %s is alive!\n", status, serviceValue.ServiceHostname)
						} else {
							log.Printf("Status Code: %d -> Node %s is dead!\n", status, serviceValue.ServiceHostname)
							log.Printf("Removing node %s from registry...\n", serviceValue.ServiceHostname)
							registry.RemoveService(serviceValue.ServiceName, service)
							// TODO Delete key if there is no service
						}
					}
				}
			}
		}()
	}
}

func (registry *Registry) RetryHealthcheck(service *list.Element) {
	serviceValue := service.Value.(Service)

	threshold, err := strconv.Atoi(os.Getenv("SR_HEALTHCHECK_FAILURE_THRESHOLD"))

	if err != nil {
		log.Println(err)
	}

	success := false

	for try := 0; try < threshold; try++ {
		log.Printf("Retrying %d times\n", try)

		response, err := http.Get("http://" + serviceValue.ServiceHostname + ":" + strconv.Itoa(serviceValue.ServicePort) + serviceValue.ServiceHeartbeatEndpoint)

		if err != nil {
			log.Println(err)
		} else {
			status := response.StatusCode

			if status == 204 {
				log.Printf("Status Code: %d -> Node %s is alive!\n", status, serviceValue.ServiceHostname)
				success = true
				break
			}
		}

		registry.Wait(os.Getenv("SR_HEALTHCHECK_INTERVAL"))
	}

	if !success {
		log.Printf("Node %s is dead!\n", serviceValue.ServiceHostname)
		log.Printf("Removing node %s from registry...\n", serviceValue.ServiceHostname)
		registry.RemoveService(serviceValue.ServiceName,service)
	}
}

func (registry *Registry) ManageHealthcheckers() {
	go func() {
		for {
			for service, serviceList := range registry.Services {
				if serviceList != nil {
					if serviceList.Len() > 0 {
						registry.GetHealthcheckerChannel(service) <- true
					} else {
						if registry.GetHealthcheckerChannel(service) != nil {
							registry.GetHealthcheckerChannel(service) <- false
							delete(registry.HealthcheckerChannels, service)
						}
					}
				}
			}

			registry.Wait(os.Getenv("SR_HEALTHCHECK_INTERVAL"))
		}
	}()
}

func (registry *Registry) UpdateObservers(serviceName string) {
	if registry.GetObservers(serviceName) != nil {
		for observer := registry.GetObservers(serviceName).Front(); observer != nil; observer = observer.Next() {
			observerValue := observer.Value.(Observer)
			_, err := http.Post("http://" + observerValue.ObserverHostname + ":" + strconv.Itoa(observerValue.ObserverPort) + observerValue.ObserverUpdateEndpoint, "application/text", bytes.NewBufferString("UPDATE!"))

			if err != nil {
				log.Println(err)
			}
		}
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

			registry.Mutex.Lock()
			// --------------------- CRITICAL SECTION ------------------
			registry.PrintInfo(err, fmt.Sprintf("New service has registered -> %+v\n", service))

			registry.AllocateServiceList(service.ServiceName)
			registry.AddService(service.ServiceName, service)

			registry.CreateHealthchecker(service)

			registry.UpdateObservers(service.ServiceName)
			// --------------------------------------------------------------
			registry.Mutex.Unlock()
		} else if queryValues.Get("type") == "observer" {
			observer := Observer{}
			err := json.Unmarshal(data,&observer)

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
			foundedServicesMap := make(map[string][]Service)

			var success = true

			for _, requestedService := range requestedServices {
				foundedServices := registry.GetServices(requestedService)

				if foundedServices == nil {
					success = false
					fmt.Fprintf(w, "There is no service named \"%s\"", requestedService)
					break
				} else {
					foundedServicesMap[requestedService] = make([]Service, foundedServices.Len())

					index := 0
					for service := foundedServices.Front(); service != nil; service = service.Next() {
						serviceValue := service.Value.(Service)
						foundedServicesMap[requestedService][index].ServiceName = serviceValue.ServiceName
						foundedServicesMap[requestedService][index].ServiceHeartbeatEndpoint = serviceValue.ServiceHeartbeatEndpoint
						foundedServicesMap[requestedService][index].ServiceHostname = serviceValue.ServiceHostname
						foundedServicesMap[requestedService][index].ServicePort = serviceValue.ServicePort
						index++
					}
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


