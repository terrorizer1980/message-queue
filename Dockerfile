# Golang build step / Debian Buster 20-09 / Golang 1.15.2
FROM quay.io/mullvad/golang@sha256:1d7a70635fb0b7703984db2d0655f710ff9c13ff2c8ec4db023b03c17b1d3034 as gobuilder
RUN apt-get update && apt-get install -y git
ADD . /message-queue
WORKDIR /message-queue
# Run golang tests
RUN make

# Copy go binary
FROM quay.io/mullvad/debian@sha256:f49f6dc4d85279a57bedfb9a725bd01ec1ec3ec1f0b1a5d5effe91361f3d073a
RUN mkdir /app
WORKDIR /app
COPY --from=gobuilder /go/bin/message-queue .

CMD ["./message-queue"]
