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

	collector := collector.NewCCloudCollector()
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Listening on http://localhost:2112/metrics")
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		panic(err)
	}
}
