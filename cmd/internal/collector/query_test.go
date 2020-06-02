package collector

//
// context_test.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "testing"

func TestBuildQuery(t *testing.T) {
	metric := MetricDescription{
		Name: "io.confluent.kafka.server/retained_bytes",
		Labels: []MetricLabel{{
			Key: "topic",
		}, {
			Key: "cluster_id",
		}, {
			Key: "partition",
		}},
	}

	query := BuildQuery(metric, []string{"cluster"}, []string{"cluster_id", "topic"}, nil)

	if len(query.Filter.Filters) != 1 || len(query.Filter.Filters[0].Filters) != 1 {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Field != "metric.label.cluster_id" {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Value != "cluster" {
		t.Fail()
		return
	}
}

func TestBuildQueryWithTopic(t *testing.T) {
	metric := MetricDescription{
		Name: "io.confluent.kafka.server/retained_bytes",
		Labels: []MetricLabel{{
			Key: "topic",
		}, {
			Key: "cluster_id",
		}, {
			Key: "partition",
		}},
	}

	query := BuildQuery(metric, []string{"cluster"}, []string{"cluster_id", "topic", "partition"}, []string{"topic"})

	if len(query.Filter.Filters) != 2 || len(query.Filter.Filters[1].Filters) != 1 {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Field != "metric.label.cluster_id" {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Value != "cluster" {
		t.Fail()
		return
	}

	if query.Filter.Filters[1].Filters[0].Field != "metric.label.topic" {
		t.Fail()
		return
	}

	if query.Filter.Filters[1].Filters[0].Value != "topic" {
		t.Fail()
		return
	}
}
