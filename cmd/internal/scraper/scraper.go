//
// main.go
// Copyright (C) 2020 gaspar_d d.gasparina@gmail.com
//
// Distributed under terms of the MIT license.
//

package scraper

import (
	"bufio"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"os"
	"strings"
	"time"
)

var (
	topicList    = make(map[string]*kafka.TopicMetadata)
	gaugeMetrics = make(map[string]*prometheus.GaugeVec)
)

func RetrieveTopicListRoutine(adminClient *kafka.AdminClient) {
	go func(adminClient *kafka.AdminClient) {
		for {
			metadata, err := adminClient.GetMetadata(nil, true, 15000)
			if err != nil {
				fmt.Printf("Can not retrieve metadata information\n")
				panic(err)
			}

			for topicName, topicMetadata := range metadata.Topics {
				topicList[topicName] = &topicMetadata
			}

			time.Sleep(time.Minute)
		}
	}(adminClient)
}

func handleResponseMetric(cluster string, response QueryResponse, vec *prometheus.GaugeVec) {
	exploredTopic := make(map[string]int)
	for _, data := range response.Data {
		_, exist := topicList[data.Topic]
		if !exist {
			topicList[data.Topic] = nil
		}
		exploredTopic[data.Topic] = 0
		vec.With(prometheus.Labels{"topic": data.Topic, "cluster": cluster}).Set(data.Value)
	}

	for topic := range topicList {
		_, exist := exploredTopic[topic]
		if !exist {
			vec.With(prometheus.Labels{"topic": topic, "cluster": cluster}).Set(0.0)
		}
	}
}

// Start a routine what will scrape the Confluent Cloud
// Metrics every 15 seconds
func FetchMetricsFromEndpointRoutine(cluster string) {
	// Before starting the routing, invoking the descriptors endpoints
	// to fetch the list of metric to scap
	descriptorResponse := SendDescriptorQuery()
	for _, metric := range descriptorResponse.Data {
		if metric.Type == "GAUGE_INT64" {
			gaugeMetrics[metric.Name] = promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "ccloud_metric_" + GetNiceNameForMetric(metric),
				Help: metric.Description,
			}, []string{"topic", "cluster"})
		}
	}

	// Main routine scraping the Confluent Cloud API endpoint
	go func() {
		for {
			from := time.Now().Add(time.Minute * -5)
			to := time.Now().Add(time.Minute)
			for metric, gauge := range gaugeMetrics {
				query := BuildQuery(metric, cluster, from, to)
				response := SendQuery(query)
				handleResponseMetric(cluster, response, gauge)
			}

			time.Sleep(time.Second * 15)
		}
	}()
}

func ParsePropertyFile(path string) kafka.ConfigMap {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	reader := bufio.NewReader(f)

	var config = kafka.ConfigMap{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		splitResults := strings.SplitN(line, "=", 2)
		if len(splitResults) < 2 {
			fmt.Printf("can not parse line: %s\n", line)
			os.Exit(1)
		}
		config.SetKey(splitResults[0], splitResults[1])
	}

	return config
}
