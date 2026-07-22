.PHONY: all build fmt vet test race clean

all: fmt vet test build

build:
	go build -o bin/supervisor ./cmd/supervisor

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

clean:
	rm -rf bin
