package collector

//
// collector.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

type CCloudCollectorMetric struct {
	metric MetricDescription
	desc   *prometheus.Desc
}

type CCloudCollector struct {
	userName  string
	password  string
	cluster   string
	topicList map[string]*string
	metrics   []CCloudCollectorMetric
}

func (cc CCloudCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cc.metrics {
		ch <- desc.desc
	}
}

var (
	httpClient http.Client
)

func (cc CCloudCollector) Collect(ch chan<- prometheus.Metric) {
	from := time.Now().Add(time.Minute * -2)
	to := time.Now().Add(time.Minute * -1)
	for _, pair := range cc.metrics {
		desc := pair.desc
		query := BuildQuery(pair.metric, cc.cluster, from, to)
		response, err := SendQuery(query)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		set := make(map[string]struct{}, len(response.Data))
		for _, dataPoint := range response.Data {
			var metric prometheus.Metric
			if pair.metric.hasLabel("topic") {
				metric = prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					dataPoint.Value,
					dataPoint.Topic, cc.cluster,
				)

				_, ok := set[dataPoint.Topic]
				if ok {
					continue
				}
				set[dataPoint.Topic] = struct{}{}
				metricWithTime := prometheus.NewMetricWithTimestamp(dataPoint.Timestamp, metric)
				ch <- metricWithTime

			} else {
				metric = prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					dataPoint.Value,
					cc.cluster,
				)
				metricWithTime := prometheus.NewMetricWithTimestamp(dataPoint.Timestamp, metric)
				ch <- metricWithTime
				break
			}
		}
	}
}

func NewCCloudCollector() CCloudCollector {
	collector := CCloudCollector{cluster: Cluster}
	descriptorResponse := SendDescriptorQuery()
	for _, metr := range descriptorResponse.Data {
		var labels []string
		if metr.hasLabel("topic") {
			labels = []string{"topic", "cluster"}
		} else {
			labels = []string{"cluster"}
		}
		desc := prometheus.NewDesc(
			"ccloud_metric_"+GetNiceNameForMetric(metr),
			metr.Description,
			labels,
			nil,
		)

		metric := CCloudCollectorMetric{
			metric: metr,
			desc:   desc,
		}
		collector.metrics = append(collector.metrics, metric)
	}

	httpClient = http.Client{
		Timeout: time.Second * time.Duration(HttpTimeout),
	}
	return collector
}
