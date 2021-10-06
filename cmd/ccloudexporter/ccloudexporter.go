//
// main.go
// Copyright (C) 2020 gaspar_d d.gasparina@gmail.com
//
// Distributed under terms of the MIT license.
//

package main

import (
	"fmt"
	"net/http"

	"github.com/Dabz/ccloudexporter/cmd/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	collector.ParseOption()
	log.WithFields(log.Fields{
		"Configuration": fmt.Sprintf("%+v", collector.Context),
	}).Info("ccloudexporter is starting")

	ccollector := collector.NewCCloudCollector()
	prometheus.MustRegister(ccollector)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	log.WithFields(log.Fields{
		"PrometheusEndpoint": fmt.Sprintf("http://%s/metrics", collector.Context.Listener),
	}).Info("ccloudexporter is running")
	err := http.ListenAndServe(collector.Context.Listener, nil)
	if err != nil {
		panic(err)
	}
}
