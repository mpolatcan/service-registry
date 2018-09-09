FROM influxdb

ENV SR_ADDR="NULL"

MAINTAINER Mutlu Polatcan <mutlupolatcan@gmail.com>

WORKDIR /

COPY influxdb_service ./
RUN chmod +x /influxdb_service

ENTRYPOINT ["/entrypoint.sh"]
CMD ./influxdb_service && influxd