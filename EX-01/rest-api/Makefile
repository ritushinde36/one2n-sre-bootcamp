.PHONY: build run test test-integration

build:
	go build -o bin/rest-api .

run:
	go run main.go

test:
	go test ./...

test-integration:
	go test -tags=integration ./integration_test/...
