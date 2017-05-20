package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"net/http"

	"github.com/SimonRichardson/crwlr/pkg/group"
	"github.com/SimonRichardson/crwlr/pkg/static"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// runStatic creates host to walk.
func runStatic(args []string) error {
	// flags for the static command
	var (
		flagset = flag.NewFlagSet("static", flag.ExitOnError)

		debug   = flagset.Bool("debug", false, "debug logging")
		apiAddr = flagset.String("api", defaultAPIAddr, "listen address for static APIs")
		uiLocal = flagset.Bool("ui.local", true, "Use local files straight from the file system")
	)
	flagset.Usage = usageFor(flagset, "static [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
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

	apiNetwork, apiAddress, err := parseAddr(*apiAddr, defaultAPIPort)
	if err != nil {
		return err
	}

	apiListener, err := net.Listen(apiNetwork, apiAddress)
	if err != nil {
		return err
	}
	level.Info(logger).Log("API", fmt.Sprintf("%s://%s", apiNetwork, apiAddress))

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
		g.Add(func() error {
			mux := http.NewServeMux()
			mux.Handle("/", static.NewAPI(*uiLocal, logger))
			return http.Serve(apiListener, mux)
		}, func(error) {
			apiListener.Close()
		})
	}
	{
		cancel := make(chan struct{})
		g.Add(func() error {
			return interrupt(cancel)
		}, func(error) {
			close(cancel)
		})
	}
	return g.Run()
}
