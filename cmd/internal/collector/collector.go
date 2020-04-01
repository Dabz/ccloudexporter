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
	"strings"
	"sync"
	"time"
)

type CCloudCollectorMetric struct {
	metric   MetricDescription
	desc     *prometheus.Desc
	duration prometheus.Gauge
	labels   []string
}

type CCloudCollector struct {
	userName  string
	password  string
	cluster   string
	topicList map[string]*string
	metrics   []CCloudCollectorMetric
}

// Describing all metrics
func (cc CCloudCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cc.metrics {
		ch <- desc.desc
		ch <- desc.duration.Desc()
	}
}

var (
	httpClient http.Client
)

// Collect all metrics for Prometheus
// to avoid reaching the scrape_timeout, metrics are fetched in multiple goroutine
func (cc CCloudCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	from := time.Now().Add(time.Minute * -2) // the last minute might contains data that is not yet finalized
	to := time.Now().Add(time.Minute * -1)   // in order to avoid fetching intermediate data, we always fetch current time - 1
	for _, ccmetric := range cc.metrics {
		wg.Add(1)
		go cc.CollectMetric(&wg, ch, ccmetric, from, to)
	}
	wg.Wait()
}

// Collecting a specific metric for a time window
func (cc CCloudCollector) CollectMetric(wg *sync.WaitGroup, ch chan<- prometheus.Metric, ccmetric CCloudCollectorMetric, from time.Time, to time.Time) {
	defer wg.Done()

	desc := ccmetric.desc
	query := BuildQuery(ccmetric.metric, cc.cluster, from, to)
	timer := prometheus.NewTimer(prometheus.ObserverFunc(ccmetric.duration.Set))
	response, err := SendQuery(query)
	timer.ObserveDuration()
	ch <- ccmetric.duration
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, dataPoint := range response.Data {
		labels := []string{}
		for _, label := range ccmetric.labels {
			labels = append(labels, getField(&dataPoint, strings.Title(strings.ToLower(label))))
		}

		metric := prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			dataPoint.Value,
			labels...,
		)

		metricWithTime := prometheus.NewMetricWithTimestamp(dataPoint.Timestamp, metric)
		ch <- metricWithTime
	}
}

// Create a new instance of the collector
// During the creation, we invoke the descriptor endpoint to fetcha all
// existing metrics and their labels
func NewCCloudCollector() CCloudCollector {
	collector := CCloudCollector{cluster: Cluster}
	descriptorResponse := SendDescriptorQuery()
	for _, metr := range descriptorResponse.Data {
		var labels []string
		for _, metrLabel := range metr.Labels {
			labels = append(labels, metrLabel.Key)
		}

		desc := prometheus.NewDesc(
			"ccloud_metric_"+GetNiceNameForMetric(metr),
			metr.Description,
			labels,
			nil,
		)

		requestDuration := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "ccloud_metrics_api_request_latency",
			Help:        "Metrics API request latency",
			ConstLabels: map[string]string{"metric": metr.Name},
		})

		metric := CCloudCollectorMetric{
			metric:   metr,
			desc:     desc,
			duration: requestDuration,
			labels:   labels,
		}
		collector.metrics = append(collector.metrics, metric)
	}

	httpClient = http.Client{
		Timeout: time.Second * time.Duration(HttpTimeout),
	}

	return collector
}
