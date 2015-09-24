all: fmt build

build:
	gb build

clean:
	rm -fr target bin pkg

fmt:
	@gofmt -w ./

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
		-e NSQD_HOST="nsqd:4150" \
		-e LOOKUPD_HOSTS="http://lookupd:4161" \
	  -e BASTION_DISCOVERY_TOPIC="discovery" \
		-e FIERI_ONBOARDING_TOPIC="onboarding" \
		-e FIERI_HTTP_ADDR=":9092" \
		-p 9092:9092 \
		--rm \
		quay.io/opsee/fieri

.PHONY: docker run migrate clean all
