package main

import (
	"errors"
	"flag"
	"os"

	"github.com/SimonRichardson/crwlr/pkg/crawler"
	"github.com/SimonRichardson/crwlr/pkg/group"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	defaultDomain string = "0.0.0.0:0"
)

// runCrawl crawls a specific domain.
func runCrawl(args []string) error {
	// flags for the crawl command
	var (
		flagset = flag.NewFlagSet("crawl", flag.ExitOnError)

		debug  = flagset.Bool("debug", false, "debug logging")
		domain = flagset.String("domain", defaultDomain, "domain to start crawling")
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

	level.Debug(logger).Log("domain", *domain)

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
		c.Filter(crawler.Domain(*domain))

		g.Add(func() error {
			return c.Run(*domain)
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
