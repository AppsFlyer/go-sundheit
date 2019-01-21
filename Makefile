all: test

deps:
	##### ensuring deps #####
	dep ensure -v

build: deps
	##### building #####
	go build -v

test: build
	##### vetting #####
	go vet
	##### linting #####
	golint
	##### testing #####
	go test -v ./...

