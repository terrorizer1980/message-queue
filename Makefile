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

package:
	docker build -t mullvadvpn/message-queue .
