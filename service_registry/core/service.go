package core

import (
	"encoding/json"
	"log"
	"net/http"
	"bytes"
)

// TODO Connection timeout will be added
// TODO Optional healthcheck mechanism for services

type Service struct {
	RegistryAddr string
	ServiceName string
	ServiceHostname string
	ServicePort int
	ServiceHeartbeatEndpoint string
}

func (service *Service) StartService(registryAddr string, serviceName string, serviceHostname string, servicePort int, serviceHeartbeatEndpoint string) {
	service.RegistryAddr = registryAddr
	service.ServiceName = serviceName
	service.ServiceHostname = serviceHostname
	service.ServicePort = servicePort
	service.ServiceHeartbeatEndpoint = serviceHeartbeatEndpoint

	service.Connect()
}

func (service *Service) Connect() {
	log.Println("Connecting to registry...")
	for {
		jsonData, err := json.Marshal(map[string]interface{}{"serviceName": service.ServiceName,
															 "serviceHostname": service.ServiceHostname,
															 "servicePort": service.ServicePort,
															 "serviceHeartbeatEndpoint": service.ServiceHeartbeatEndpoint})

		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(jsonData))
		}


		_, err = http.Post("http://"+ service.RegistryAddr +"/register?type=service", "application/json", bytes.NewBufferString(string(jsonData)))

		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}
}