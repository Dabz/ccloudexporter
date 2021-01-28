package collector

//
// context_test.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "testing"
import "strings"
import "time"

var (
	resource = ResourceDescription{
		Type:        "kafka",
		Description: "",
		Labels: []MetricLabel{
			{
				Key: "kafka.id",
			},
		},
	}
)

func TestBuildQuery(t *testing.T) {
	metric := MetricDescription{
		Name: "io.confluent.kafka.server/retained_bytes",
		Labels: []MetricLabel{{
			Key: "topic",
		}, {
			Key: "kafka_id",
		}, {
			Key: "partition",
		}},
	}

	query := BuildQuery(metric, []string{"cluster"}, []string{"kafka_id", "topic"}, nil, resource)

	if len(query.Filter.Filters) != 1 || len(query.Filter.Filters[0].Filters) != 1 {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Field != "resource.kafka.id" {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Value != "cluster" {
		t.Fail()
		return
	}

	if len(query.Intervals) == 0 {
		t.Fail()
		return
	}

	timeFrom, _ := time.Parse(time.RFC3339, strings.Split(query.Intervals[0], "/")[0])
	if timeFrom.Second() != 0 {
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
			Key: "kafka_id",
		}, {
			Key: "partition",
		}},
	}

	query := BuildQuery(metric, []string{"cluster"}, []string{"kafka_id", "topic", "partition"}, []string{"topic"}, resource)

	if len(query.Filter.Filters) != 2 || len(query.Filter.Filters[1].Filters) != 1 {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Field != "resource.kafka.id" {
		t.Fail()
		return
	}

	if query.Filter.Filters[0].Filters[0].Value != "cluster" {
		t.Fail()
		return
	}

	if query.Filter.Filters[1].Filters[0].Field != "metric.topic" {
		t.Fail()
		return
	}

	if query.Filter.Filters[1].Filters[0].Value != "topic" {
		t.Fail()
		return
	}
}

func TestOptimizationRemoveSuperfelousGroupBy(t *testing.T) {
	metric := MetricDescription{
		Name: "io.confluent.kafka.server/retained_bytes",
		Labels: []MetricLabel{{
			Key: "topic",
		}, {
			Key: "kafka_id",
		}, {
			Key: "partition",
		}},
	}

	query, _ := OptimizeQuery(BuildQuery(metric, []string{"cluster"}, []string{"kafka_id", "topic"}, nil, resource))

	if len(query.GroupBy) > 1 {
		t.Errorf("Unexepected groupBy list: %s\n", query.GroupBy)
		t.Fail()
		return
	}
}

func TestOptimizationDoesNotRemoveRequiredGroupBy(t *testing.T) {
	metric := MetricDescription{
		Name: "io.confluent.kafka.server/retained_bytes",
		Labels: []MetricLabel{{
			Key: "topic",
		}, {
			Key: "kafka_id",
		}, {
			Key: "partition",
		}},
	}

	query, _ := OptimizeQuery(BuildQuery(metric, []string{"cluster1", "cluster2"}, []string{"kafka_id", "topic"}, nil, resource))

	if len(query.GroupBy) <= 1 {
		t.Errorf("Unexepected groupBy list: %s\n", query.GroupBy)
		t.Fail()
		return
	}
}
