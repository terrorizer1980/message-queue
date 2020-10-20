.PHONY: fmt test vet install build

NAME = message-queue
REPO = quay.io/mullvad
VERSION = $(file < VERSION)

build:
	docker build -t $(REPO)/$(NAME):$(VERSION) -f Dockerfile .

clean:
	docker rmi -f $(REPO)/$(NAME):$(VERSION)

release: build
	docker push $(REPO)/$(NAME):$(VERSION)

fmt:
	go fmt ./...

test:
	go test -short ./...

vet:
	go vet ./...

install:
	go install ./...

all: test vet install

default: build
