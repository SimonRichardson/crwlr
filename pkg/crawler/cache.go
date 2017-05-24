package crawler

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/temoto/robotstxt"
)

// Cache returns a metric cache of urls visited along with misses, errors and
// duration of the request.
type Cache struct {
	mutex   sync.RWMutex
	metrics map[string]*Metric
	logger  log.Logger
}

// NewCache returns a cache
func NewCache(logger log.Logger) *Cache {
	return &Cache{
		mutex:   sync.RWMutex{},
		metrics: map[string]*Metric{},
		logger:  logger,
	}
}

// Exists returns truthy if the value exists
func (c *Cache) Exists(v string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, ok := c.metrics[v]
	return ok
}

// Get a cache metric value based on the value.
func (c *Cache) Get(v string) (*Metric, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	m, ok := c.metrics[v]
	if !ok {
		return nil, errors.New("not found")
	}
	return m, nil
}

// Set a cache metric based on the value.
// Note: if the metric already exists it will over write it.
func (c *Cache) Set(v string, m *Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics[v] = m
}

// Metric holds some very simple primitive metric values for reporting.
type Metric struct {
	Requested, Received *Clock
	Filtered, Errorred  *Clock
	Duration            time.Duration
	Robots              *robotstxt.RobotsData
}

// NewMetric creates a new Metric
func NewMetric() *Metric {
	return &Metric{
		Requested: NewClock(),
		Received:  NewClock(),
		Filtered:  NewClock(),
		Errorred:  NewClock(),
		Duration:  0,
	}
}

// WithRobots returns a new Metric with the associated robots data
func (m *Metric) WithRobots(r *robotstxt.RobotsData) *Metric {
	return &Metric{
		Robots: r,
	}
}

// Clock defines a metric for monitoring how many times something occurred.
type Clock struct {
	times int64
}

// NewClock creates a Clock
func NewClock() *Clock {
	return &Clock{0}
}

// Increment a clock timing
func (c *Clock) Increment() {
	atomic.AddInt64(&c.times, 1)
}

// Time returns how much movement the clock has changed.
func (c *Clock) Time() int64 {
	return atomic.LoadInt64(&c.times)
}
