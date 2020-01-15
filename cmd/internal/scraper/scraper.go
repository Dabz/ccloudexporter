//
// main.go
// Copyright (C) 2020 gaspar_d d.gasparina@gmail.com
//
// Distributed under terms of the MIT license.
//

package scraper

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

var (
	topicList    = make(map[string]*string)
	gaugeMetrics = make(map[string]*prometheus.GaugeVec)
)

func handleResponseMetric(cluster string, response QueryResponse, vec *prometheus.GaugeVec) {
	exploredTopic := make(map[string]int)
	for _, data := range response.Data {
		_, exist := topicList[data.Topic]
		if !exist {
			topicList[data.Topic] = nil
		}
		exploredTopic[data.Topic] = 0
		vec.With(prometheus.Labels{"topic": data.Topic, "cluster": cluster}).Set(data.Value)
	}

	for topic := range topicList {
		_, exist := exploredTopic[topic]
		if !exist {
			vec.With(prometheus.Labels{"topic": topic, "cluster": cluster}).Set(0.0)
		}
	}
}

// Start a routine what will scrape the Confluent Cloud
// Metrics every 15 seconds
func FetchMetricsFromEndpointRoutine(cluster string) {
	// Before starting the routing, invoking the descriptors endpoints
	// to fetch the list of metric to scap
	descriptorResponse := SendDescriptorQuery()
	for _, metric := range descriptorResponse.Data {
		if metric.Type == "GAUGE_INT64" {
			gaugeMetrics[metric.Name] = promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "ccloud_metric_" + GetNiceNameForMetric(metric),
				Help: metric.Description,
			}, []string{"topic", "cluster"})
		}
	}

	// Main routine scraping the Confluent Cloud API endpoint
	go func() {
		for {
			from := time.Now().Add(time.Minute * - 3)
			to := time.Now().Add(time.Minute * - 1)
			for metric, gauge := range gaugeMetrics {
				query := BuildQuery(metric, cluster, from, to)
				response := SendQuery(query)
				handleResponseMetric(cluster, response, gauge)
			}

			time.Sleep(time.Second * 15)
		}
	}()
}
