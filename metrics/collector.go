package metrics

import (
	"sync"
	"time"
)

type Collector struct {
	mu              sync.RWMutex
	startTime       time.Time
	requestsTotal   int64
	requestsSuccess int64
	requestsFailed  int64
	totalDuration   time.Duration
	browsersUsed    map[string]int64
}

func NewCollector() *Collector {
	return &Collector{
		startTime:    time.Now(),
		browsersUsed: make(map[string]int64),
	}
}

func (c *Collector) RecordRequest(browser string, success bool, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestsTotal++
	if success {
		c.requestsSuccess++
	} else {
		c.requestsFailed++
	}
	c.totalDuration += duration
	c.browsersUsed[browser]++
}

func (c *Collector) GetMetrics() (uptime int64, total, success, failed int64, avgDuration float64, browsers map[string]int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	uptime = int64(time.Since(c.startTime).Seconds())
	total = c.requestsTotal
	success = c.requestsSuccess
	failed = c.requestsFailed

	if c.requestsTotal > 0 {
		avgDuration = float64(c.totalDuration.Milliseconds()) / float64(c.requestsTotal)
	}

	browsers = make(map[string]int64)
	for k, v := range c.browsersUsed {
		browsers[k] = v
	}

	return
}
