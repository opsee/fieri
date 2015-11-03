all: fmt build

build:
	gb build

clean:
	rm -fr target bin pkg

fmt:
	@gofmt -w ./

deps:
	docker pull sameersbn/postgresql:9.4-3
	@docker rm -f postgresql || true
	@docker run --name postgresql -d -e PSQL_TRUST_LOCALNET=true -e DB_USER=postgres -e DB_PASS= -e DB_NAME=fieri_test sameersbn/postgresql:9.4-3
	@echo "started postgresql"
	docker pull nsqio/nsq:latest
	@docker rm -f lookupd || true
	docker run --name lookupd -d nsqio/nsq /nsqlookupd
	@echo "started lookupd"
	@docker rm -f nsqd || true
	@docker run --name nsqd --link lookupd:lookupd -d nsqio/nsq /nsqd --broadcast-address=nsqd --lookupd-tcp-address=lookupd:4160
	@echo "started nsqd"

migrate:
	migrate -url $(POSTGRES_CONN) -path ./migrations up

docker: fmt
	docker run -e POSTGRES_CONN="postgres://postgres@postgresql/fieri_test?sslmode=disable" \
		--link postgresql:postgresql \
		--link nsqd:nsqd \
		--link lookupd:lookupd \
		-e NSQD_HOST="nsqd:4150" \
		-e LOOKUPD_HOSTS="http://lookupd:4161" \
		-e "TARGETS=linux/amd64" \
		-v `pwd`:/build quay.io/opsee/build-go \
		&& docker build -t quay.io/opsee/fieri .

run: docker
	docker run -e POSTGRES_CONN="postgres://postgres@postgresql/fieri_test?sslmode=disable" \
		--link postgresql:postgresql \
		--link nsqd:nsqd \
		--link lookupd:lookupd \
		-e VAPE_ENDPOINT=$(VAPE_ENDPOINT) \
		-e SLACK_ENDPOINT=$(SLACK_ENDPOINT) \
		-e NSQD_HOST="nsqd:4150" \
		-e LOOKUPD_HOSTS="http://lookupd:4161" \
	  -e BASTION_DISCOVERY_TOPIC="discovery" \
		-e FIERI_ONBOARDING_TOPIC="onboarding" \
		-e FIERI_HTTP_ADDR=":9092" \
		-p 9092:9092 \
		--rm \
		quay.io/opsee/fieri

.PHONY: docker run migrate clean all
