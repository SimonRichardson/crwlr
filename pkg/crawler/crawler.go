package crawler

import "github.com/go-kit/kit/log"
import "net/url"

type Crawler struct {
	filters []Filter
	logger  log.Logger
}

func NewCrawler(logger log.Logger) *Crawler {
	return &Crawler{
		logger: logger,
	}
}

// Filter defines a way to add a filter to a series of filters to define if
// a url should be executed when crawling.
// Note: This is only additive filter, removing of filters is not supported.
func (c *Crawler) Filter(f Filter) {
	c.filters = append(c.filters, f)
}

// TODO : Run(string) should be Run(*url.URL)
func (c *Crawler) Run(u *url.URL) error {
	return nil
}

func (c *Crawler) Close() {

}
