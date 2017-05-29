package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/SimonRichardson/crwlr/pkg/crawler"
	"github.com/SimonRichardson/crwlr/pkg/peer"
	"github.com/SimonRichardson/crwlr/pkg/static"
	"github.com/go-kit/kit/log"
)

func benchmarkCrawl(local bool, b *testing.B) {
	var (
		agent = peer.NewUserAgent(defaultUserAgent, defaultUserAgentRobot)
		c     = crawler.NewCrawler(http.DefaultClient, agent, true, false, log.NewNopLogger())

		server = httptest.NewServer(static.NewAPI(local, log.NewNopLogger()))

		u   *url.URL
		err error
	)

	if u, err = url.Parse(server.URL); err != nil {
		b.Fatal(err)
	}

	c.Filter(crawler.Addr(u))

	b.ResetTimer()

	wg := sync.WaitGroup{}
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		go func() {
			wg.Done()
			if err := c.Run(u); err != nil {
				b.Error(err)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkCrawl_NonLocal(b *testing.B) { benchmarkCrawl(false, b) }
