package collector

//
// cache.go
// Copyright (C) 2021 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// CCloudCollectorCache is used to cache Prometheus metrics
// The main goal of this cache is to avoid to overload the Metrics API
type CCloudCollectorCache struct {
	cachedValue  []prometheus.Metric
	cachedTime   time.Time
	cachedSecond int
}

// MaybeSendToChan populate the chan if required
// return true if it populates, otherwise false
func (ccc *CCloudCollectorCache) MaybeSendToChan(ch chan<- prometheus.Metric) bool {
	now := time.Now()
	if ccc.cachedTime.Add(time.Second*10).After(now) && len(ccc.cachedValue) > 0 {
		log.Trace("Returning cached values")
		ccc.SendToChan(ch)
		return true
	}
	return false
}

// Hijack all data from the chanel into the cache and forward them to another chan
func (ccc *CCloudCollectorCache) Hijack(ch chan prometheus.Metric, origCh chan<- prometheus.Metric, wg *sync.WaitGroup) {
	log.Trace("Populating cache")
	ccc.cachedValue = []prometheus.Metric{}
	ccc.cachedTime = time.Now()
	wg.Add(1)
	go func(ch chan prometheus.Metric, origCh chan<- prometheus.Metric) {
		for metric := range ch {
			ccc.cachedValue = append(ccc.cachedValue, metric)
			origCh <- metric
		}
		wg.Done()
	}(ch, origCh)
}

// SendToChan all cached data
func (ccc *CCloudCollectorCache) SendToChan(ch chan<- prometheus.Metric) {
	for _, metric := range ccc.cachedValue {
		ch <- metric
	}
}

// NewCache returns a newly created cache
func NewCache(duration int) CCloudCollectorCache {
	ccc := CCloudCollectorCache{}
	ccc.cachedValue = []prometheus.Metric{}
	ccc.cachedTime = time.UnixMilli(0)
	ccc.cachedSecond = duration
	return ccc
}
