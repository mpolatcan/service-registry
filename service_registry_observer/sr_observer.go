package main

import (
	"../service_registry/core"
	"os"
	"../service_registry/plugins/influxdb"
)

func main() {
	var observer = &core.Observer{}
	observer.StartObserver(os.Getenv("SR_ADDR"),"influxdb",os.Getenv("HOSTNAME"),3030,"/update", new(influxdb.InfluxDB).Update)
}
