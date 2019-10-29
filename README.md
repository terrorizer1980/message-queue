# message-queue

A message queue that reads messages from redis pubsub, and publishes them to clients connected via websocket

## Building

Clone this repository, and run `make` to build.
This will produce a `message-queue` binary and put them in your `GOBIN`.

## Testing
To run the tests, run `make test`.
To run the integration tests as well, run `go test ./...`. Note that this requires redis to be running on `127.0.0.1:6379`

## Usage
All options can be either configured via command line flags, or via their respective environment variable, as denoted by `[ENVIRONMENT_VARIABLE]`.
To get a list of all the options, run `message-queue -h`.

## Packaging
In order to deploy message-queue, we use docker.

Current docker hash: ``

To build a new image, run `make package`. This will create a new image tagged as `mullvadvpn/message-queue`.
