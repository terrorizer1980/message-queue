.PHONY: fmt test vet install package
all: test vet install

fmt:
	go fmt ./...

test:
	go test -short ./...

vet:
	go vet ./...

install:
	go install ./...

# Return contents of VERSION and interpolate
version := $(shell cat VERSION)
package:
	docker build -t quay.io/mullvad/message-queue:$(version) .
