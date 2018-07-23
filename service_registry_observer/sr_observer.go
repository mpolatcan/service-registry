package main

import (
	"net/http"
	"log"
	"container/list"
	"os"
	"os/exec"
	"strconv"
	"github.com/naoina/toml"
	"io/ioutil"
	"bytes"
)

// TODO Sharding logic
// TODO Query handling
// TODO Communication with service service_registry-docker

var hostnames = list.New()
var relayPid = 0

type Service struct {
	ServiceName string
	ServiceHostname string
	ServicePort int
	ServiceHeartbeatEndpoint string
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
	} else {
		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(data))
		}
	}
}

func updateInfluxDBRelay() {
	log.Println("Updating configuration and restarting InfluxDB Relay")

	if relayPid != 0 {
		exec.Command("kill","-9", string(relayPid)).Run()
	}

	file, err := os.Create("/config.toml")

	if err != nil {
		panic("File could not be opened")
	} else {
		httpConf := HTTPConfig{}
		httpOutputConf := make([]HTTPOutputConfig, hostnames.Len())

		httpConf.Name = os.Getenv("HTTP_NAME")
		httpConf.Addr = os.Getenv("HTTP_BIND_ADDR")

		if os.Getenv("HTTP_DEFAULT_RETENTION_POLICY") != "NULL" {
			httpConf.DefaultRetentionPolicy = os.Getenv("HTTP_DEFAULT_RETENTION_POLICY")
		}

		index := 0
		for hostname := hostnames.Front(); hostname != nil; hostname = hostname.Next() {
			httpOutputConf[index].Name = hostname.Value.(string)
			httpOutputConf[index].Location = "http://" + hostname.Value.(string) + ":8086/write"
			httpOutputConf[index].Timeout = os.Getenv("HTTP_TIMEOUT")

			if os.Getenv("HTTP_BUFFER_SIZE_MB") != "NULL" {
				httpOutputConf[index].BufferSizeMB, err = strconv.Atoi(os.Getenv("HTTP_BUFFER_SIZE_MB"))
			}

			if os.Getenv("HTTP_MAX_BATCH_KB") != "NULL" {
				httpOutputConf[index].MaxBatchKB, err = strconv.Atoi(os.Getenv("HTTP_MAX_BATCH_KB"))
			}

			if os.Getenv("HTTP_MAX_DELAY_INTERVAL") != 	"NULL" {
				httpOutputConf[index].MaxDelayInterval = os.Getenv("HTTP_MAX_DELAY_INTERVAL")
			}

			if os.Getenv("HTTP_SKIP_TLS_VERIFICATION") != "NULL" {
				httpOutputConf[index].SkipTLSVerification, err = strconv.ParseBool(os.Getenv("HTTP_SKIP_TLS_VERIFICATION"))
			}
			index++
		}

		httpConf.Outputs = httpOutputConf
		err = toml.NewEncoder(file).Encode(&httpConf)

		if err != nil {
			log.Println(err)
		}

		file.Sync()
		file.Close()

		cmd := exec.Command("influxdb-relay","-config","/config.toml")
		err := cmd.Start()

		if err != nil {
			log.Println(err)
		} else {
			relayPid = cmd.Process.Pid
			log.Printf("InfluxDB Relay started with PID %d", relayPid)
		}
	}
}

func main() {
	for {
		_, err := http.Post("http://" + os.Getenv("SR_ADDR") + "/register?type=observer", "application/json",
			bytes.NewBufferString("{\"observedServices\": \"influxdb\", \"observerHostname\":\"" + os.Getenv("HOSTNAME") + "\", \"observerPort\": 3030, \"observerUpdateEndpoint\": \"/update\"}"))

		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}

	http.HandleFunc("/update", updateHandler)
	log.Fatal(http.ListenAndServe(":3030", nil))
}
