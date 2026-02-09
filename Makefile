.PHONY: build test lint install clean vet fmt

BINARY := paymo
PKG    := ./...

build:
	go build -o $(BINARY) .

test:
	go test -race -cover $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -s -w .

lint: vet
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. See https://golangci-lint.run/"; exit 1; }
	golangci-lint run $(PKG)

install:
	go install .

clean:
	rm -f $(BINARY)
	go clean
