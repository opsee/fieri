APPENV ?= testenv
PROJECT := $(shell basename $$PWD)
REV ?= latest

all: build

clean:
	rm -fr target bin pkg

fmt:
	@gofmt -w ./

deps:
	docker-compose up -d
	docker run --link fieri_postgresql:postgres aanand/wait
	
migrate:
	migrate -url $(POSTGRES_CONN) -path ./migrations up

build: deps $(APPENV)
	docker run \
	  --env-file ./$(APPENV) \
		--link fieri_postgresql:postgresql \
		--link fieri_nsqd:nsqd \
		--link fieri_lookupd:lookupd \
		-e "TARGETS=linux/amd64" \
		-e PROJECT=github.com/opsee/$(PROJECT) \
		-v `pwd`:/gopath/src/github.com/opsee/$(PROJECT) \
		quay.io/opsee/build-go:16
	docker build -t quay.io/opsee/$(PROJECT):$(REV) .

run: deps $(APPENV)
	docker run \
	  --env-file ./$(APPENV) \
		--link fieri_postgresql:postgresql \
		--link fieri_nsqd:nsqd \
		--link fieri_lookupd:lookupd \
		-p 9092:9092 \
		--rm \
		quay.io/opsee/$(PROJECT):$(REV)

.PHONY: build run migrate clean all
