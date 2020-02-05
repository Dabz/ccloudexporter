package scraper

//
// collector.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type CCloudCollector struct {
	userName  string
	password  string
	cluster   string
	topicList map[string]*string
	metrics   map[MetricDescription]*prometheus.Desc
}

func (cc CCloudCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cc.metrics {
		ch <- desc
	}
}

func (cc CCloudCollector) Collect(ch chan<- prometheus.Metric) {
	from := time.Now().Add(time.Minute * -3)
	to := time.Now().Add(time.Minute * -1)
	for metric, desc := range cc.metrics {
		query := BuildQuery(metric.Name, cc.cluster, from, to)
		response, err := SendQuery(query)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		for _, dataPoint := range response.Data {
			metric := prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				dataPoint.Value,
				cc.cluster, dataPoint.Topic,
			)
			metric = prometheus.NewMetricWithTimestamp(dataPoint.Timestamp, metric)
			ch <- metric
		}
	}
}

func NewCCloudCollector(cluster string) CCloudCollector {
	collector := CCloudCollector{cluster: cluster}
	descriptorResponse := SendDescriptorQuery()
	collector.metrics = make(map[MetricDescription]*prometheus.Desc)
	for _, metric := range descriptorResponse.Data {
		desc := prometheus.NewDesc(
			"ccloud_metric_"+GetNiceNameForMetric(metric),
			metric.Description,
			[]string{"topic", "cluster"},
			nil,
		)
		collector.metrics[metric] = desc
	}
	return collector
}
