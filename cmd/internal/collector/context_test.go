package collector

//
// context_test.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "testing"

var rule1 = Rule{
	id:            0,
	Topics:        []string{"myTopic"},
	Clusters:      []string{"cluster"},
	Metrics:       []string{"metric", "metric2"},
	GroupByLabels: []string{"topic", "partition"},
}

var rule2 = Rule{
	id:            1,
	Topics:        []string{},
	Clusters:      []string{"cluster"},
	Metrics:       []string{"metric", "metric3"},
	GroupByLabels: []string{"topic"},
}

func TestEnsureResultsAreIgnored(t *testing.T) {
	Context = ExporterContext{
		Rules: []Rule{rule1, rule2},
	}

	if !rule2.ShouldIgnoreResultForRule("myTopic", "cluster", "metric") {
		t.Errorf("Result not ignored but should be as there is an override for a topic")
		t.Fail()
		return
	}

	if !rule2.ShouldIgnoreResultForRule("myTopic", "cluster", "metric") {
		t.Errorf("Result not ignored but should be - likely the cache is invalid")
		t.Fail()
		return
	}
}

func TestEnsureResultsAreNotIgnored(t *testing.T) {
	Context = ExporterContext{
		Rules: []Rule{rule1, rule2},
	}

	if rule2.ShouldIgnoreResultForRule("myTopic", "cluster", "metric3") {
		t.Errorf("Result ignored but should not be as this is another topic")
		t.Fail()
		return
	}
}

func TestGetMetrics(t *testing.T) {
	Context = ExporterContext{
		Rules: []Rule{rule1, rule2},
	}

	if len(Context.GetMetrics()) != 3 {
		t.Errorf("Unexpected number of metric returned: %+v", Context.GetMetrics())
		t.Fail()
	}
}

func TestGetMapOfMetrics(t *testing.T) {
	Context = ExporterContext{
		Rules: []Rule{rule1, rule2},
	}

	if len(Context.GetMapOfMetrics()) != 3 {
		t.Errorf("Unexpected number of metric returned: %+v", Context.GetMapOfMetrics())
		t.Fail()
	}
}