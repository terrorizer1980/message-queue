# Golang build step
FROM mullvadvpn/golang@sha256:4002fd88abf6894ba5ffcc87e06ee7fe6bed15c01028ce5bf662ff93df488cd3 as gobuilder
RUN apt-get update && apt-get install -y git
ADD . /message-queue
WORKDIR /message-queue
# Run golang tests
RUN make

# Copy go binary
FROM mullvadvpn/debian@sha256:72aa3e8b82527ed3247cf8872cddb22e82ab8821e9278ca98a055fe3317d892c
RUN mkdir /app
WORKDIR /app
COPY --from=gobuilder /go/bin/message-queue .

CMD ["./message-queue"]
