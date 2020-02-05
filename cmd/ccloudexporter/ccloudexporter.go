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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: %s <cluster id>\n", os.Args[0])
		os.Exit(1)
	}

	cluster := os.Args[1]
	collector := scraper.NewCCloudCollector(cluster)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Listening on http://localhost:2112/metrics")
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		panic(err)
	}
}
