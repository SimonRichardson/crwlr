package main

import (
	"errors"
	"flag"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// runStatic creates host to walk.
func runStatic(args []string) error {
	// flags for the static command
	var (
		flagset = flag.NewFlagSet("static", flag.ExitOnError)

		debug = flagset.Bool("debug", false, "debug logging")
		addr  = flagset.String("addr", defaultAddr, "addr to use for static creation")
	)
	flagset.Usage = usageFor(flagset, "static [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}
	if flagset.NFlag() == 0 {
		// Nothing found.
		return errorFor(flagset, "static [flags]", errors.New("specify at least argument"))
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

	return nil
}
