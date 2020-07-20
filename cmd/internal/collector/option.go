package collector

//
// option.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/nerdynick/confluent-cloud-metrics-go-sdk/ccloudmetrics"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	verbose           bool
	extraVerbose      bool
	extraExtraVerbose bool
)

// ParseOption parses options provided by the CLI and the configuration file
// This function will panic if the options are invalid
func ParseOption() {
	var cluster string
	var configPath string

	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&extraVerbose, "vv", false, "Extra Verbose output")
	flag.BoolVar(&extraExtraVerbose, "vvv", false, "Extra Extra Verbose output")
	flag.StringVar(&configPath, "config", "", "Path to configuration file used to override default behavior of ccloudexporter")
	flag.IntVar(&Context.HTTPTimeout, "timeout", 60, "Timeout, in second, to use for all REST call with the Metric API")
	flag.StringVar(&Context.HTTPBaseURL, "endpoint", ccloudmetrics.DefaultBaseURL, "Base URL for the Metric API")
	flag.StringVar(&Context.Granularity, "granularity", ccloudmetrics.GranularityOneMin.String(), fmt.Sprintf("Granularity for the metrics query, by default set to 1 minutes. Available options are `%s`", strings.Join(ccloudmetrics.AvailableGranularities, ", ")))
	flag.IntVar(&Context.Delay, "delay", 120, "Delay, in seconds, to fetch the metrics. By default set to 120, this, in order to avoid temporary data points.")
	flag.StringVar(&cluster, "cluster", "", "Cluster ID to fetch metric for. If not specified, the environment variable CCLOUD_CLUSTER will be used")
	flag.StringVar(&Context.Listener, "listener", ":2112", "Listener for the HTTP interface")
	flag.BoolVar(&Context.NoTimestamp, "no-timestamp", false, "Do not propagate the timestamp from the the metrics API to prometheus")
	versionFlag := flag.Bool("version", false, "Print the current version and exit")

	flag.Parse()

	//Setup Logging Levels
	if verbose || extraVerbose || extraExtraVerbose {
		log.SetLevel(log.InfoLevel)

		if extraVerbose {
			log.SetLevel(log.DebugLevel)
		}

		if extraExtraVerbose {
			log.SetLevel(log.TraceLevel)
		}
	} else {
		log.SetLevel(log.WarnLevel)
	}

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	if cluster == "" {
		clusterEnv, present := os.LookupEnv("CCLOUD_CLUSTER")
		if present {
			cluster = clusterEnv
		}
	}

	if configPath != "" {
		parseConfigFile(configPath)
	} else {
		createDefaultRule(cluster)
	}
	validateConfiguration()
}

func validateConfiguration() {
	log.WithField("Context", Context).Debug("Pre-Validation Context")
	if !contains(ccloudmetrics.AvailableGranularities, Context.Granularity) {
		fmt.Printf("Granularity %s is invalid \n", Context.Granularity)
		os.Exit(1)
	}

	for i, rule := range Context.Rules {
		//Check Cluster
		if len(rule.Clusters) == 0 {
			fmt.Println("No cluster ID has been specified in a rule")
			flag.Usage()
			os.Exit(1)
		}

		//Check Labels
		if len(rule.GroupByLabels) == 0 {
			fmt.Println("Labels are required while defining a rule")
			os.Exit(1)
		}
		if contains(rule.GroupByLabels, "partition") && len(rule.Topics) == 0 {
			fmt.Println("Topic filtering is required while grouping per partition")
			os.Exit(1)
		}

		whitelistedLabels := []ccloudmetrics.MetricLabel{}
		for _, l := range rule.GroupByLabels {
			label := ccloudmetrics.NewMetricLabel(l)
			if !label.IsValid() {
				log.WithFields(log.Fields{
					"Label":           l,
					"AvailableLabels": ccloudmetrics.AvailableMetricLabels,
				}).Error("Invalid label")
				os.Exit(1)
			}
			whitelistedLabels = append(whitelistedLabels, label)
		}
		Context.Rules[i].WhitelistedLabels = whitelistedLabels

		log.WithFields(log.Fields{
			"GroupLabels":       rule.GroupByLabels,
			"WhitelistedLabels": rule.WhitelistedLabels,
		}).Debug("Rule Labels->Whitelist")

		//Check Rules
		if len(rule.Topics) > 100 {
			fmt.Println("A rule can not have more than 100 topics")
			fmt.Println("Note: Dispatching your topics over multiple rule should fix this issue")
			os.Exit(1)
		}
	}

	log.WithField("Context", Context).Info("Final Context")
}

func parseConfigFile(configPath string) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()

	if err != nil {
		log.Panic("Can not read configuration file")
	}

	setIntIfExit(&Context.Delay, "config.delay")
	setStringIfExit(&Context.Granularity, "config.granularity")
	setStringIfExit(&Context.Listener, "config.listener")
	setStringIfExit(&Context.HTTPBaseURL, "config.http.baseUrl")
	setIntIfExit(&Context.HTTPTimeout, "config.http.timeout")
	setBoolIfExist(&Context.NoTimestamp, "config.noTimestamp")

	viper.UnmarshalKey("rules", &Context.Rules)
	for i, rule := range Context.Rules {
		rule.id = i
		Context.Rules[i] = rule
	}
}

func createDefaultRule(cluster string) {
	if cluster == "" {
		log.Panic("No Cluster was provided")
	}

	Context.Rules = make([]Rule, 1)
	Context.Rules[0] = Rule{
		id:            0,
		Clusters:      []string{cluster},
		Metrics:       DefaultMetrics,
		GroupByLabels: DefaultGroupingLabels,
		Topics:        []string{},
	}
}

func setStringIfExit(destination *string, key string) {
	if viper.Get(key) != nil {
		*destination = viper.GetString(key)
	}
}

func setStringSliceIfExist(destination *[]string, key string) {
	if viper.Get(key) != nil {
		*destination = viper.GetStringSlice(key)
	}
}

func setIntIfExit(destination *int, key string) {
	if viper.Get(key) != nil {
		*destination = viper.GetInt(key)
	}
}

func setBoolIfExist(destination *bool, key string) {
	if viper.Get(key) != nil {
		*destination = viper.GetBool(key)
	}
}

func printVersion() {
	fmt.Printf("ccloudexporter: %s\n", Version)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
