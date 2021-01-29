package collector

//
// collector.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// ConnectorCCloudCollector is a custom prometheu collector to collect data from
// Confluent Cloud Metrics API. It fetches Connector resources types metrics
type ConnectorCCloudCollector struct {
	metrics  map[string]CCloudCollectorMetric
	rules    []Rule
	ccloud   CCloudCollector
	resource ResourceDescription
}

// Describe collect all metrics for ccloudexporter
func (cc ConnectorCCloudCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cc.metrics {
		ch <- desc.desc
		desc.duration.Describe(ch)
	}
}

// Collect all metrics for Prometheus
// to avoid reaching the scrape_timeout, metrics are fetched in multiple goroutine
func (cc ConnectorCCloudCollector) Collect(ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	for _, rule := range cc.rules {
		for _, metric := range rule.Metrics {
			_, present := cc.metrics[metric]
			if !present {
				continue
			}

			if len(rule.Connectors) <= 0 {
				log.WithFields(log.Fields{"rule": rule}).Errorln("connector rule has no cluster specified")
				continue
			}

			wg.Add(1)
			go cc.CollectMetricsForRule(wg, ch, rule, cc.metrics[metric])
		}
	}
}

// CollectMetricsForRule collects all metrics for a specific rule
func (cc ConnectorCCloudCollector) CollectMetricsForRule(wg *sync.WaitGroup, ch chan<- prometheus.Metric, rule Rule, ccmetric CCloudCollectorMetric) {
	defer wg.Done()
	query := BuildConnectorsQuery(ccmetric.metric, rule.Connectors, cc.resource)
	log.WithFields(log.Fields{"query": query}).Traceln("The following query has been created")
	optimizedQuery, additionalLabels := OptimizeQuery(query)
	log.WithFields(log.Fields{"optimizedQuery": optimizedQuery, "additionalLabels": additionalLabels}).Traceln("Query has been optimized")
	durationMetric, _ := ccmetric.duration.GetMetricWithLabelValues(strconv.Itoa(rule.id))
	timer := prometheus.NewTimer(prometheus.ObserverFunc(durationMetric.Set))
	response, err := SendQuery(optimizedQuery)
	timer.ObserveDuration()
	ch <- durationMetric
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"optimizedQuery": optimizedQuery, "response": response}).Errorln("Query did not succeed")
		return
	}
	log.WithFields(log.Fields{"response": response}).Traceln("Response has been received")
	cc.handleResponse(response, ccmetric, ch, rule, additionalLabels)
}

func (cc ConnectorCCloudCollector) handleResponse(response QueryResponse, ccmetric CCloudCollectorMetric, ch chan<- prometheus.Metric, rule Rule, additionalLabels map[string]string) {
	desc := ccmetric.desc
	for _, dataPoint := range response.Data {
		value, ok := dataPoint["value"].(float64)
		if !ok {
			log.WithField("datapoint", dataPoint["value"]).Errorln("Can not convert result to float")
			return
		}

		labels := []string{}
		for _, label := range ccmetric.labels {
			name := cc.resource.datapointFieldNameForLabel(label)
			labelValue, labelValuePresent := dataPoint[name].(string)
			if !labelValuePresent {
				labelValue, labelValuePresent = additionalLabels[name]
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
				log.WithError(err).Errorln("Can not parse timestamp, ignoring the response")
				return
			}
			metricWithTime := prometheus.NewMetricWithTimestamp(timestamp, metric)
			ch <- metricWithTime
		}
	}
}

// NewConnectorCCloudCollector create a new Confluent Cloud Connector collector
func NewConnectorCCloudCollector(ccloudcollecter CCloudCollector, resource ResourceDescription) ConnectorCCloudCollector {
	collector := ConnectorCCloudCollector{
		rules:    Context.GetConnectorRules(),
		metrics:  make(map[string]CCloudCollectorMetric),
		ccloud:   ccloudcollecter,
		resource: resource,
	}
	descriptorResponse := SendDescriptorQuery(resource.Type)
	log.WithField("descriptor response", descriptorResponse).Traceln("The following response for the descriptor endpoint has been received")
	mapOfWhiteListedMetrics := Context.GetMapOfMetrics("io.confluent.kafka.connect")

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

		for _, rsrcLabel := range resource.Labels {
			labels = append(labels, GetPrometheusNameForLabel(rsrcLabel.Key))
		}

		desc := prometheus.NewDesc(
			"ccloud_metric_connector_"+GetNiceNameForMetric(metr),
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

	if len(mapOfWhiteListedMetrics) > 0 {
		log.WithField("Ignored metrics", mapOfWhiteListedMetrics).Warnln("The following metrics will not be gathered as they are not exposed by the Metrics API")
	}

	return collector
}
