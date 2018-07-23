package main

import (
	"net/http"
	"os"
	"bytes"
	"log"
)

func main() {
	for {
		_, err := http.Post("http://" + os.Getenv("SR_ADDR") + "/register?type=service", "application/json",
			bytes.NewBufferString("{\"serviceName\": \"influxdb\", \"serviceHostname\":\"" + os.Getenv("HOSTNAME") + "\", \"servicePort\": 8086, \"serviceHeartbeatEndpoint\": \"/ping\"}"))

		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}
}