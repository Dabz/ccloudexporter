//
// main.go
// Copyright (C) 2020 gaspar_d d.gasparina@gmail.com
//
// Distributed under terms of the MIT license.
//

package main

import (
	"fmt"
	"github.com/Dabz/ccloudexporter/cmd/internal/scraper"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: %s <cluster id> [kafka client configuration path]\n", os.Args[0])
		fmt.Printf("Note: If a client configuration file is specified, the scraper will connect to Apache Kafka to discover topics\n")
		fmt.Printf("If not specified, only the information from the metric API endpoints will be exposed (which omit topics with no activity)\n")
		fmt.Printf("For further information regarding the creation of the configuration file, please refer to: https://docs.confluent.io/current/cloud/using/config-client.html#librdkafka-based-c-clients\n")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		kafkaConfig := scraper.ParsePropertyFile(os.Args[2])
		adminClient, err := kafka.NewAdminClient(&kafkaConfig)
		if err != nil {
			fmt.Println("Can not instantiate Apache Kafka admin client")
			panic(err)
		}

		_, err = adminClient.GetMetadata(nil, true, 6000)
		if err != nil {
			fmt.Println("Can not fetch initial metadata information from Apache Kafka")
			panic(err)
		}

		scraper.RetrieveTopicListRoutine(adminClient)
	}

	cluster := os.Args[1]
	scraper.FetchMetricsFromEndpointRoutine(cluster)

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening on http://localhost:2112/metrics")
}
