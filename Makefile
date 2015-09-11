all: fmt build

build:
	gb build

clean:
	rm -fr target bin pkg

fmt:
	@gofmt -w ./

migrate:
	migrate -url $(POSTGRES_CONN) -path ./migrations up

.PHONY: migrate clean all
