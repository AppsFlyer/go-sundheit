all: test

build: 
	##### building #####
	go build -v

test: build
	##### vetting #####
	go vet
	##### linting #####
	golint
	##### testing #####
	go test -v ./...

