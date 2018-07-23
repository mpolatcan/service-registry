.PHONY: server
server:
	sudo env GOOS=linux GOARCH=arm go build -o ./influxdb-relay-docker/sr_observer ./service_registry_observer/sr_observer.go ./service_registry_observer/config.go
	sudo docker build -t mpolatcan/influxdb-relay ./influxdb-relay-docker/

.PHONY: client
client:
	sudo env GOOS=linux GOARCH=arm go build -o ./influxdb-docker/sr_service ./service_registry_service/sr_service.go
	sudo docker build -t mpolatcan/influxdb ./influxdb-docker/

.PHONY: registry
registry:
	sudo env GOOS=linux GOARCH=arm go build -o ./service-registry-docker/service_registry ./service_registry/main.go
	sudo docker build -t mpolatcan/service-registry ./service-registry-docker/

.PHONY: compose
compose:
	sudo docker-compose up --scale influxdb=3

all:
	$(MAKE) server client registry compose