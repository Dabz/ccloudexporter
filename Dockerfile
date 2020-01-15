FROM golang:1.13

WORKDIR /go/src/app

RUN go get github.com/Dabz/ccloudexporter/cmd/ccloudexporter
RUN go install github.com/Dabz/ccloudexporter/cmd/ccloudexporter 

CMD ["ccloudexporter"]
