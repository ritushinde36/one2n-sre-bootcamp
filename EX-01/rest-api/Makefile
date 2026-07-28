.PHONY: build run test

build:
	go build -o bin/rest-api .

run:
	go run main.go

# test requires Docker to be running (spins up a real MySQL container).
test:
	go test ./...
