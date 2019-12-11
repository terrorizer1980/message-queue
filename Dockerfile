# Golang build step
FROM mullvadvpn/golang@sha256:0e2105f55f7137671e1b5e71108f490cf6a7f90011f7b06ba4ec0908fea1b1df as gobuilder
RUN apt-get update && apt-get install -y git
ADD . /message-queue
WORKDIR /message-queue
# Run golang tests
RUN make

# Copy go binary
FROM mullvadvpn/debian@sha256:c6728a03350ba492e786c02a4bdaf958fe0ba6062fe4b5a1c76609e4c3a31836
RUN mkdir /app
WORKDIR /app
COPY --from=gobuilder /go/bin/message-queue .

CMD ["./message-queue"]
