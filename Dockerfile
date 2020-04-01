# STEP 1 build executable binary
FROM golang:1.14 AS builder
COPY . src/github.com/Dabz/ccloudexporter
WORKDIR /go/src/github.com/Dabz/ccloudexporter/cmd/ccloudexporter
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/ccloudexporter ./ccloudexporter.go

# STEP 2 build a small image
FROM scratch
ADD docker/telemetry-confluent-cloud-chain.pem /etc/ssl/certs/
COPY --from=builder /go/bin/ccloudexporter /go/bin/ccloudexporter
CMD ["/go/bin/ccloudexporter"]
