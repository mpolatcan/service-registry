package main

import (
	"../../core"
	"os"
)

func main() {
	var service = &core.Service{}
	service.StartService(
		os.Getenv("SR_ADDR"), "influxdb",os.Getenv("HOSTNAME"),8086,"/ping")
}