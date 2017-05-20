package main

import (
	"flag"
	"net/url"
	"os"

	"github.com/SimonRichardson/crwlr/pkg/crawler"
	"github.com/SimonRichardson/crwlr/pkg/group"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

// runCrawl crawls a specific addr.
func runCrawl(args []string) error {
	// flags for the crawl command
	var (
		flagset = flag.NewFlagSet("crawl", flag.ExitOnError)

		debug = flagset.Bool("debug", false, "debug logging")
		addr  = flagset.String("addr", defaultAddr, "addr to start crawling")
	)
	flagset.Usage = usageFor(flagset, "crawl [flags]")
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
		c := crawler.NewCrawler(logger)
		c.Filter(crawler.Addr(u))

		g.Add(func() error {
			return c.Run(u)
		}, func(error) {
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
