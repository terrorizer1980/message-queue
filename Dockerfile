# Golang build step
FROM mullvadvpn/golang@sha256:0e2105f55f7137671e1b5e71108f490cf6a7f90011f7b06ba4ec0908fea1b1df as gobuilder
RUN apt-get update && apt-get install -y git
ADD . /message-queue
WORKDIR /message-queue
# Run golang tests
RUN make

# Copy go binary
FROM quay.io/mullvad/debian@sha256:c199b5d707db6f472d25eb8d78b0ff69f14e819c7affea70f9cad6a33e2b67de
RUN mkdir /app
WORKDIR /app
COPY --from=gobuilder /go/bin/message-queue .

CMD ["./message-queue"]
