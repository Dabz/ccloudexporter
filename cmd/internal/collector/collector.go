package collector

//
// collector.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"regexp"
	"github.com/prometheus/client_golang/prometheus"
)

// CCloudCollectorMetric describes a single Metric from Confluent Cloud
type CCloudCollectorMetric struct {
	metric   MetricDescription
	desc     *prometheus.Desc
	duration *prometheus.GaugeVec
	labels   []string
	rule     Rule
	global   bool
}

// CCloudCollector is a custom prometheus collector to collect data from
// Confluent Cloud Metrics API
type CCloudCollector struct {
	metrics map[string]CCloudCollectorMetric
	rules   []Rule
}

// Describe collect all metrics for ccloudexporter
func (cc CCloudCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cc.metrics {
		ch <- desc.desc
		desc.duration.Describe(ch)
	}
}

var (
	httpClient http.Client
)

// Collect all metrics for Prometheus
// to avoid reaching the scrape_timeout, metrics are fetched in multiple goroutine
func (cc CCloudCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, rule := range cc.rules {
		for _, metric := range rule.Metrics {
			wg.Add(1)
			go cc.CollectMetricsForRule(&wg, ch, rule, cc.metrics[metric])
		}
	}

	wg.Wait()
}

// CollectMetricsForRule collects all metrics for a specific rule
func (cc CCloudCollector) CollectMetricsForRule(wg *sync.WaitGroup, ch chan<- prometheus.Metric, rule Rule, ccmetric CCloudCollectorMetric) {
	defer wg.Done()
	query := BuildQuery(ccmetric.metric, rule.Clusters, rule.GroupByLabels, rule.Topics, rule.excludeTopics)
	optimizedQuery, additionalLabels := OptimizeQuery(query)
	durationMetric, _ := ccmetric.duration.GetMetricWithLabelValues(strconv.Itoa(rule.id))
	timer := prometheus.NewTimer(prometheus.ObserverFunc(durationMetric.Set))
	response, err := SendQuery(optimizedQuery)
	timer.ObserveDuration()
	ch <- durationMetric
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cc.handleResponse(response, ccmetric, ch, rule, additionalLabels)
}

func (cc CCloudCollector) handleResponse(response QueryResponse, ccmetric CCloudCollectorMetric, ch chan<- prometheus.Metric, rule Rule, additionalLabels map[string]string) {
	desc := ccmetric.desc
	METRICSLOOP:
	for _, dataPoint := range response.Data {
		// Some data points might need to be ignored if it is the global query
		topic, topicPresent := dataPoint["metric.label.topic"].(string)
		cluster, clusterPresent := dataPoint["metric.label.cluster_id"].(string)

		if ! clusterPresent {
			cluster, clusterPresent = additionalLabels["metric.label.cluster_id"]
		}

		if ! topicPresent {
			topic, topicPresent = additionalLabels["metric.label.topic"]
		}

		if topicPresent && clusterPresent && rule.ShouldIgnoreResultForRule(topic, cluster, ccmetric.metric.Name) {
			continue METRICSLOOP
		}

		for _, currentRegex := range rule.excludeTopicsRegex {
			if matchesRegex(topic, currentRegex){
				continue METRICSLOOP
			}
		}

		value, ok := dataPoint["value"].(float64)
		if !ok {
			fmt.Println("Can not convert result to float:", dataPoint["value"])
			return
		}

		labels := []string{}
		for _, label := range ccmetric.labels {
			labelValue, labelValuePresent := dataPoint["metric.label."+label].(string)
			if ! labelValuePresent {
				labelValue, labelValuePresent = additionalLabels["metric.label."+label]
			}
			labels = append(labels, labelValue)
		}

		metric := prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			value,
			labels...,
		)

		if Context.NoTimestamp {
			ch <- metric
		} else {
			timestamp, err := time.Parse(time.RFC3339, fmt.Sprint(dataPoint["timestamp"]))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			metricWithTime := prometheus.NewMetricWithTimestamp(timestamp, metric)
			ch <- metricWithTime
		}
	}
}

// NewCCloudCollector creates a new instance of the collector
// During the creation, we invoke the descriptor endpoint to fetch all
// existing metrics and their labels
func NewCCloudCollector() CCloudCollector {
	collector := CCloudCollector{rules: Context.Rules, metrics: make(map[string]CCloudCollectorMetric)}
	descriptorResponse := SendDescriptorQuery()
	mapOfWhiteListedMetrics := Context.GetMapOfMetrics()

	for _, metr := range descriptorResponse.Data {
		_, metricPresent := mapOfWhiteListedMetrics[metr.Name]
		if !metricPresent {
			continue
		}
		delete(mapOfWhiteListedMetrics, metr.Name)
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

		requestDuration := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "ccloud_metrics_api_request_latency",
			Help:        "Metrics API request latency",
			ConstLabels: map[string]string{"metric": metr.Name},
		}, []string{"ruleNumber"})

		metric := CCloudCollectorMetric{
			metric:   metr,
			desc:     desc,
			duration: requestDuration,
			labels:   labels,
		}
		collector.metrics[metr.Name] = metric
	}

	httpClient = http.Client{
		Timeout: time.Second * time.Duration(Context.HTTPTimeout),
	}

	if len(mapOfWhiteListedMetrics) > 0 {
		fmt.Println("WARNING: The following metrics will not be gathered as they are not exposed by the Metrics API:")
		for key := range mapOfWhiteListedMetrics {
			fmt.Println("  -", key)
		}
	}

	return collector
}

func matchesRegex (candidate string, regex string) bool {
	matched, err := regexp.MatchString(regex,candidate)
	if err != nil {
		panic(err)
	}
	if matched {
		return true
	}
	return false
}
