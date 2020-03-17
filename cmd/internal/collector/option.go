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
)

func ParseOption() {
	flag.IntVar(&HttpTimeout, "timeout", 60, "Timeout, in second, to use for all REST call with the Metric API")
	flag.StringVar(&HttpBaseUrl, "endpoint", "https://api.telemetry.confluent.cloud/", "Base URL for the Metric API")
	flag.StringVar(&Cluster, "cluster", "", "Cluster ID to fetch metric for (e.g. lkc-xxxxx)")

	flag.Parse()

	if Cluster == "" {
		Cluster = flag.Arg(0)
	}

	if Cluster == "" {
		fmt.Println("No cluster ID has been specified")
		flag.Usage()
		os.Exit(1)
	}

}
