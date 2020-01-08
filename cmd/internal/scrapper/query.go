//
// query.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

package scrapper

import "time"
import "fmt"
import "bytes"
import "os"
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
type Response struct {
	Data []Data `json:"data"`
}

type Data struct {
	Topic     string  `json:"metric.label.topic"`
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

var (
	endpoint   = "https://api.telemetry.confluent.cloud/v1/metrics/cloud/query"
	httpClient = http.Client{
		Timeout: time.Second * 6,
	}
)

func BuildQueries(cluster string) (sentBytes Query, receivedBytes Query, retainedBytes Query) {
	from := time.Now().Add(time.Minute * -1)
	to := time.Now().Add(time.Minute)

	sentBytes = BuildQuery("io.confluent.kafka.server/sent_bytes/delta", cluster, from, to)
	receivedBytes = BuildQuery("io.confluent.kafka.server/received_bytes/delta", cluster, from, to)
	retainedBytes = BuildQuery("io.confluent.kafka.server/retained_bytes", cluster, from, to)

	return
}

func BuildQuery(metric string, cluster string, timeFrom time.Time, timeTo time.Time) Query {
	aggregation := Aggregation{
		Agg:    "SUM",
		Metric: metric,
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

	return Query{
		Aggreations: []Aggregation{aggregation},
		Filter:      filterHeader,
		Granularity: "PT1M",
		GroupBy:     []string{"metric.label.topic"},
		Limit:       2,
		Intervals:   []string{fmt.Sprintf("%s/%s", timeFrom.Format(time.RFC3339), timeTo.Format(time.RFC3339))},
	}
}

func SendQuery(query Query) Response {
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
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonQuery))
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(user, password)
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		fmt.Printf("Received status code %d instead of 200", res.StatusCode)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	response := Response{}
	json.Unmarshal(body, &response)

	return response
}
