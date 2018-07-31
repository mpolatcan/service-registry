FROM influxdb

ENV SR_ADDR="NULL"

MAINTAINER Mutlu Polatcan <mutlupolatcan@gmail.com>

WORKDIR /

COPY sr_service ./
RUN chmod +x /sr_service

ENTRYPOINT ["/entrypoint.sh"]
CMD ./sr_service && influxd