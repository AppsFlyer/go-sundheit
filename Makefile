BIN_DIR := ./bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

all: build lint test

build: 
	##### building #####
	go build -v

GOLANGCI_LINT_VERSION := v2.12.2

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh | sh -s $(GOLANGCI_LINT_VERSION)

lint: $(GOLANGCI_LINT)
	##### linting #####
	$(GOLANGCI_LINT) run

test: build
	##### testing #####
	go test $(testflags) -v -race ./...
