package core

import (
	"encoding/json"
	"log"
	"net/http"
	"bytes"
	"strconv"
)

type Observer struct {
	RegistryAddr string
	ObservedServices string
	ObserverHostname string
	ObserverPort int
	ObserverUpdateEndpoint string
	UpdaterFunction func(r *http.Request)
}

func (observer *Observer) StartObserver(registryAddr string, observedServices string, observerHostname string, observerPort int, observerUpdateEndpoint string, updaterFunction func(r *http.Request)) {
	observer.RegistryAddr = registryAddr
	observer.ObservedServices = observedServices
	observer.ObserverHostname = observerHostname
	observer.ObserverPort = observerPort
	observer.ObserverUpdateEndpoint = observerUpdateEndpoint
	observer.UpdaterFunction = updaterFunction

	observer.Connect()

	http.HandleFunc(observerUpdateEndpoint, observer.UpdateHandler)
	log.Println(http.ListenAndServe(":" + strconv.Itoa(observer.ObserverPort), nil))
}

func (observer *Observer) Connect() {
	log.Println("Connecting to registry...")

	for {
		jsonData, err := json.Marshal(map[string]interface{}{"observedServices": observer.ObservedServices,
															 "observerHostname": observer.ObserverHostname,
															 "observerPort": observer.ObserverPort,
															 "observerUpdateEndpoint": observer.ObserverUpdateEndpoint})

		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(jsonData))
		}

		_, err = http.Post("http://"+ observer.RegistryAddr +"/register?type=observer", "application/json", bytes.NewBufferString(string(jsonData)))

		if err == nil {
			log.Println("Connection established with registry!")
			break
		}
	}
}

func (observer *Observer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	observer.UpdaterFunction(r)
}

