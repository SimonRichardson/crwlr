package crawler

import (
	"sync"
	"sync/atomic"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type Cache struct {
	mutex   sync.Mutex
	metrics map[string]*Metric
	logger  log.Logger
}

func NewCache(logger log.Logger) *Cache {
	return &Cache{
		mutex:   sync.Mutex{},
		metrics: map[string]*Metric{},
		logger:  logger,
	}
}

func (c *Cache) Exists(v string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.metrics[v]
	return ok
}

func (c *Cache) Get(v string) (*Metric, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	m, ok := c.metrics[v]
	if !ok {
		return nil, errors.New("not found")
	}
	return m, nil
}

type Metric struct {
	Requested, Received Clock
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
