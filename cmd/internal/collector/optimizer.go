package collector

//
// query.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

// OptimizeQuery try to optimize the query to make it more
// lightweight and performant for the Metrics API
//
// This function also returns a list of labels that
// need to be added to the Prometheus result
// This is required as optimization might remove potential useful
// labels (e.g. cluster_id will be removed if there is only cluster, thus
// no need to group_by the result)
func OptimizeQuery(input Query) (Query, map[string]string) {
	optimizedQuery, optimizedLabel := removeSuperfluousGroupBy(input)
	return optimizedQuery, optimizedLabel
}

func removeSuperfluousGroupBy(input Query) (Query, map[string]string) {
	optimizedGroupByList := make([]string, 0)
	labels := make(map[string]string)
	for _, groupBy := range input.GroupBy {
		filters := equalityFiltering(input.Filter.Filters, groupBy, input.Filter.Op)
		if len(filters) != 1 {
			optimizedGroupByList = append(optimizedGroupByList, groupBy)
		} else {
			labels[groupBy] = filters[0].Value
		}
	}

	input.GroupBy = optimizedGroupByList
	return input, labels
}

func equalityFiltering(filters []Filter, metric string, parentOp string) []Filter {
	possibilities := make([]Filter, 0)
	for _, filter := range filters {
		if len(filter.Filters) > 0 {
			possibilities = append(possibilities, equalityFiltering(filter.Filters, metric, filter.Op)...)
			continue
		}

		if filter.Field == metric && parentOp == "OR" && filter.Op == "EQ" {
			possibilities = append(possibilities, filter)
		}
	}
	return possibilities
}
