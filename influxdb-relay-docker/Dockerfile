FROM golang:1.9

MAINTAINER Mutlu Polatcan <mutlupolatcan@gmail.com>

ENV HTTP_NAME="relay-http"
ENV HTTP_BIND_ADDR="0.0.0.0:9086"
ENV HTTP_TIMEOUT="10s"
ENV HTTP_BUFFER_SIZE_MB="NULL"
ENV HTTP_MAX_BATCH_KB="NULL"
ENV HTTP_MAX_DELAY_INTERVAL="NULL"
ENV HTTP_SKIP_TLS_VERIFICATION="NULL"
ENV HTTP_DEFAULT_RETENTION_POLICY="NULL"
ENV SR_ADDR="NULL"

RUN go get -u github.com/influxdata/influxdb-relay

EXPOSE 3030

WORKDIR /

COPY sr_observer ./
RUN chmod +x /sr_observer

ENTRYPOINT ["./sr_observer"]
CMD influxd
