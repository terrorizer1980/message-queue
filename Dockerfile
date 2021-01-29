# Golang build step / Debian Buster 21.01 / Golang 1.15.7
FROM quay.io/mullvad/golang@sha256:bed1bb603f38e0349cd06f2281b7e59a008192d8e99df4ea8792dc40950e8ab9 as gobuilder
RUN apt-get update && apt-get install -y git
ADD . /message-queue
WORKDIR /message-queue
# Run golang tests
RUN make all

# Copy go binary
FROM quay.io/mullvad/debian@sha256:3cd089ccc7d89c2488795c8828a698baf77000791df2cd335982019ec838a849
RUN mkdir /app
WORKDIR /app
COPY --from=gobuilder /go/bin/message-queue .

CMD ["./message-queue"]
