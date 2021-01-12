package collector

//
// collector.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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

// CCloudCollector is a custom prometheu collector to collect data from
// Confluent Cloud Metrics API
type CCloudCollector struct {
	metrics            map[string]CCloudCollectorMetric
	rules              []Rule
	kafkaCollector     *KafkaCCloudCollector
	connectorCollector *ConnectorCCloudCollector
	ksqlCollector      *KsqlCCloudCollector
}

var (
	httpClient http.Client
)

// Describe collect all metrics for ccloudexporter
func (cc CCloudCollector) Describe(ch chan<- *prometheus.Desc) {
	cc.kafkaCollector.Describe(ch)
	cc.connectorCollector.Describe(ch)
	cc.ksqlCollector.Describe(ch)
}

// Collect all metrics for Prometheus
// to avoid reaching the scrape_timeout, metrics are fetched in multiple goroutine
func (cc CCloudCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	cc.kafkaCollector.Collect(ch, &wg)
	cc.connectorCollector.Collect(ch, &wg)
	cc.ksqlCollector.Collect(ch, &wg)
	wg.Wait()
}

// NewCCloudCollector creates a new instance of the collector
// During the creation, we invoke the descriptor endpoint to fetcha all
// existing metrics and their labels
func NewCCloudCollector() CCloudCollector {

	log.Traceln("Creating http client")
	httpClient = http.Client{
		Timeout: time.Second * time.Duration(Context.HTTPTimeout),
	}

	var (
		connectorResource ResourceDescription
		kafkaResource     ResourceDescription
		ksqlResource      ResourceDescription
	)
	resourceDescription := SendResourceDescriptorQuery()
	for _, resource := range resourceDescription.Data {
		if resource.Type == "connector" {
			connectorResource = resource
		} else if resource.Type == "kafka" {
			kafkaResource = resource
		} else if resource.Type == "ksql" {
			ksqlResource = resource
		}
	}

	if connectorResource.Type == "" {
		log.WithField("descriptorResponse", resourceDescription).Fatalln("No connector resource available")
	}

	if kafkaResource.Type == "" {
		log.WithField("descriptorResponse", resourceDescription).Fatalln("No kafka resource available")
	}

	if ksqlResource.Type == "" {
		log.WithField("descriptorResponse", resourceDescription).Fatalln("No ksqlDB resource available")
	}

	collector := CCloudCollector{rules: Context.Rules, metrics: make(map[string]CCloudCollectorMetric)}
	kafkaCollector := NewKafkaCCloudCollector(collector, kafkaResource)
	connectorCollector := NewConnectorCCloudCollector(collector, connectorResource)
	ksqlCollector := NewKsqlCCloudCollector(collector, ksqlResource)

	collector.kafkaCollector = &kafkaCollector
	collector.connectorCollector = &connectorCollector
	collector.ksqlCollector = &ksqlCollector

	return collector
}
