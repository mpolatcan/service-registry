.PHONY: server
server:
	sudo env GOOS=linux GOARCH=arm go build -o ./docker/influxdb-relay-docker/influxdb_relay_observer ./examples/observers/influxdb_relay_observer.go
	sudo docker build -t mpolatcan/influxdb-relay ./docker/influxdb-relay-docker/

.PHONY: client
client:
	sudo env GOOS=linux GOARCH=arm go build -o ./docker/influxdb-docker/influxdb_service ./examples/services/influxdb_service.go
	sudo docker build -t mpolatcan/influxdb ./docker/influxdb-docker/

.PHONY: registry
registry:
	sudo env GOOS=linux GOARCH=arm go build -o ./docker/service-registry-docker/service_registry main.go
	sudo docker build -t mpolatcan/service-registry ./docker/service-registry-docker/

.PHONY: compose
compose:
	sudo docker-compose up --scale influxdb=3

all:
	$(MAKE) server client registry compose