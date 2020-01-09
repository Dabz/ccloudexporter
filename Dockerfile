FROM golang:1.13

WORKDIR /go/src/app

RUN git clone -b v1.3.0 https://github.com/edenhill/librdkafka.git && cd librdkafka && ./configure --prefix /usr && make && make install
RUN go get github.com/Dabz/ccloudexporter/cmd/ccloudexporter
RUN go install github.com/Dabz/ccloudexporter/cmd/ccloudexporter 

CMD ["ccloudexporter"]
