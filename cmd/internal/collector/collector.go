package collector

//
// collector.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nerdynick/confluent-cloud-metrics-go-sdk/ccloudmetrics"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// CCloudCollectorMetric describes a single Metric from Confluent Cloud
type CCloudCollectorMetric struct {
	metric   ccloudmetrics.Metric
	desc     *prometheus.Desc
	duration *prometheus.GaugeVec
	labels   []string
	rule     Rule
	global   bool
}

// CCloudCollector is a custom prometheu collector to collect data from
// Confluent Cloud Metrics API
type CCloudCollector struct {
	metrics map[string]CCloudCollectorMetric
	rules   []Rule
	client  ccloudmetrics.MetricsClient
}

// Describe collect all metrics for ccloudexporter
func (cc CCloudCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cc.metrics {
		ch <- desc.desc
		desc.duration.Describe(ch)
	}
}

// Collect all metrics for Prometheus
// to avoid reaching the scrape_timeout, metrics are fetched in multiple goroutine
func (cc CCloudCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	now := time.Now()
	granualrity := ccloudmetrics.Granularity(Context.Granularity)
	for _, rule := range cc.rules {
		for _, cluster := range rule.Clusters {
			for _, metric := range rule.Metrics {
				ccmetric := cc.metrics[metric]
				wg.Add(1)
				go cc.CollectMetricsForRule(&wg, ch, now, granualrity, cluster, ccmetric, rule)
			}
		}
	}

	wg.Wait()
}

// CollectMetricsForRule collects all metrics for a specific rule
func (cc CCloudCollector) CollectMetricsForRule(wg *sync.WaitGroup, ch chan<- prometheus.Metric, now time.Time, granualrity ccloudmetrics.Granularity, cluster string, metric CCloudCollectorMetric, rule Rule) {
	defer wg.Done()
	var response []ccloudmetrics.QueryData
	var err error
	if len(rule.Topics) > 0 && metric.metric.HasLabel(ccloudmetrics.MetricLabelTopic) {
		hasPartition := false
		for _, label := range rule.WhitelistedLabels {
			if ccloudmetrics.MetricLabelPartition == label {
				hasPartition = true
			}
		}
		response, err = cc.client.QueryMetricAndTopics(cluster, metric.metric, rule.Topics, granualrity, granualrity.GetStartTimeFromGranularity(now), now, hasPartition, blacklistedTopics)
	} else {
		response, err = cc.client.QueryMetricWithAggs(cluster, metric.metric, granualrity, granualrity.GetStartTimeFromGranularity(now), now, rule.WhitelistedLabels)
	}

	//CCloudMetrics SDK will allow for partial results even in the event of errors
	if response != nil {
		cc.handleResponse(response, metric, rule, ch)
	}

	if err != nil {
		log.Error(fmt.Sprintf("Error fetching results for rule. Error: %s", err.Error()))
		return
	}
}

func (cc CCloudCollector) handleResponse(response []ccloudmetrics.QueryData, ccmetric CCloudCollectorMetric, rule Rule, ch chan<- prometheus.Metric) {
	for _, dataPoint := range response {

		if dataPoint.HasTopic() && dataPoint.HasCluster() && rule.ShouldIgnoreResultForRule(dataPoint.Topic, dataPoint.Cluster, ccmetric.metric.Name) {
			continue
		}

		labels := []string{}
		for _, label := range ccmetric.metric.Labels {
			switch label.MetricLabel() {
			case ccloudmetrics.MetricLabelCluster:
				labels = append(labels, dataPoint.Cluster)
			case ccloudmetrics.MetricLabelTopic:
				labels = append(labels, dataPoint.Topic)
			case ccloudmetrics.MetricLabelPartition:
				labels = append(labels, dataPoint.Partition)
			case ccloudmetrics.MetricLabelType:
				labels = append(labels, dataPoint.Type)
			}
		}

		metric := prometheus.MustNewConstMetric(
			ccmetric.desc,
			prometheus.GaugeValue,
			dataPoint.Value,
			labels...,
		)

		if Context.NoTimestamp {
			ch <- metric
		} else {
			timestamp, err := dataPoint.Time()
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
// During the creation, we invoke the descriptor endpoint to fetcha all
// existing metrics and their labels
func NewCCloudCollector() CCloudCollector {
	user, present := os.LookupEnv("CCLOUD_USER")
	if !present || user == "" {
		fmt.Print("CCLOUD_USER environment variable has not been specified")
		os.Exit(1)
	}
	password, present := os.LookupEnv("CCLOUD_PASSWORD")
	if !present || password == "" {
		fmt.Print("CCLOUD_PASSWORD environment variable has not been specified")
		os.Exit(1)
	}

	apiContext := ccloudmetrics.NewAPIContext(user, password)
	apiContext.BaseURL = Context.HTTPBaseURL

	headers := make(map[string]string)
	headers["Correlation-Context"] = fmt.Sprintf("service.name=ccloudexporter,service.version=%s", Version)

	httpContext := ccloudmetrics.NewHTTPContext()
	httpContext.UserAgent = fmt.Sprintf("ccloudexporter/%s", Version)
	httpContext.HTTPHeaders = headers

	collector := CCloudCollector{
		rules:   Context.Rules,
		metrics: make(map[string]CCloudCollectorMetric),
		client:  ccloudmetrics.NewClientFromContext(apiContext, httpContext),
	}
	descriptorResponse, err := collector.client.GetAvailableMetrics()
	if err != nil {
		os.Exit(1)
	}

	mapOfWhiteListedMetrics := Context.GetMapOfMetrics()

	for _, metr := range descriptorResponse {
		_, metricPresent := mapOfWhiteListedMetrics[metr.Name]
		if !metricPresent {
			continue
		}
		delete(mapOfWhiteListedMetrics, metr.Name)

		var labels []string
		for _, metrLabel := range metr.Labels {
			labels = append(labels, metrLabel.Name)
		}

		desc := prometheus.NewDesc(
			"ccloud_metric_"+metr.ShortName(),
			metr.Desc,
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

	if len(mapOfWhiteListedMetrics) > 0 {
		fmt.Println("WARNING: The following metrics will not be gathered as they are not exposed by the Metrics API:")
		for key := range mapOfWhiteListedMetrics {
			fmt.Println("  -", key)
		}
	}

	return collector
}
