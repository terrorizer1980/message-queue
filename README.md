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

Current docker hash: `299d632e52925dca8a6fbdc56903384431b7ba33931d1f7a853503db187ecf99`

To build a new image, run `make package`. This will create a new image tagged as `mullvadvpn/message-queue`.
