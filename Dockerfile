# STEP 1 build executable binary
FROM golang:1.14 AS builder
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go install ./...

# STEP 2 build a small image
FROM scratch
ADD docker/telemetry-confluent-cloud-chain.pem /etc/ssl/certs/
COPY --from=builder /go/bin/ccloudexporter /

EXPOSE 2112
ENTRYPOINT [ "/ccloudexporter" ]
