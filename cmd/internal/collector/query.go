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
import "os"
import "errors"
import "io/ioutil"
import "encoding/json"
import "net/http"

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
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

// Response from the cloud endpoint
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
	queryUri = "/v1/metrics/cloud/query"
)

// Create a new Query for a metric for a specific cluster and time interval
func BuildQuery(metric MetricDescription, cluster string) Query {
	timeFrom := time.Now().Add(time.Duration(Delay*-1) * time.Second) // the last minute might contains data that is not yet finalized

	aggregation := Aggregation{
		Agg:    "SUM",
		Metric: metric.Name,
	}

	filter := Filter{
		Field: "metric.label.cluster_id",
		Op:    "EQ",
		Value: cluster,
	}

	filterHeader := FilterHeader{
		Op:      "AND",
		Filters: []Filter{filter},
	}

	groupBy := []string{}
	for _, label := range metric.Labels {
		if label.Key == "partition" {
			continue
		}
		groupBy = append(groupBy, "metric.label."+label.Key)
	}

	return Query{
		Aggreations: []Aggregation{aggregation},
		Filter:      filterHeader,
		Granularity: Granularity,
		GroupBy:     groupBy,
		Limit:       1000,
		Intervals:   []string{fmt.Sprintf("%s/%s", timeFrom.Format(time.RFC3339), Granularity)},
	}
}

// Send Query to Confluent Cloud API metrics and wait for the response synchronously
func SendQuery(query Query) (QueryResponse, error) {
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

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	endpoint := HttpBaseUrl + queryUri
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonQuery))
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(user, password)
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf(err.Error())
		return QueryResponse{}, err
	}

	if res.StatusCode != 200 {
		errorMsg := fmt.Sprintf("Received status code %d instead of 200 for POST on %s with %s", res.StatusCode, endpoint, jsonQuery)
		return QueryResponse{}, errors.New(errorMsg)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	response := QueryResponse{}
	json.Unmarshal(body, &response)

	return response, nil
}
