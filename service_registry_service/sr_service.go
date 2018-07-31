package main

import (
	"../service_registry/core"
	"os"
)

func main() {
	var service = &core.Service{}
	service.StartService(os.Getenv("SR_ADDR"),"influxdb",os.Getenv("HOSTNAME"),8086,"/ping")
}