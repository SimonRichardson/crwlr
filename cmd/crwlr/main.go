package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
)

var version = "dev"

const (
	defaultAddr string = "0.0.0.0:0"
)

func usage() {
	fmt.Fprintf(os.Stderr, "USAGE\n")
	fmt.Fprintf(os.Stderr, "  %s <mode> [flags]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "MODES\n")
	fmt.Fprintf(os.Stderr, "  crawl      Crawling service\n")
	fmt.Fprintf(os.Stderr, "  static     Static template site for crawling\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "VERSION\n")
	fmt.Fprintf(os.Stderr, "  %s (%s)\n", version, runtime.Version())
	fmt.Fprintf(os.Stderr, "\n")
}

func usageFor(fs *flag.FlagSet, name string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", name)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")

		writer := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(writer, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		writer.Flush()

		fmt.Fprintf(os.Stderr, "\n")
	}
}

func errorFor(fs *flag.FlagSet, name string, err error) error {
	defer usageFor(fs, name)()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "ERROR\n")
		fmt.Fprintf(os.Stderr, "  %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "\n---------------------------------------------\n\n")

		// Surpress the original error.
		return errors.Errorf("")
	}

	return err
}

type command func([]string) error

func (c command) Run(args []string) {
	if err := c(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var cmd command
	switch strings.ToLower(os.Args[1]) {
	case "static":
		cmd = runStatic
	case "crawl":
		cmd = runCrawl
	default:
		usage()
	}

	cmd.Run(os.Args[2:])
}
