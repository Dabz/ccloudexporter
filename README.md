# Prometheus exporter for Confluent Cloud Metrics API

A simple prometheus exporter that can be used to extract metrics from [Confluent Cloud Metric API](https://docs.confluent.io/current/cloud/metrics-api.html).
By default, the scrapper will be exposing the metrics on [port 2112](http://localhost:2112)
To use the scrapper, the following environment variables need to be specified:

* `CCLOUD_USER`: Your Confluent Cloud login
* `CCLOUD_PASSWORD`: Your Confluent Cloud password

`CCLOUD_USER` and `CCLOUD_PASSWORD` environment variables will be used to invoke the https://api.telemetry.confluent.cloud endpoint.

## Usage
```
./ccloudexporter <cluster_id> [kafka client configuration]
````

## Examples
```
export CCLOUD_USER=toto@confluent.io
export CCLOUD_PASSWORD=totopassword
./ccloudexporter lkc-abc123  
```

```
export CCLOUD_USER=toto@confluent.io
export CCLOUD_PASSWORD=totopassword
export CCLOUD_CLUSTER=lkc-abc123
docker-compose up -d
```

```
export CCLOUD_USER=toto@confluent.io
export CCLOUD_PASSWORD=totopassword
export CCLOUD_CLUSTER=lkc-abc123
docker-compose up -d
```

```
docker run -e CCLOUD_USER=$CCLOUD_USER -e CCLOUD_PASSWORD=$CCLOUD_PASSWORD dabz/ccloudexporter:latest ccloudexporter lkc-ldrq7
```

## How to build

```
go get github.com/Dabz/ccloudexporter/cmd/ccloudexporter
```
