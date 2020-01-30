# message-queue

A message queue that reads messages from redis pubsub, and publishes them to clients connected via websocket

## Building

Clone this repository, and run `make` to build.
This will produce a `message-queue` binary and put them in your `GOBIN`.

## Testing
To run the tests, run `make test`.
To run the integration tests as well, run `go test ./...`. Note that this requires a local instance of redis and redis-sentinel.

## Usage
All options can be either configured via command line flags, or via their respective environment variable, as denoted by `[ENVIRONMENT_VARIABLE]`.
To get a list of all the options, run `message-queue -h`.

## Packaging
In order to deploy message-queue, we use docker.

Current docker hash: `8722d0ba33eec49c809af017f70c56cbfe6593b4b0b05c3e13012776f610b590`

To build a new image, run `make package`. This will create a new image tagged as `mullvadvpn/message-queue`.
