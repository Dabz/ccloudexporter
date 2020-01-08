//
// main.go
// Copyright (C) 2020 gaspar_d d.gasparina@gmail.com
//
// Distributed under terms of the MIT license.
//

package scrapper

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
	topicList = make(map[string]*kafka.TopicMetadata)

	receivedBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ccloud_metric_received_bytes",
		Help: "The delta count of bytes received from the network. Each sample is the number of bytes received since the previous data sample. The count is sampled every 60 seconds.",
	}, []string{"topic", "cluster"})
	sentBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ccloud_metric_sent_bytes",
		Help: "The delta count of bytes sent over the network. Each sample is the number of bytes sent since the previous data point. The count is sampled every 60 seconds.",
	}, []string{"topic", "cluster"})
	retainedBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ccloud_metric_retained_bytes",
		Help: "The current count of bytes retained by the cluster, summed across all partitions. The count is sampled every 60 seconds.",
	}, []string{"topic", "cluster"})
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

func handleResponseMetric(cluster string, response Response, vec *prometheus.GaugeVec) {
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

func FetchMetricsFromEndpointRoutine(cluster string) {
	go func() {
		for {
			sentBytesQuery, receivedBytesQuery, retainedBytesQuery := BuildQueries(cluster)
			sentBytesResponse := SendQuery(sentBytesQuery)
			receivedBytesResponse := SendQuery(receivedBytesQuery)
			retainedBytesResponse := SendQuery(retainedBytesQuery)

			handleResponseMetric(cluster, sentBytesResponse, sentBytes)
			handleResponseMetric(cluster, receivedBytesResponse, receivedBytes)
			handleResponseMetric(cluster, retainedBytesResponse, retainedBytes)

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
