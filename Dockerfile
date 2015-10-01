FROM quay.io/opsee/vinz:latest

ENV POSTGRES_CONN="postgres://postgres@postgresql/fieri_test?sslmode=disable"
ENV LOOKUPD_HOSTS=""
ENV NSQD_HOST=""
ENV AWS_ACCESS_KEY_ID=""
ENV AWS_SECRET_ACCESS_KEY=""
ENV AWS_DEFAULT_REGION=""
ENV AWS_INSTANCE_ID=""
ENV AWS_SESSION_TOKEN=""
ENV FIERI_CONCURRENCY=""
ENV BASTION_DISCOVERY_TOPIC=""
ENV FIERI_ONBOARDING_TOPIC=""
ENV FIERI_HTTP_ADDR=""
ENV YELLER_KEY=""

RUN apk add --update bash ca-certificates curl
RUN curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate
RUN curl -Lo /opt/bin/ec2-env https://s3-us-west-2.amazonaws.com/opsee-releases/go/ec2-env/ec2-env && \
    chmod 755 /opt/bin/ec2-env

COPY run.sh /
COPY target/linux/amd64/bin/* /
COPY migrations /migrations

EXPOSE 9092
CMD ["/fieri"]
