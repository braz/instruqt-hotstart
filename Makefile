BINARY := instruqt-hotstart

.PHONY: build test vet fmt

build:
	go build -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .
