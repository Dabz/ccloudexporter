//
// main.go
// Copyright (C) 2020 gaspar_d d.gasparina@gmail.com
//
// Distributed under terms of the MIT license.
//

package main

import (
	"fmt"
	"github.com/Dabz/ccloudexporter/cmd/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	collector.ParseOption()

	ccollector := collector.NewCCloudCollector()
	prometheus.MustRegister(ccollector)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Printf("Listening on http://%s/metrics\n", collector.Context.Listener)
	err := http.ListenAndServe(collector.Context.Listener, nil)
	if err != nil {
		panic(err)
	}
}
