.PHONY: build test run fmt vet lint snapshot clean

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

# snapshot builds the release artifacts locally without publishing, to
# sanity-check .goreleaser.yaml before pushing a tag.
snapshot:
	goreleaser release --snapshot --clean --skip=publish

clean:
	rm -rf bin/ dist/
