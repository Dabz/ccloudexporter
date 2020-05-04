package collector

//
// option.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "flag"
import "fmt"
import "os"

var (
	HttpTimeout int
	HttpBaseUrl string
	Cluster     string
	Delay       int
	Granularity string
	NoTimestamp bool
)

var supportedGranularity = []string{"PT1M", "PT5M", "PT15M", "PT30M", "PT1H"}

func ParseOption() {
	flag.IntVar(&HttpTimeout, "timeout", 60, "Timeout, in second, to use for all REST call with the Metric API")
	flag.StringVar(&HttpBaseUrl, "endpoint", "https://api.telemetry.confluent.cloud/", "Base URL for the Metric API")
	flag.StringVar(&Granularity, "granularity", "PT1M", "Granularity for the metrics query, by default set to 1 minutes")
	flag.IntVar(&Delay, "delay", 120, "Delay, in seconds, to fetch the metrics. By default set to 120, this, in order to avoid temporary data points.")
	flag.StringVar(&Cluster, "cluster", "", "Cluster ID to fetch metric for (e.g. lkc-xxxxx)")
	flag.BoolVar(&NoTimestamp, "no-timestamp", false, "Do not propagate the timestamp from the the metrics API to prometheus")

	flag.Parse()

	if Cluster == "" {
		Cluster = flag.Arg(0)
	}

	if Cluster == "" {
		fmt.Println("No cluster ID has been specified")
		flag.Usage()
		os.Exit(1)
	}

	if !contains(supportedGranularity, Granularity) {
		fmt.Printf("Granularity %s is invalid \n", Granularity)
		os.Exit(1)
	}

}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
