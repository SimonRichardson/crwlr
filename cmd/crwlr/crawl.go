package main

import (
	"bufio"
	"flag"
	"net"
	"net/url"
	"os"
	"text/tabwriter"

	"strings"

	"net/http"
	"time"

	"fmt"

	"github.com/SimonRichardson/crwlr/pkg/crawler"
	"github.com/SimonRichardson/crwlr/pkg/group"
	"github.com/SimonRichardson/crwlr/pkg/peer"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

const (
	defaultFollowRedirects  = true
	defaultFilterSameDomain = true
	defaultRobotsRequest    = true
	defaultRobotsCrawlDelay = false
	defaultReportSitemap    = true
	defaultReportMetrics    = false

	defaultUserAgent      = "Mozilla/5.0 (compatible; crwlr/0.1; +http://crwlr.com)"
	defaultUserAgentRobot = "Googlebot (crwlr/0.1)"
)

// runCrawl crawls a specific addr.
func runCrawl(args []string) error {
	// flags for the crawl command
	var (
		flagset = flag.NewFlagSet("crawl", flag.ExitOnError)

		debug            = flagset.Bool("debug", false, "debug logging")
		addr             = flagset.String("addr", defaultAddr, "addr to start crawling")
		reportSitemap    = flagset.Bool("report.sitemap", defaultReportSitemap, "report the sitemap of the crawl")
		reportMetrics    = flagset.Bool("report.metrics", defaultReportMetrics, "report the metric outcomes of the crawl")
		followRedirects  = flagset.Bool("follow-redirects", defaultFollowRedirects, "should the crawler follow redirects")
		userAgent        = flagset.String("useragent.full", defaultUserAgent, "full user agent the crawler should use")
		userAgentRobot   = flagset.String("useragent.robot", defaultUserAgentRobot, "robot user agent the crawler should use")
		filterSameDomain = flagset.Bool("filter.same-domain", defaultFilterSameDomain, "filter other domains that aren't the same")
		robotsRequest    = flagset.Bool("robots.request", defaultRobotsRequest, "request the robots.txt when crawling")
		robotsCrawlDelay = flagset.Bool("robots.crawl-delay", defaultRobotsCrawlDelay, "use the robots.txt crawl delay when crawling")
	)
	flagset.Usage = usageFor(flagset, "crawl [flags]")

	// Crawl can be used in a pipe constructor
	// example: `crwlr static -output.addr=true | crwlr crawl`
	if info, err := os.Stdin.Stat(); err == nil && (info.Mode()&os.ModeCharDevice) != os.ModeCharDevice {
		// Read from the stdin for any possible pipe arguments.
		var (
			reader    = bufio.NewReader(os.Stdin)
			line, err = reader.ReadString('\n')
		)
		if err != nil || (len(args) == 0 && len(line) == 0) {
			return errorFor(flagset, "crawl [flags]", errors.New("specify addr for crawing via pipe"))
		}

		parts := strings.Split(strings.TrimSpace(line), " ")
		args = append(args, parts...)
	}

	if err := flagset.Parse(args); err != nil {
		return err
	}
	if flagset.NFlag() == 0 {
		// Nothing found.
		return errorFor(flagset, "crawl [flags]", errors.New("specify at least argument"))
	}

	// Setup the logger.
	var logger log.Logger
	{
		logLevel := level.AllowInfo()
		if *debug {
			logLevel = level.AllowAll()
		}
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = level.NewFilter(logger, logLevel)
	}

	level.Debug(logger).Log("addr", *addr)

	// Parse the addr URL
	u, err := url.Parse(*addr)
	if err != nil {
		return errorFor(flagset, "crawl [flags]", errors.Wrap(err, "expected valid domain"))
	}

	// Create the HTTP client that the crawler will use.
	timeoutClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			ResponseHeaderTimeout: 5 * time.Second,
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 1,
		},
	}

	// This allows us to prevent redirects on certain domains.
	if !*followRedirects {
		timeoutClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Execution group.
	var g group.Group
	{
		cancel := make(chan struct{})
		g.Add(func() error {
			<-cancel
			return nil
		}, func(error) {
			close(cancel)
		})
	}
	{
		// Go consume the domain.
		var (
			agent = peer.NewUserAgent(*userAgent, *userAgentRobot)
			c     = crawler.NewCrawler(timeoutClient, agent, *robotsRequest, *robotsCrawlDelay, logger)
			began = time.Now()
		)

		// Filter only on the same domain i.e. don't crawl the internet.
		if *filterSameDomain {
			c.Filter(crawler.Addr(u))
		}

		g.Add(func() error {
			return c.Run(u)
		}, func(error) {
			if *reportSitemap {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
				c.SiteReport().Write(w)
				w.Flush()
				if *reportMetrics {
					fmt.Fprintln(os.Stdout, "")
				}
			}
			if *reportMetrics {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
				c.MetricsReport(time.Since(began)).Write(w)
				w.Flush()
			}

			c.Close()
		})
	}
	{
		// Setup os signal interruptions.
		cancel := make(chan struct{})
		g.Add(func() error {
			return interrupt(cancel)
		}, func(error) {
			close(cancel)
		})
	}

	return g.Run()
}
