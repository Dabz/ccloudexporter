package collector

//
// context_test.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "testing"

func TestGetNiceNameForMetric(t *testing.T) {
	metric := MetricDescription{
		Name: "io.confluent.kafka.server/retained_bytes",
	}
	name := GetNiceNameForMetric(metric)

	if name != "retained_bytes" {
		t.Fail()
		t.Errorf("Expected %s got %s", "retained_bytes", name)
		return
	}
}

func TestHasLabel(t *testing.T) {
	metric := MetricDescription{
		Name: "io.confluent.kafka.server/retained_bytes",
		Labels: []MetricLabel{{
			Key:         "label1",
			Description: "",
		}, {
			Key:         "label2",
			Description: "",
		}},
	}

	if !metric.hasLabel("label1") {
		t.Fail()
	}

	if metric.hasLabel("na") {
		t.Fail()
	}
}
