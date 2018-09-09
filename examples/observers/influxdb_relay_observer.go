package main

import (
	"../../core"
	"os"
	"../../plugins/influxdb-relay"
	"net/http"
)

func main() {
	var observer = &core.Observer{}
	var influxDBRelay = new(influxdb_relay.InfluxDBRelay)
	observer.StartObserver(os.Getenv("SR_ADDR"),"influxdb",os.Getenv("HOSTNAME"),3030,"/update", influxDBRelay.Update,
		[]core.ObserverHandler{
			{ Endpoint: "/query", Handler: influxDBRelay.QueryHandler },
			{ Endpoint: "/write", Handler: func(w http.ResponseWriter, r *http.Request) {
				// TODO Redirect request to InfluxDB Relay's write endpoint
			}},
		}...)
}
