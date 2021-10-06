package collector

//
// context.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "strings"

// ExporterContext define the global context for ccloudexporter
// This global variables define all timeout, user configuration,
// and cluster information
type ExporterContext struct {
	HTTPTimeout  int
	HTTPBaseURL  string
	Delay        int
	CachedSecond int
	Granularity  string
	NoTimestamp  bool
	Listener     string
	Rules        []Rule
}

// Rule defines one or multiple metrics that the exporter
// should collect for a specific set of topics or clusters
type Rule struct {
	Topics                           []string `mapstructure:"topics"`
	Clusters                         []string `mapstructure:"clusters"`
	Connectors                       []string `mapstructure:"connectors"`
	Ksql                             []string `mapstructure:"ksqls"`
	SchemaRegistries                 []string `mapstructure:"schemaregistries"`
	Metrics                          []string `mapstructure:"metrics"`
	GroupByLabels                    []string `mapstructure:"labels"`
	cachedIgnoreGlobalResultForTopic map[TopicClusterMetric]bool
	id                               int
}

// TopicClusterMetric represents a combination of a topic, a cluster and a metric
type TopicClusterMetric struct {
	Topic   string
	Cluster string
	Metric  string
}

// Version is the git short SHA1 hash provided at build time
var Version = "homecooked"

// Context is the global variable defining the context for the expoter
var Context = ExporterContext{}

// DefaultGroupingLabels is the default value for groupBy.labels
var DefaultGroupingLabels = []string{
	"kafka.id",
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
	"io.confluent.kafka.server/successful_authentication_count",
	"io.confluent.kafka.connect/sent_bytes",
	"io.confluent.kafka.connect/received_bytes",
	"io.confluent.kafka.connect/received_records",
	"io.confluent.kafka.connect/sent_records",
	"io.confluent.kafka.connect/dead_letter_queue_records",
	"io.confluent.kafka.ksql/streaming_unit_count",
	"io.confluent.kafka.schema_registry/schema_count",
}

// GetMapOfMetrics returns the whitelist of metrics in a map
// where the key is the metric and the value is true if it is comming from an override
func (context ExporterContext) GetMapOfMetrics(prefix string) map[string]bool {
	mapOfWhiteListedMetrics := make(map[string]bool)

	for _, rule := range Context.Rules {
		for _, metric := range rule.Metrics {
			if strings.HasPrefix(metric, prefix) {
				mapOfWhiteListedMetrics[metric] = true
			}
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

// GetKafkaRules return all rules associated to a Kafka cluster
func (context ExporterContext) GetKafkaRules() []Rule {
	kafkaRules := make([]Rule, 0)
	for _, irule := range Context.Rules {
		if len(irule.Clusters) > 0 {
			kafkaRules = append(kafkaRules, irule)
		}
	}

	return kafkaRules
}

// GetConnectorRules return all rules associated to at least one connector
func (context ExporterContext) GetConnectorRules() []Rule {
	connectorRule := make([]Rule, 0)
	for _, irule := range Context.Rules {
		if len(irule.Connectors) > 0 {
			connectorRule = append(connectorRule, irule)
		}
	}

	return connectorRule
}

// GetKsqlRules return all rules associated to at least one ksql application
func (context ExporterContext) GetKsqlRules() []Rule {
	ksqlRules := make([]Rule, 0)
	for _, irule := range Context.Rules {
		if len(irule.Ksql) > 0 {
			ksqlRules = append(ksqlRules, irule)
		}
	}

	return ksqlRules
}

// GetSchemaRegistryRules return all rules associated to at least one Schema Registry instance
func (context ExporterContext) GetSchemaRegistryRules() []Rule {
	schemaRegistryRules := make([]Rule, 0)
	for _, irule := range Context.Rules {
		if len(irule.SchemaRegistries) > 0 {
			schemaRegistryRules = append(schemaRegistryRules, irule)
		}
	}

	return schemaRegistryRules
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
