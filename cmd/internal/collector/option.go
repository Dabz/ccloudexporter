package collector

//
// option.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"flag"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var supportedGranularity = []string{"PT1M", "PT5M", "PT15M", "PT30M", "PT1H"}

// ParseOption parses options provided by the CLI and the configuration file
// This function will panic if the options are invalid
func ParseOption() {
	var cluster string
	var configPath string

	flag.StringVar(&configPath, "config", "", "Path to configuration file used to override default behavior of ccloudexporter")
	flag.IntVar(&Context.HTTPTimeout, "timeout", 60, "Timeout, in second, to use for all REST call with the Metric API")
	flag.StringVar(&Context.HTTPBaseURL, "endpoint", "https://api.telemetry.confluent.cloud/", "Base URL for the Metric API")
	flag.StringVar(&Context.Granularity, "granularity", "PT1M", "Granularity for the metrics query, by default set to 1 minutes")
	flag.IntVar(&Context.Delay, "delay", 120, "Delay, in seconds, to fetch the metrics. By default set to 120, this, in order to avoid temporary data points.")
	flag.StringVar(&cluster, "cluster", "", "Cluster ID to fetch metric for. If not specified, the environment variable CCLOUD_CLUSTER will be used")
	flag.StringVar(&Context.Listener, "listener", ":2112", "Listener for the HTTP interface")
	flag.BoolVar(&Context.NoTimestamp, "no-timestamp", false, "Do not propagate the timestamp from the the metrics API to prometheus")
	versionFlag := flag.Bool("version", false, "Print the current version and exit")
	verboseFlag := flag.Bool("verbose", false, "Print trace level logs to stdout")

	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
	log.SetOutput(os.Stdout)
	if *verboseFlag {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
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

// MustGetAPIKey returns the API Key from environment variables
// if an API Key can not be find, it exits the process
func MustGetAPIKey() string {
	key, present := os.LookupEnv("CCLOUD_API_KEY")
	if present && key != "" {
		return key
	}

	key, present = os.LookupEnv("CCLOUD_USER")
	if present && key != "" {
		return key
	}

	log.Fatalln("CCLOUD_API_KEY environment variable has not been specified")
	panic(nil)
}

// MustGetAPISecret returns the API Key from environment variables
// if an API Key can not be find, it exits the process
func MustGetAPISecret() string {
	secret, present := os.LookupEnv("CCLOUD_API_SECRET")
	if present && secret != "" {
		return secret
	}

	secret, present = os.LookupEnv("CCLOUD_PASSWORD")
	if present && secret != "" {
		return secret
	}

	log.Fatalln("CCLOUD_API_SECRET environment variable has not been specified")
	panic(nil)
}

func validateConfiguration() {

	if !contains(supportedGranularity, Context.Granularity) {
		log.WithFields(log.Fields{"granularity": Context.Granularity}).Fatalf("Granularity %s is invalid\n", Context.Granularity)
	}

	for _, rule := range Context.Rules {
		if len(rule.Clusters) == 0 {
			log.Errorln("No cluster ID has been specified in a rule")
			flag.Usage()
			os.Exit(1)
		}

		if contains(rule.GroupByLabels, "partition") && len(rule.Topics) == 0 {
			log.Fatalln("Topic filtering is required while grouping per partition")
		}

		if len(rule.Topics) > 100 {
			log.Errorln("A rule can not have more than 100 topics")
			log.Fatalln("Note: Dispatching your topics over multiple rule should fix this issue")
		}

		if len(rule.GroupByLabels) == 0 {
			log.Fatalln("Labels is required while defining a rule")
		}

		for _, currentRegex := range rule.excludeTopicsRegex {
			regexp.MustCompile(currentRegex)
		}
	}
}

func parseConfigFile(configPath string) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()

	if err != nil {
		log.WithError(err).Fatalln("Can not read configuration file")
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
	Context.Rules = make([]Rule, 1)
	Context.Rules[0] = Rule{
		id:            0,
		Clusters:      []string{cluster},
		Metrics:       DefaultMetrics,
		GroupByLabels: DefaultGroupingLabels,
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
	log.WithField("version", Version).Println()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
