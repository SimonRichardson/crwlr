package crawler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/SimonRichardson/crwlr/pkg/document"
	"github.com/SimonRichardson/crwlr/pkg/peer"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// Crawler enables the crawling of a specific domain.
type Crawler struct {
	client  *http.Client
	filters []Filter
	stack   chan *url.URL
	stop    chan chan struct{}
	peers   sync.Pool
	mutex   sync.Mutex
	metrics map[string]Metric
	logger  log.Logger
}

// NewCrawler creates a Crawler from a http.Client
func NewCrawler(client *http.Client, agent *peer.UserAgent, logger log.Logger) *Crawler {
	return &Crawler{
		client:  client,
		stack:   make(chan *url.URL),
		filters: []Filter{},
		peers: sync.Pool{
			New: func() interface{} {
				return peer.NewAgent(client, agent, logger)
			},
		},
		mutex:   sync.Mutex{},
		metrics: map[string]Metric{},
		logger:  logger,
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
	// Put the first url on the stack. This will be the starting position
	go func() { c.stack <- u }()

	var errs []string
	for {
		select {
		case pop := <-c.stack:
			// Make sure it's not already got a metric, if it hasn't make one!
			if pop == nil || c.metric(pop) {
				continue
			}
			c.newMetric(pop)

			go func(u *url.URL) {
				// Step 1. Request the url.
				agent := c.peers.Get().(*peer.Agent)
				defer c.peers.Put(agent)

				body, err := request(agent, pop)
				if err != nil {
					return
				}

				// Step 2. Parse the html creating a document to extract links
				links, err := collect(body, pop, c.logger)
				if err != nil {
					return
				}

				// Step 3. Look through and get the links as they come!
				for _, link := range links {
					if c.filtered(link) && !c.metric(link) {
						fmt.Println(link.String())
						c.stack <- link
					}
				}

				if len(links) == 0 && len(c.stack) == 0 {
					close(c.stack)
				}
			}(pop)
		}
	}

	fmt.Println("feck")

	if len(errs) > 0 {
		return errors.Errorf("crawler errors: %s", strings.Join(errs, ", "))
	}

	return nil
}

// Close terminates any workers currently executing.
func (c *Crawler) Close() {
	q := make(chan struct{})
	c.stop <- q
	<-q
}

func (c *Crawler) metric(u *url.URL) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.metrics[u.String()]; ok {
		return true
	}
	return false
}

func (c *Crawler) newMetric(u *url.URL) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics[u.String()] = Metric{}
}

func (c *Crawler) filtered(u *url.URL) bool {
	for _, v := range c.filters {
		if !v.Valid(u) {
			return false
		}
	}
	return true
}

func request(agent *peer.Agent, u *url.URL) (body []byte, err error) {
	var resp *http.Response
	if resp, err = agent.Request(peer.NewAgentContext(u)); err != nil {
		return
	}

	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		err = errors.New("bad status code")
		return
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	return
}

func collect(body []byte, u *url.URL, logger log.Logger) (links []*url.URL, err error) {
	var node *html.Node
	if node, err = html.Parse(bytes.NewBuffer(body)); err != nil {
		return
	}

	doc := document.NewDocument(u, node, log.With(logger, "component", "document"))
	err = doc.WalkLinks(func(url *url.URL, err error) error {
		if err != nil {
			return err
		}

		links = append(links, url)
		return nil
	})
	return
}

type Metric struct {
	Req      int
	Duration time.Time
}
