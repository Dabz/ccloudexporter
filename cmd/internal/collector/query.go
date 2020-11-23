package collector

//
// query.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "time"
import "fmt"
import "bytes"
import "errors"
import "io/ioutil"
import "encoding/json"
import log "github.com/sirupsen/logrus"

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
	Agg    string `json:"agg"`
	Metric string `json:"metric"`
}

// FilterHeader to use for a query
type FilterHeader struct {
	Op      string   `json:"op"`
	Filters []Filter `json:"filters"`
}

// Filter structure
type Filter struct {
	Field       string   `json:"field,omitempty"`
	Op          string   `json:"op"`
	Value       string   `json:"value,omitempty"`
	Filters     []Filter `json:"filters,omitempty"`
	UnaryFilter *Filter  `json:"filter,omitempty"`
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
	queryURI = "/v1/metrics/cloud/query"
)

// BuildQuery creates a new Query for a metric for a specific cluster and time interval
// This function will return the main global query, override queries will not be generated
func BuildQuery(metric MetricDescription, clusters []string, groupByLabels []string, topicFiltering []string, excludeTopics []string) Query {
	timeFrom := time.Now().Add(time.Duration(-Context.Delay) * time.Second)  // the last minute might contains data that is not yet finalized
	timeFrom = timeFrom.Add(time.Duration(-timeFrom.Second()) * time.Second) // the seconds need to be stripped to have an effective delay

	aggregation := Aggregation{
		Agg:    "SUM",
		Metric: metric.Name,
	}

	filters := make([]Filter, 0)

	clusterFilters := make([]Filter, 0)
	for _, cluster := range clusters {
		clusterFilters = append(clusterFilters, Filter{
			Field: "metric.label.cluster_id",
			Op:    "EQ",
			Value: cluster,
		})
	}

	filters = append(filters, Filter{
		Op:      "OR",
		Filters: clusterFilters,
	})

	// Remove topics from the topicFiltering list if they also exist in the ExcludeTopics list
	eligibleTopics := make([]string,0)
	if len(topicFiltering) > 0 && len(excludeTopics) > 0{
		for _, inTopic := range topicFiltering {
			if !contains(excludeTopics,inTopic) {
				eligibleTopics = append(eligibleTopics,inTopic)
			}
		}
		topicFiltering = eligibleTopics
	}

	topicFilters := make([]Filter, 0)
	for _, topic := range topicFiltering {
		topicFilters = append(topicFilters, Filter{
			Field: "metric.label.topic",
			Op:    "EQ",
			Value: topic,
		})
	}

	// Exclude topics filter, this isn't necessary if a list of topics is already defined
	// A list of topics provided is by definition filtering other non-wanted topics
	if len(topicFiltering) == 0 {
		excludeTopicFilters := []Filter{}
		for _, exTopic := range excludeTopics {
			excludeFilter := Filter {
				Field: "metric.label.topic",
				Op:    "EQ",
				Value: exTopic,
			}
			wrapperNotFilter := Filter{
				Op:          "NOT",
				UnaryFilter: &excludeFilter,
			}
			excludeTopicFilters = append(excludeTopicFilters, wrapperNotFilter)
		}
		if len(excludeTopicFilters) > 0 {
			filters = append(filters, Filter{
				Op:      "AND",
				Filters: excludeTopicFilters,
			})
		}
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
			groupBy = append(groupBy, "metric.label."+label.Key)
		}
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
		return QueryResponse{}, errors.New("Failed serialize query in JSON")
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