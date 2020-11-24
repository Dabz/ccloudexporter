package collector

//
// collector_test.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestHandleResponse(t *testing.T) {
	metric := CCloudCollectorMetric{
		labels: []string{"topic", "cluster_id"},
		metric: MetricDescription{Name: "metric"},
		desc:   prometheus.NewDesc("metric", "help", []string{"topic", "cluster_id"}, nil),
	}

	collector := CCloudCollector{
		metrics: map[string]CCloudCollectorMetric{
			"metric": metric,
		},
	}

	var rule = Rule{
		id:            0,
		Topics:        []string{"topic"},
		Clusters:      []string{"cluster"},
		Metrics:       []string{"metric", "metric2"},
		GroupByLabels: []string{"topic", "cluster_id"},
	}

	responseString := `
{
   "data": [
			{
					"metric.label.cluster_id": "cluster",
					"metric.label.topic": "topic",
					"timestamp": "2020-06-03T13:37:00Z",
					"value": 1.0
			},
			{
					"metric.label.cluster_id": "cluster",
					"metric.label.topic": "topic2",
					"timestamp": "2020-06-03T13:37:00Z",
					"value": 1.0
			}
	]
}`

	responseBytes, err := ioutil.ReadAll(strings.NewReader(responseString))
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
		return
	}

	response := QueryResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	pchan := make(chan prometheus.Metric, 10)

	collector.handleResponse(response, metric, pchan, rule, make(map[string]string))

	if len(pchan) != 2 {
		t.Errorf("Invalid number of metrics returned, expected 2 got %d", len(pchan))
		t.Fail()
		return
	}

}

func TestHandleResponseForRegexFiltering(t *testing.T) {
	metric := CCloudCollectorMetric{
		labels: []string{"topic", "cluster_id"},
		metric: MetricDescription{Name: "metric"},
		desc:   prometheus.NewDesc("metric", "help", []string{"topic", "cluster_id"}, nil),
	}

	var rule = Rule{
		id:                 0,
		Topics:             []string{"topic"},
		Clusters:           []string{"cluster"},
		Metrics:            []string{"metric", "metric2"},
		GroupByLabels:      []string{"topic", "cluster_id"},
		ExcludeTopicsRegex: []string{"excludedTopic*", "excludedThing*"},
	}

	// Compile the Regex
	Context.Granularity = "PT1M"
	Context.Rules = []Rule{rule}
	validateConfiguration()

	collector := CCloudCollector{
		metrics: map[string]CCloudCollectorMetric{
			"metric": metric,
		},
	}

	responseString := `
{
   "data": [
			{
					"metric.label.cluster_id": "cluster",
					"metric.label.topic": "topic",
					"timestamp": "2020-06-03T13:37:00Z",
					"value": 1.0
			},
			{
					"metric.label.cluster_id": "cluster",
					"metric.label.topic": "topic2",
					"timestamp": "2020-06-03T13:37:00Z",
					"value": 1.0
			},
			{
					"metric.label.cluster_id": "cluster",
					"metric.label.topic": "excludedTopicA",
					"timestamp": "2020-06-03T13:37:00Z",
					"value": 1.0
			},
			{
					"metric.label.cluster_id": "cluster",
					"metric.label.topic": "excludedThingB",
					"timestamp": "2020-06-03T13:37:00Z",
					"value": 1.0
			}
	]
}`

	responseBytes, err := ioutil.ReadAll(strings.NewReader(responseString))
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
		return
	}

	response := QueryResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	pchan := make(chan prometheus.Metric, 10)

	collector.handleResponse(response, metric, pchan, rule, make(map[string]string))

	if len(pchan) != 2 {
		t.Errorf("Invalid number of metrics returned, expected 2 got %d", len(pchan))
		t.Fail()
		return
	}

}
