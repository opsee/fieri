FROM alpine:3.3

RUN apk add --update bash ca-certificates curl
RUN mkdir -p /opt/bin && \
		curl -Lo /opt/bin/s3kms https://s3-us-west-2.amazonaws.com/opsee-releases/go/vinz-clortho/s3kms-linux-amd64 && \
    chmod 755 /opt/bin/s3kms && \
    curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate

ENV POSTGRES_CONN="postgres://postgres@postgresql/fieri_test?sslmode=disable"
ENV LOOKUPD_HOSTS=""
ENV NSQD_HOST=""
ENV FIERI_CONCURRENCY=""
ENV BASTION_DISCOVERY_TOPIC=""
ENV FIERI_ONBOARDING_TOPIC=""
ENV FIERI_HTTP_ADDR=""
ENV YELLER_KEY=""
ENV VAPE_ENDPOINT=""
ENV SLACK_ENDPOINT=""
ENV APPENV=""

COPY run.sh /
COPY target/linux/amd64/bin/* /
COPY migrations /migrations

EXPOSE 9092
CMD ["/fieri"]
