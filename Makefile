.PHONY: build test run fmt vet lint clean

build:
	go build -o bin/paywall-sandbox ./cmd/paywall-sandbox

test:
	go test -race ./...

run: build
	./bin/paywall-sandbox serve

fmt:
	gofmt -l -w .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
