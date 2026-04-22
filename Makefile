.PHONY: build test test-verbose coverage coverage-html mutation mutation-docker clean

BINARY := njord
PKG := ./...

build:
	go build -o $(BINARY) ./cmd/njord/

test:
	go test $(PKG)

test-verbose:
	go test -v $(PKG)

coverage:
	go test -cover $(PKG)

coverage-html:
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório: coverage.html"

mutation:
	gremlins unleash $(PKG)

mutation-docker:
	gremlins unleash ./internal/docker/

mutation-app:
	gremlins unleash ./internal/app/

clean:
	rm -f $(BINARY) coverage.out coverage.html
