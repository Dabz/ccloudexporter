# Prometheus exporter for Confluent Cloud Metrics API

A simple prometheus exporter that can be used to extract metrics from [Confluent Cloud Metric API](https://docs.confluent.io/current/cloud/metrics-api.html).
By default, the scrapper will be exposing the metrics on [port 2112](http://localhost:2112)
In order to use the scrapper, the following environment variable need to be specified:

* `CCLOUD_USER`: Your Confluent Cloud login
* `CCLOUD_PASSWORD`: Your Confluent Cloud password

`CCLOUD_USER` and `CCLOUD_PASSWORD` environment variable will be used to invoke the https://api.telemetry.confluent.cloud endpoint

## Usage
```
./ccloudexporter <cluster_id> [kafka client configuration]
````


## Examples
```
./ccloudexporter lkc-abc123  
```

## How to build

TODO
