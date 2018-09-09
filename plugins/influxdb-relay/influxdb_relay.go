package influxdb_relay

import (
	"os"
	"log"
	"os/exec"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"../../core"
	"fmt"
	"net/url"
)

type InfluxDBRelay struct {
	Nodes []core.Service
	RelayPid int
}

func (influxdb_relay *InfluxDBRelay) QueryHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body); if err != nil { log.Println(err)}
	values, err := url.ParseQuery(r.URL.RawQuery)

	log.Println(data)
	log.Println(values)

	if len(influxdb_relay.Nodes) > 0 {
		for _, node := range influxdb_relay.Nodes {
			// TODO Get queries and send to InfluxDB nodes
			log.Println(node)
		}
	}
}

func (influxdb_relay *InfluxDBRelay) ConfigWriter(file *os.File, Nodes []core.Service) {
	file.WriteString("[[http]]\n")
	file.WriteString(fmt.Sprintf("name = \"%s\"\n", os.Getenv("HTTP_NAME")))
	file.WriteString(fmt.Sprintf("bind-addr = \"%s\"\n", os.Getenv("HTTP_BIND_ADDR")))
	if os.Getenv("HTTP_DEFAULT_RETENTION_POLICY") != "NULL" { file.WriteString(fmt.Sprintf("default-retention-policy = %s", os.Getenv("HTTP_DEFAULT_RETENTION_POLICY"))) }
	file.WriteString("output = [\n")

	for index, node := range Nodes {
		file.WriteString(fmt.Sprintf("   { name=\"%s%d\", location=\"http://%s:%d/write\", timeout=\"%s\" ", node.ServiceName, index, node.ServiceHostname, 8086, os.Getenv("HTTP_TIMEOUT")))
		if os.Getenv("HTTP_BUFFER_SIZE_MB") != "NULL" { file.WriteString(fmt.Sprintf(", buffer-size-mb=%s, ", os.Getenv("HTTP_BUFFER_SIZE_MB"))) }
		if os.Getenv("HTTP_MAX_BATCH_KB") != "NULL" { file.WriteString(fmt.Sprintf(", max-batch-kb=%s, ", os.Getenv("HTTP_MAX_BATCH_KB"))) }
		if os.Getenv("HTTP_MAX_DELAY_INTERVAL") != "NULL" { file.WriteString(fmt.Sprintf(", max-delay-interval=\"%s\"", os.Getenv("HTTP_MAX_DELAY_INTERVAL"))) }
		if os.Getenv("HTTP_SKIP_TLS_VERIFICATION") != "NULL" { file.WriteString(fmt.Sprintf(", skip-tls-verification=%s ", os.Getenv("HTTP_SKIP_TLS_VERIFICATION")))}
		file.WriteString("}")

		if index != len(Nodes) - 1 {
			file.WriteString(",\n")
		}
	}

	file.WriteString("\n]")
	file.Sync()
	file.Close()
}

func (influxdb_relay *InfluxDBRelay) Update(r *http.Request) {
	log.Println("Updating configuration and restarting InfluxDB Relay")

	data, err := ioutil.ReadAll(r.Body)
	var Nodes = make([]core.Service, 0)

	if err != nil {
		log.Println(err)
	} else {
		err := json.Unmarshal(data,&Nodes); if err != nil { log.Println(err) }
	}

	influxdb_relay.Nodes = Nodes

	if influxdb_relay.RelayPid != 0 {
		exec.Command("kill","-9", string(influxdb_relay.RelayPid)).Run()
	}

	file, err := os.Create("/config.toml")

	if err != nil {
		panic("File could not be opened")
	} else {
		influxdb_relay.ConfigWriter(file,Nodes)

		cmd := exec.Command("influxdb-relay","-config","/config.toml")
		err := cmd.Start()

		if err != nil {
			log.Println(err)
		} else {
			influxdb_relay.RelayPid = cmd.Process.Pid
			log.Printf("InfluxDB Relay started with PID %d", influxdb_relay.RelayPid)
		}
	}
}
