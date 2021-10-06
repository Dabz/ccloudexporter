package collector

//
// query.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Query to Confluent Cloud API metric endpoint
// This is the JSON structure for the endpoint
// https://api.telemetry.confluent.cloud/v1/metrics/cloud/descriptors
type Query struct {
	Aggreations []Aggregation `json:"aggregations"`
	Filter      FilterHeader  `json:"filter"`
	Granularity string        `json:"granularity"`
	GroupBy     []string      `json:"group_by"`
	Intervals   []string      `json:"intervals"`
	Limit       int           `json:"limit"`
}

// Aggregation for a Confluent Cloud API metric
type Aggregation struct {
	Agg    string `json:"agg,omitempty"`
	Metric string `json:"metric"`
}

// FilterHeader to use for a query
type FilterHeader struct {
	Op      string   `json:"op"`
	Filters []Filter `json:"filters"`
}

// Filter structure
type Filter struct {
	Field   string   `json:"field,omitempty"`
	Op      string   `json:"op"`
	Value   string   `json:"value,omitempty"`
	Filters []Filter `json:"filters,omitempty"`
}

// QueryResponse from the cloud endpoint
type QueryResponse struct {
	Data []map[string]interface{} `json:"data"`
}

// Actual data point from the query response
// Only there for reference
// type Data struct {
// 	Topic      string    `json:"metric.label.topic"`
// 	Cluster_id string    `json:"metric.label.cluster_id"`
// 	Type       string    `json:"metric.label.type"`
// 	Partition  string    `json:"metric.label.partition"`
// 	Timestamp  time.Time `json:"timestamp"`
// 	Value      float64   `json:"value"`
// }

var (
	queryURI = "/v2/metrics/cloud/query"
)

// BuildQuery creates a new Query for a metric for a specific cluster and time interval
// This function will return the main global query, override queries will not be generated
func BuildQuery(metric MetricDescription, clusters []string, groupByLabels []string, topicFiltering []string, resource ResourceDescription) Query {
	timeFrom := time.Now().Add(time.Duration(-Context.Delay) * time.Second)  // the last minute might contains data that is not yet finalized
	timeFrom = timeFrom.Add(time.Duration(-timeFrom.Second()) * time.Second) // the seconds need to be stripped to have an effective delay

	aggregation := Aggregation{
		Metric: metric.Name,
	}

	filters := make([]Filter, 0)

	clusterFilters := make([]Filter, 0)
	for _, cluster := range clusters {
		clusterFilters = append(clusterFilters, Filter{
			Field: "resource.kafka.id",
			Op:    "EQ",
			Value: cluster,
		})
	}

	filters = append(filters, Filter{
		Op:      "OR",
		Filters: clusterFilters,
	})

	topicFilters := make([]Filter, 0)
	for _, topic := range topicFiltering {
		topicFilters = append(topicFilters, Filter{
			Field: "metric.topic",
			Op:    "EQ",
			Value: topic,
		})
	}
	if len(topicFilters) > 0 {
		filters = append(filters, Filter{
			Op:      "OR",
			Filters: topicFilters,
		})
	}

	filterHeader := FilterHeader{
		Op:      "AND",
		Filters: filters,
	}

	groupBy := []string{}
	for _, label := range metric.Labels {
		if contains(groupByLabels, label.Key) {
			if resource.hasLabel(label.Key) {
				groupBy = append(groupBy, "resource."+strings.Replace(label.Key, "_", ".", -1))
			} else {
				groupBy = append(groupBy, "metric."+label.Key)
			}
		}
	}

	for _, label := range resource.Labels {
		groupBy = append(groupBy, "resource."+label.Key)
	}

	return Query{
		Aggreations: []Aggregation{aggregation},
		Filter:      filterHeader,
		Granularity: Context.Granularity,
		GroupBy:     groupBy,
		Limit:       1000,
		Intervals:   []string{fmt.Sprintf("%s/%s", timeFrom.Format(time.RFC3339), Context.Granularity)},
	}
}

// BuildConnectorsQuery creates a new Query for a metric for a set of connectors
// This function will return the main global query, override queries will not be generated
func BuildConnectorsQuery(metric MetricDescription, connectors []string, resource ResourceDescription) Query {
	timeFrom := time.Now().Add(time.Duration(-Context.Delay) * time.Second)  // the last minute might contains data that is not yet finalized
	timeFrom = timeFrom.Add(time.Duration(-timeFrom.Second()) * time.Second) // the seconds need to be stripped to have an effective delay

	aggregation := Aggregation{
		Metric: metric.Name,
	}

	filters := make([]Filter, 0)

	connectorFilters := make([]Filter, 0)
	for _, connector := range connectors {
		connectorFilters = append(connectorFilters, Filter{
			Field: "resource.connector.id",
			Op:    "EQ",
			Value: connector,
		})
	}

	filters = append(filters, Filter{
		Op:      "OR",
		Filters: connectorFilters,
	})

	filterHeader := FilterHeader{
		Op:      "AND",
		Filters: filters,
	}

	groupBy := make([]string, len(resource.Labels))
	for i, rsrcLabel := range resource.Labels {
		groupBy[i] = "resource." + rsrcLabel.Key
	}

	return Query{
		Aggreations: []Aggregation{aggregation},
		Filter:      filterHeader,
		Granularity: Context.Granularity,
		GroupBy:     groupBy,
		Limit:       1000,
		Intervals:   []string{fmt.Sprintf("%s/%s", timeFrom.Format(time.RFC3339), Context.Granularity)},
	}
}

// BuildKsqlQuery creates a new Query for a metric for a specific ksql application
// This function will return the main global query, override queries will not be generated
func BuildKsqlQuery(metric MetricDescription, ksqlAppIds []string, resource ResourceDescription) Query {
	timeFrom := time.Now().Add(time.Duration(-Context.Delay) * time.Second)  // the last minute might contains data that is not yet finalized
	timeFrom = timeFrom.Add(time.Duration(-timeFrom.Second()) * time.Second) // the seconds need to be stripped to have an effective delay

	aggregation := Aggregation{
		Metric: metric.Name,
	}

	filters := make([]Filter, 0)

	connectorFilters := make([]Filter, 0)
	for _, ksqlID := range ksqlAppIds {
		connectorFilters = append(connectorFilters, Filter{
			Field: "resource.ksql.id",
			Op:    "EQ",
			Value: ksqlID,
		})
	}

	filters = append(filters, Filter{
		Op:      "OR",
		Filters: connectorFilters,
	})

	filterHeader := FilterHeader{
		Op:      "AND",
		Filters: filters,
	}

	groupBy := make([]string, len(resource.Labels))
	for i, rsrcLabel := range resource.Labels {
		groupBy[i] = "resource." + rsrcLabel.Key
	}

	return Query{
		Aggreations: []Aggregation{aggregation},
		Filter:      filterHeader,
		Granularity: Context.Granularity,
		GroupBy:     groupBy,
		Limit:       1000,
		Intervals:   []string{fmt.Sprintf("%s/%s", timeFrom.Format(time.RFC3339), Context.Granularity)},
	}
}

// BuildSchemaRegistryQuery creates a new Query for a metric for a specific schema registry id
// This function will return the main global query, override queries will not be generated
func BuildSchemaRegistryQuery(metric MetricDescription, schemaregistries []string, resource ResourceDescription) Query {
	timeFrom := time.Now().Add(time.Duration(-Context.Delay) * time.Second)  // the last minute might contains data that is not yet finalized
	timeFrom = timeFrom.Add(time.Duration(-timeFrom.Second()) * time.Second) // the seconds need to be stripped to have an effective delay

	aggregation := Aggregation{
		Metric: metric.Name,
	}

	filters := make([]Filter, 0)

	connectorFilters := make([]Filter, 0)
	for _, schemaRegistryID := range schemaregistries {
		connectorFilters = append(connectorFilters, Filter{
			Field: "resource.schema_registry.id",
			Op:    "EQ",
			Value: schemaRegistryID,
		})
	}

	filters = append(filters, Filter{
		Op:      "OR",
		Filters: connectorFilters,
	})

	filterHeader := FilterHeader{
		Op:      "AND",
		Filters: filters,
	}

	groupBy := make([]string, len(resource.Labels))
	for i, rsrcLabel := range resource.Labels {
		groupBy[i] = "resource." + rsrcLabel.Key
	}

	return Query{
		Aggreations: []Aggregation{aggregation},
		Filter:      filterHeader,
		Granularity: Context.Granularity,
		GroupBy:     groupBy,
		Limit:       1000,
		Intervals:   []string{fmt.Sprintf("%s/%s", timeFrom.Format(time.RFC3339), Context.Granularity)},
	}
}

// SendQuery sends a query to Confluent Cloud API metrics and wait for the response synchronously
func SendQuery(query Query) (QueryResponse, error) {
	jsonQuery, err := json.Marshal(query)
	if err != nil {
		log.WithError(err).Errorln("Failed serialize query in JSON")
		return QueryResponse{}, errors.New("failed serializing query in JSON")
	}
	endpoint := Context.HTTPBaseURL + queryURI
	req := MustGetNewRequest("POST", endpoint, bytes.NewBuffer(jsonQuery))

	res, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Errorln("Failed to send query")
		return QueryResponse{}, err
	}

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		if IsFatal(res) {
			log.WithFields(log.Fields{"StatusCode": res.StatusCode, "Endpoint": endpoint, "body": string(body)}).Fatalln("Stopping the exporter due to fatal issue while querying the Metrics API")
		}
		if res.StatusCode == 429 {
			log.WithFields(log.Fields{
				"StatusCode": res.StatusCode,
				"Endpoint":   endpoint,
				"body":       string(body),
				"reasons":    "You probably scrape the ccloudexporter too frequently, you should probably increase Prometheus `scrape_interval`",
			}).Errorln("Received invalid response")
			errorMsg := fmt.Sprintf("Received status code %d instead of 200 for POST on %s (%s). It generally means that you scrape too frequently, you should probably increase Prometheus `scrape_interval`", res.StatusCode, endpoint, string(body))
			return QueryResponse{}, errors.New(errorMsg)
		}
		log.WithFields(log.Fields{"StatusCode": res.StatusCode, "Endpoint": endpoint, "body": string(body)}).Errorln("Received invalid response")
		errorMsg := fmt.Sprintf("Received status code %d instead of 200 for POST on %s (%s)", res.StatusCode, endpoint, string(body))
		return QueryResponse{}, errors.New(errorMsg)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Errorln("Can not read response")
		return QueryResponse{}, err
	}

	response := QueryResponse{}
	json.Unmarshal(body, &response)

	return response, nil
}

// IsFatal returns true if the http response is not worth retrying
func IsFatal(res *http.Response) bool {
	if res.StatusCode == 403 {
		return true
	}

	if res.StatusCode == 401 {
		return true
	}

	return false
}
