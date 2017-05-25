package crawler

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"net/http/httptest"

	"net/url"

	"testing/quick"

	"github.com/SimonRichardson/crwlr/pkg/peer"
	"github.com/SimonRichardson/crwlr/pkg/static"
	"github.com/SimonRichardson/crwlr/pkg/test"
	"github.com/go-kit/kit/log"
)

func TestCrawl_Collect(t *testing.T) {
	t.Parallel()

	// Setup
	var (
		client = http.DefaultClient
		agent  = peer.NewUserAgent("", "")
		logger = log.NewNopLogger()
		u, _   = url.Parse("http://a.com")
	)

	t.Run("fetch cache", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			body := fmt.Sprintf(`<a href="/%s">%s</a>`, a.String(), a.String())

			c := NewCrawler(client, agent, false, false, logger)
			links, err := c.collect([]byte(body), u)
			if err != nil {
				t.Error(err)
				return false
			}

			if expected, actual := 1, len(links); expected != actual {
				t.Errorf("expected: %d, actual: %d", expected, actual)
				return false
			}

			if expected, actual := fmt.Sprintf("%s/%s", u.String(), a.String()), links[0].String(); expected != actual {
				t.Errorf("expected: %s, actual: %s", expected, actual)
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestCrawl_Request(t *testing.T) {
	t.Parallel()

	// Setup
	var (
		client = http.DefaultClient
		agent  = peer.NewUserAgent("", "")
		logger = log.NewNopLogger()
		server = httptest.NewServer(static.NewAPI(true, logger))
	)
	defer server.Close()

	// Make sure we've got a valid url
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("fetch cache", func(t *testing.T) {
		c := NewCrawler(client, agent, false, false, logger)
		c.fetch(u)

		if !c.cache.Exists(u.String()) {
			t.Error("expected to have populated the cache")
		}
	})

	t.Run("fetch metric", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			path, err := url.Parse(fmt.Sprintf("/%s", a.String()))
			if err != nil {
				t.Error(err)
				return false
			}

			base := u.ResolveReference(path)

			c := NewCrawler(client, agent, false, false, logger)
			c.fetch(base)

			if m, err := c.cache.Get(base.String()); err == nil {
				if actual := m.Requested.Time(); actual != 1 {
					t.Errorf("requested: expected: 1, actual: %d", actual)
				}

				return true
			}

			t.Error("expected to have a set of metrics")
			return false
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestCrawl_RequestRobots(t *testing.T) {
	t.Parallel()

	// Setup
	var (
		client = http.DefaultClient
		agent  = peer.NewUserAgent("", "")
		logger = log.NewNopLogger()
		server = httptest.NewServer(static.NewAPI(true, logger))
	)
	defer server.Close()

	// Make sure we've got a valid url
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("fetch cache", func(t *testing.T) {
		c := NewCrawler(client, agent, false, false, logger)
		c.requestRobots(u)

		if !c.cache.Exists(u.String()) {
			t.Error("expected to have populated the cache")
		}
	})

	t.Run("fetch metric", func(t *testing.T) {
		c := NewCrawler(client, agent, false, false, logger)
		c.requestRobots(u)

		if m, err := c.cache.Get(u.String()); err == nil {
			if actual := m.Requested.Time(); actual != 1 {
				t.Errorf("requested: expected: 1, actual: %d", actual)
			}
			if actual := m.Received.Time(); actual != 1 {
				t.Errorf("received: expected: 1, actual: %d", actual)
			}
			if actual := m.Duration.Nanoseconds(); actual == 0 {
				t.Errorf("duration: expected: > 0, actual: %d", actual)
			}

			return
		}

		t.Error("expected to have a set of metrics")
	})
}

func TestCrawl_GetRobotsGroup(t *testing.T) {
	t.Parallel()

	// Setup
	var (
		client = http.DefaultClient
		agent  = peer.NewUserAgent("", "")
		logger = log.NewNopLogger()
		server = httptest.NewServer(static.NewAPI(true, logger))
	)
	defer server.Close()

	// Make sure we've got a valid url
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("get", func(t *testing.T) {
		c := NewCrawler(client, agent, false, false, logger)
		group := c.getRobotsGroup(u)

		if group == nil {
			t.Error("expected to have a group")
		}
	})
}

func TestCrawl_AssignFilterMetric(t *testing.T) {
	t.Parallel()

	// Setup
	var (
		client = http.DefaultClient
		agent  = peer.NewUserAgent("", "")
		logger = log.NewNopLogger()
		u, _   = url.Parse("http://a.com")
	)

	t.Run("increment", func(t *testing.T) {
		fn := func(a test.ASCII, b uint) bool {
			path, err := url.Parse(fmt.Sprintf("/%s", a.String()))
			if err != nil {
				t.Error(err)
				return false
			}

			base := u.ResolveReference(path)

			c := NewCrawler(client, agent, false, false, logger)

			amount := int(b % 1000)
			for i := 0; i < amount; i++ {
				c.assignFilterMetric(base)
			}

			metric, err := c.cache.Get(base.String())
			if err != nil {
				t.Error(err)
				return false
			}

			if expected, actual := amount, int(metric.Filtered.Time()); expected != actual {
				t.Errorf("expected: %d, actual: %d", expected, actual)
				return false
			}
			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestGauge(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		gauge := NewGauge()

		for i := 0; i < 100; i++ {
			gauge.Increment()
		}

		for i := 0; i < 100; i++ {
			gauge.Decrement()
		}

		if val := gauge.Value(); val != 0 {
			t.Errorf("expected: 100, actual: %d", val)
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		gauge := NewGauge()

		wg := sync.WaitGroup{}
		wg.Add(200)
		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				gauge.Increment()
			}()
			go func() {
				defer wg.Done()
				gauge.Decrement()
			}()
		}

		wg.Wait()

		if val := gauge.Value(); val != 0 {
			t.Errorf("expected: 100, actual: %d", val)
		}
	})
}
