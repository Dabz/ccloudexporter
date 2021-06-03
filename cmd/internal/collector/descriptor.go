package collector

//
// descriptor.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "strings"
import "encoding/json"
import "io/ioutil"
import log "github.com/sirupsen/logrus"

// DescriptorMetricResponse is the response from Confluent Cloud API metric endpoint
// This is the JSON structure for the endpoint
// https://api.telemetry.confluent.cloud/v1/metrics/cloud/descriptors
type DescriptorMetricResponse struct {
	Data []MetricDescription `json:"data"`
}

// MetricDescription is the metric from the https://api.telemetry.confluent.cloud/v1/metrics/cloud/descriptors
// response
type MetricDescription struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Unit        string        `json:"unit"`
	Description string        `json:"description"`
	Labels      []MetricLabel `json:"labels"`
}

// MetricLabel is the label of a metric, should contain a key and a description
// e.g.
//  {
//      "description": "Name of the Kafka topic",
//      "key": "topic"
//  }
type MetricLabel struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// DescriptorResourceResponse is the result of the Metrics API resource description
type DescriptorResourceResponse struct {
	Data []ResourceDescription `json:"data"`
}

// ResourceDescription describes one resource returned by the Metrics API
type ResourceDescription struct {
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Labels      []MetricLabel `json:"labels"`
}

var (
	excludeListForMetric = map[string]string{
		"io.confluent.kafka.server":  "",
		"io.confluent.kafka.connect": "",
		"io.confluent.kafka.ksql":    "",
		"delta":                      "",
	}
	descriptorURI         = "v2/metrics/cloud/descriptors/metrics"
	descriptorResourceURI = "v2/metrics/cloud/descriptors/resources"
)

// Return true if the metric has this label
func (metric MetricDescription) hasLabel(label string) bool {
	for _, l := range metric.Labels {
		if l.Key == label {
			return true
		}
	}
	return false
}

// Return true if the resource has this label
func (resource ResourceDescription) hasLabel(label string) bool {
	stripLabel := strings.Replace(strings.Replace(label, "resource.", "", 1), ".", "_", -1)
	for _, l := range resource.Labels {
		stripKey := strings.Replace(strings.Replace(l.Key, "resource.", "", 1), ".", "_", -1)
		if stripKey == stripLabel {
			return true
		}
	}
	return false
}

func (resource ResourceDescription) datapointFieldNameForLabel(label string) string {
	if resource.hasLabel(label) {
		return "resource." + strings.Replace(label, "_", ".", -1)
	}
	return "metric." + label
}

// GetNiceNameForMetric returns a human friendly metric name from a Confluent Cloud API metric
func GetNiceNameForMetric(metric MetricDescription) string {
	splits := strings.Split(metric.Name, "/")
	for _, split := range splits {
		_, contain := excludeListForMetric[split]
		if !contain {
			return split
		}
	}

	log.WithField("metric", metric).Fatalln("Invalid metric")
	panic(nil)
}

// GetPrometheusNameForLabel returns a prometheus friendly name for a label
func GetPrometheusNameForLabel(label string) string {
	return strings.Join(strings.Split(label, "."), "_")
}

// SendDescriptorQuery calls the https://api.telemetry.confluent.cloud/v2/metrics/cloud/descriptors endpoint
// to retrieve the list of metrics
func SendDescriptorQuery(ressourceType string) DescriptorMetricResponse {
	endpoint := Context.HTTPBaseURL + descriptorURI + "?resource_type=" + ressourceType
	req := MustGetNewRequest("GET", endpoint, nil)

	res, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Fatalln("HTTP query for the descriptor endpoint failed")
	}

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		log.WithFields(log.Fields{"StatusCode": res.StatusCode, "Endpoint": endpoint, "body": body}).Fatalf("Received status code %d instead of 200 for GET on %s. \n\n%s\n\n", res.StatusCode, endpoint, body)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Fatalln("Can not read the content of the descriptor query")
	}

	response := DescriptorMetricResponse{}
	json.Unmarshal(body, &response)

	return response
}

// SendResourceDescriptorQuery calls the https://api.telemetry.confluent.cloud/v2/metrics/cloud/descriptors endpoint
// to retrieve the list of available resources
func SendResourceDescriptorQuery() DescriptorResourceResponse {
	endpoint := Context.HTTPBaseURL + descriptorResourceURI
	req := MustGetNewRequest("GET", endpoint, nil)

	res, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Fatalln("HTTP query for the descriptor endpoint failed")
	}

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		log.WithFields(log.Fields{"StatusCode": res.StatusCode, "Endpoint": endpoint, "body": body}).Fatalf("Received status code %d instead of 200 for GET on %s. \n\n%s\n\n", res.StatusCode, endpoint, body)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Fatalln("Can not read the content of the descriptor query")
	}

	response := DescriptorResourceResponse{}
	json.Unmarshal(body, &response)

	return response
}
