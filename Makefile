BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

all: build lint test

build: 
	##### building #####
	go build -v

$(GOLANGCI_LINT):
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.20.0

lint: $(GOLANGCI_LINT)
	##### linting #####
	golangci-lint run -E golint -E gosec -E gofmt

test:
	##### testing #####
	go test $(testflags) -v -race ./...
