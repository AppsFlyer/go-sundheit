BIN_DIR := ./bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

all: build lint test

build: 
	##### building #####
	go build -v

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.55.2

lint: $(GOLANGCI_LINT)
	##### linting #####
	$(GOLANGCI_LINT) run -E revive -E gosec -E gofmt

test: build
	##### testing #####
	go test $(testflags) -v -race ./...
