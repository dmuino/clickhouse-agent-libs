SHELL=/bin/bash -o pipefail

all: clean test 

test:
	go test ./...

clean:
	rm -rf build
	mkdir -p build
	go get .

fmt:
	gofmt -s -w cmd/*.go

.PHONY: all test clean build fmt
