package crawler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"time"

	"github.com/SimonRichardson/crwlr/pkg/document"
	"github.com/SimonRichardson/crwlr/pkg/peer"
	"github.com/SimonRichardson/crwlr/pkg/report"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
)

const defaultRobotsTxt = "/robots.txt"

var defaultRobotsURL, _ = url.Parse(defaultRobotsTxt)

// Crawler enables the crawling of a specific domain.
type Crawler struct {
	client           *http.Client
	agent            *peer.UserAgent
	filters          []Filter
	stack            chan *url.URL
	stop             chan chan struct{}
	peers            sync.Pool
	cache            *Cache
	robotsRequest    bool
	robotsCrawlDelay bool
	done             bool
	gauge            *Gauge
	logger           log.Logger
}

// NewCrawler creates a Crawler from a http.Client
func NewCrawler(client *http.Client, agent *peer.UserAgent, robotsRequest, robotsCrawlDelay bool, logger log.Logger) *Crawler {
	return &Crawler{
		client:  client,
		agent:   agent,
		stack:   make(chan *url.URL),
		stop:    make(chan chan struct{}),
		filters: []Filter{},
		peers: sync.Pool{
			New: func() interface{} {
				return peer.NewAgent(client, agent, logger)
			},
		},
		cache:            NewCache(log.With(logger, "component", "cache")),
		robotsRequest:    robotsRequest,
		robotsCrawlDelay: robotsCrawlDelay,
		gauge:            NewGauge(),
		logger:           logger,
	}
}

// Filter defines a way to add a filter to a series of filters to define if
// a url should be executed when crawling.
// Note: This is only additive filter, removing of filters is not supported.
func (c *Crawler) Filter(f Filter) {
	c.filters = append(c.filters, f)
}

// Run executes the list of urls on the crawler stack
func (c *Crawler) Run(u *url.URL) error {
	// Increment the gauge when we add a url.
	c.gauge.Increment()
	go func() { c.stack <- u }()

loop:
	for {
		select {
		case v, ok := <-c.stack:
			if !ok {
				return errors.Errorf("unexpected stack url")
			}

			// Check to see if we're done or not.
			if c.gauge.Value() < 1 {
				c.done = true
				break loop
			}

			if !c.filtered(v) {
				continue
			}

			// Check to see if we need to request robots.txt
			if c.robotsRequest {
				// Get the robots for a giving host
				group := c.getRobotsGroup(v)
				if group.Test(v.Path) {
					if c.robotsCrawlDelay && group.CrawlDelay > 0 {
						time.Sleep(group.CrawlDelay)
					}
				} else {
					// If the path is not allowed in the robots group, cache
					// the path, so it will be bypassed if requested again.
					go c.gauge.Decrement()
					c.assignFilterMetric(v)
				}
			}

			// We can run this in a concurrent way now, as we've taken into
			// consideration the crawl delay.
			go c.fetch(v)

		case q := <-c.stop:
			close(q)
			break loop
		}
	}

	return nil
}

// Close terminates any workers currently executing.
func (c *Crawler) Close() {
	if c.done {
		return
	}

	q := make(chan struct{})
	c.stop <- q
	<-q
}

// MetricsReport returns the report of all the metric things that have been
// cached, missed or errorred.
func (c *Crawler) MetricsReport(duration time.Duration) *report.MetricReport {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	// Take a snapshot of the cache metrics
	m := map[string]*report.Row{}
	for k, v := range c.cache.metrics {
		m[k] = &report.Row{
			Requested: int(v.Requested.Time()),
			Received:  int(v.Received.Time()),
			Filtered:  int(v.Filtered.Time()),
			Errorred:  int(v.Errorred.Time()),
			Duration:  v.Duration,
		}
	}

	return report.NewMetricReport(duration, m)
}

// SiteReport returns the report of all the sites pages whist crawling
func (c *Crawler) SiteReport() *report.SiteReport {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	p := map[string]*report.Page{}
	for k, v := range c.cache.metrics {
		p[k] = &report.Page{
			Links:  v.RefLinks,
			Assets: v.RefAssetLinks,
		}
	}

	return report.NewSiteReport(p)
}

func (c *Crawler) filtered(u *url.URL) bool {
	for _, v := range c.filters {
		if !v.Valid(u) {
			return false
		}
	}
	return true
}

func (c *Crawler) fetch(u *url.URL) {
	str := u.String()
	// We don't need to match existing ones
	metric, err := c.cache.Get(str)
	if err == nil {
		metric.Requested.Increment()
		go c.gauge.Decrement()
		return
	}
	// Make sure we do have a metric
	if metric == nil {
		metric = NewMetric()
		c.cache.Set(str, metric)
	}

	began := time.Now()
	metric.Requested.Increment()

	level.Debug(c.logger).Log("url", str)

	body, err := c.request(u, peer.Host, checkResponseStatus)
	if err != nil {
		metric.Errorred.Increment()
		return
	}

	links, err := c.collect(body, u)
	if err != nil {
		metric.Errorred.Increment()
		return
	}

	metric.Received.Increment()
	metric.Duration = time.Since(began)

	for _, u := range links {
		// Preemptively remove any links that we know are invalid
		// or essentially a no-op.
		if !c.filtered(u) {
			continue
		}
		if c.cache.Exists(u.String()) {
			c.assignFilterMetric(u)
			continue
		}

		metric.AppendRefLink(u.String())

		// Increment the gauge
		c.gauge.Increment()
		go func(u *url.URL) { c.stack <- u }(u)
	}

	go c.gauge.Decrement()
}

func (c *Crawler) getRobotsGroup(u *url.URL) *robotstxt.Group {
	// Get the robot.txt from the domain.
	robotsURL := u.ResolveReference(defaultRobotsURL)
	metric, err := c.cache.Get(robotsURL.String())
	if err != nil {
		metric = c.requestRobots(robotsURL)
	}

	// Empty group
	group := &robotstxt.Group{}
	if metric.Robots != nil {
		group = metric.Robots.FindGroup(c.agent.Robot)
	}
	return group
}

// Request a document and read it's response body
func (c *Crawler) request(u *url.URL, agentType peer.AgentType, fn func(*http.Response) error) (body []byte, err error) {
	agent := c.peers.Get().(*peer.Agent)
	defer c.peers.Put(agent)

	var resp *http.Response
	if resp, err = agent.Request(peer.NewAgentContext(u), agentType); err != nil {
		return
	}

	if err = fn(resp); err != nil {
		return
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	if err = resp.Body.Close(); err != nil {
		return
	}

	return
}

func checkResponseStatus(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.Errorf("bad status code: %s", resp.Status)
	}
	return nil
}

func checkRobotsResponseStatus(resp *http.Response) error {
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return errors.Errorf("bad status code: %s", resp.Status)
	}
	return nil
}

func (c *Crawler) requestRobots(u *url.URL) *Metric {
	var (
		err        error
		body       []byte
		statusCode int

		began  = time.Now()
		metric = NewMetric()
	)

	body, err = c.request(u, peer.Robot, func(resp *http.Response) error {
		statusCode = resp.StatusCode
		return checkRobotsResponseStatus(resp)
	})

	if err == nil {
		metric.Received.Increment()

		var robots *robotstxt.RobotsData
		if robots, err = robotstxt.FromStatusAndBytes(statusCode, body); err == nil {
			metric.Robots = robots
		}
	}

	if err != nil {
		metric.Errorred.Increment()
	}
	metric.Requested.Increment()
	metric.Duration = time.Since(began)

	c.cache.Set(u.String(), metric)

	return metric
}

// Collect all the links with in a document
func (c *Crawler) collect(body []byte, u *url.URL) (links []*url.URL, err error) {
	var node *html.Node
	if node, err = html.Parse(bytes.NewBuffer(body)); err != nil {
		return
	}

	doc := document.NewDocument(u, node, log.With(c.logger, "component", "document"))
	err = doc.WalkLinks(func(url *url.URL) error {
		// Ignore bad links, as we want the system to be resilient
		if err != nil {
			level.Error(c.logger).Log("err", err, "url", url)
			return nil
		}

		links = append(links, url)
		return nil
	})
	return
}

// Assign the filtered mertic to the disallowed
func (c *Crawler) assignFilterMetric(u *url.URL) {
	s := u.String()
	metric, err := c.cache.Get(s)
	if err != nil {
		metric = NewMetric()
		c.cache.Set(s, metric)
	}
	metric.Filtered.Increment()
}

// Gauge defines a value that can go both up and down in a safe way.
type Gauge struct {
	value int64
}

// NewGauge creates a Gauge
func NewGauge() *Gauge {
	return &Gauge{0}
}

// Increment a Gauge value
func (c *Gauge) Increment() {
	atomic.AddInt64(&c.value, 1)
}

// Decrement a Gauge value
func (c *Gauge) Decrement() {
	atomic.AddInt64(&c.value, -1)
}

// Value returns how much movement the Gauge has changed.
func (c *Gauge) Value() int64 {
	return atomic.LoadInt64(&c.value)
}
