package collector

import (
	"github.com/nerdynick/confluent-cloud-metrics-go-sdk/ccloudmetrics"
)

//
// context.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

// ExporterContext define the global context for ccloudexporter
// This global variables define all timeout, user configuration,
// and cluster information
type ExporterContext struct {
	HTTPTimeout int
	HTTPBaseURL string
	Delay       int
	Granularity string
	NoTimestamp bool
	Listener    string
	Rules       []Rule
}

// Rule defines one or multiple metrics that the exporter
// should collect for a specific set of topics or clusters
type Rule struct {
	Topics                           []string `mapstructure:"topics"`
	BlacklistedTopics                []string `mapstructure:"blacklisted"`
	Clusters                         []string `mapstructure:"clusters"`
	Metrics                          []string `mapstructure:"metrics"`
	GroupByLabels                    []string `mapstructure:"labels"`
	cachedIgnoreGlobalResultForTopic map[TopicClusterMetric]bool
	id                               int
	WhitelistedLabels                []ccloudmetrics.MetricLabel
}

// TopicClusterMetric represents a combination of a topic, a cluster and a metric
type TopicClusterMetric struct {
	Topic   string
	Cluster string
	Metric  string
}

// Version is the git short SHA1 hash provided at build time
var Version string = "homecooked"

// Context is the global variable defining the context for the expoter
var Context = ExporterContext{}

// DefaultGroupingLabels is the default value for groupBy.labels
var DefaultGroupingLabels = []string{
	"cluster_id",
	"topic",
	"type",
}

// DefaultMetrics is the default value for metrics
var DefaultMetrics = []string{
	"io.confluent.kafka.server/received_bytes",
	"io.confluent.kafka.server/sent_bytes",
	"io.confluent.kafka.server/received_records",
	"io.confluent.kafka.server/sent_records",
	"io.confluent.kafka.server/retained_bytes",
	"io.confluent.kafka.server/active_connection_count",
	"io.confluent.kafka.server/request_count",
	"io.confluent.kafka.server/partition_count",
}

// GetMapOfMetrics returns the whitelist of metrics in a map
// where the key is the metric and the value is true if it is comming from an override
func (context ExporterContext) GetMapOfMetrics() map[string]bool {
	mapOfWhiteListedMetrics := make(map[string]bool)

	for _, rule := range Context.Rules {
		for _, metric := range rule.Metrics {
			mapOfWhiteListedMetrics[metric] = true
		}
	}

	return mapOfWhiteListedMetrics
}

// GetMetrics return the list of all metrics exposed in any rule
func (context ExporterContext) GetMetrics() []string {
	metrics := make([]string, 0)
	for _, rule := range Context.Rules {
		for _, metric := range rule.Metrics {
			if !contains(metrics, metric) {
				metrics = append(metrics, metric)
			}
		}
	}

	return metrics
}

// ShouldIgnoreResultForRule returns true if the result for this topic need to be ignored for this rule.
// Some results might be ignored as they are defined in another rule, thus global and override result
// could conflict if we do not ignore the global result
func (rule Rule) ShouldIgnoreResultForRule(topic string, cluster string, metric string) bool {
	if rule.cachedIgnoreGlobalResultForTopic == nil {
		rule.cachedIgnoreGlobalResultForTopic = make(map[TopicClusterMetric]bool, 0)
	}

	result, present := rule.cachedIgnoreGlobalResultForTopic[TopicClusterMetric{topic, cluster, metric}]
	if present {
		return result
	}
	for _, irule := range Context.Rules {
		if irule.id == rule.id {
			continue
		}
		if contains(irule.Metrics, metric) && contains(irule.Clusters, cluster) {
			if len(rule.Topics) == 0 && len(irule.Topics) > 0 && contains(irule.Topics, topic) {
				rule.cachedIgnoreGlobalResultForTopic[TopicClusterMetric{topic, cluster, metric}] = true
				return true
			}
		}

	}
	rule.cachedIgnoreGlobalResultForTopic[TopicClusterMetric{topic, cluster, metric}] = false
	return false
}
