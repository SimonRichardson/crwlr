# crwlr

## Command Crawler

 - [Getting started](#getting-started)
 - [Introduction](#introduction)
 - [Static](#static)
 - [Crawl](#crawl)
 - [Reports](#reports)
 - [Tests](#tests)
 - [Improvements](#improvements)

### Getting started

The crwlr command expects to have some things pre-installed via `go get` if you
would like to build the project.

 - go get github.com/Masterminds/glide
 - go get github.com/mjibson/esc

-----

Quick guide to getting started, this assumes you've got the `$GOPATH` setup
correctly and the gopath bin folder is in your `$PATH`:

```
glide install
make clean all
cd dist

./crwlr crawl -addr="http://tomblomfield.com"
```

### Introduction

The crwlr CLI is split up into two distinctive commands, `static` and `crawl`.
`static` command is only an aid to help manually test the `crawl` command along
with various benchmarking/integration tests.

### Static

The `static` command creates a series of pages that allow the `crawl` command to
walk, without hitting an external host. To help integration with `crawl`, the
`static` command can be used in combination with a pipe to send the current
address, this allows quick and fast iterative testing.

The following command launches the cli:

```
crwlr static
```

In combination with the crawl command, an extra argument is required.

```
crwlr static -output.addr=true | crwlr crawl
```

Also available is a quite descriptive `-help` section to better understand what
the static command can do:

```
crwlr static -help
USAGE
  static [flags]

FLAGS
  -api tcp://0.0.0.0:7650  listen address for static APIs
  -debug false             debug logging
  -output.addr false       Output address writes the address to stdout
  -output.prefix -addr=    Output prefix prefixes the flag to the output.addr
  -ui.local true           Use local files straight from the file system
```

### Crawl

The `crawl` command walks a host for potential new urls that it can also inturn
traverse. The command can configured (on by default) to check the `robots.txt`
of the host to follow the rules for crawling.

The command uses aggressive caching to help better improve performance and to
be more efficient when crawling a host.

As part of the command it's also possible to output a report (on by default)
of what was crawled and expose some metrics about what went on. These include,
metrics like: requested vs received or filtered and errors.

 - Requested is when a request is sent to the host, it's not know if that request
 was actually successful.
 - Received is the acknowledgement of the request succeeding.
 - Filtered describes if the host was cached already.
 - Errorred states if the request failed for some reason.

The following command launches the cli:

```
crwlr crawl -addr="http://yourhosthere.com"
```

Also available is a comprehensive `-help` section:

```
crwlr crawl -help
USAGE
  crawl [flags]

FLAGS
  -addr 0.0.0.0:0                                                         addr to start crawling
  -debug false                                                            debug logging
  -filter.same-domain true                                                filter other domains that aren't the same
  -follow-redirects true                                                  should the crawler follow redirects
  -report.metrics false                                                   report the metric outcomes of the crawl
  -report.sitemap true                                                    report the sitemap of the crawl
  -robots.crawl-delay false                                               use the robots.txt crawl delay when crawling
  -robots.request true                                                    request the robots.txt when crawling
  -useragent.full Mozilla/5.0 (compatible; crwlr/0.1; +http://crwlr.com)  full user agent the crawler should use
  -useragent.robot Googlebot (crwlr/0.1)                                  robot user agent the crawler should use

```

### Reports

The reporting part of the command outputs two different types of information;
sitemap reporting and metric reporting. Both reports can be turned off behind
a series of flags.

#### Sitemap Reports

When the command is done the sitemap report can be outputted (on by default),
which explains what was linked to what and also includes a list of static assets
that was also linked in the file.

A possible output is as follows:

```
dist/crwlr crawl
 URL                              | Ref Links                   | Ref Assets                        |
 http://0.0.0.0:7650/robots.txt   |                             |                                   |
 http://0.0.0.0:7650              |                             |                                   |
                                  | http://0.0.0.0:7650/index   | http://0.0.0.0:7650/index.css     |
                                  | http://0.0.0.0:7650/page1   | http://google.com/bootstrap.css   |
                                  | http://0.0.0.0:7650/bad     | http://0.0.0.0:7650/image.jpg     |
                                  |                             | http://google.com/image.jpg       |
 http://0.0.0.0:7650/index        |                             |                                   |
                                  |                             | http://0.0.0.0:7650/index.css     |
                                  |                             | http://google.com/bootstrap.css   |
                                  |                             | http://0.0.0.0:7650/image.jpg     |
                                  |                             | http://google.com/image.jpg       |
 http://0.0.0.0:7650/page1        |                             |                                   |
                                  | http://0.0.0.0:7650/page2   | http://0.0.0.0:7650/index1.css    |
                                  |                             | http://google.com/bootstrap.css   |
                                  |                             | http://0.0.0.0:7650/image2.jpg    |
                                  |                             | http://google.com/image.jpg       |
 http://0.0.0.0:7650/bad          |                             |                                   |
 http://0.0.0.0:7650/page2        |                             |                                   |
                                  | http://0.0.0.0:7650/page    |                                   |
                                  | http://0.0.0.0:7650/page3   |                                   |
 http://0.0.0.0:7650/page         |                             |                                   |
 http://0.0.0.0:7650/page3        |                             |                                   |
```

#### Metric Reports

When the command is done a report can be outputted (off by default), which can
help explain what the crawl actually requested vs what it filtered for example.

Example report using the `static` command is as follows:

```
dist/crwlr crawl -report.metrics=true
 URL                              | Avg Duration (ms)   | Requested   | Received   | Filtered   | Errorred   |
 http://0.0.0.0:7650/page         | 0                   | 1           | 0          | 0          | 1          |
 http://0.0.0.0:7650/page3        | 0                   | 1           | 0          | 1          | 0          |
 http://0.0.0.0:7650/robots.txt   | 5                   | 1           | 1          | 0          | 0          |
 http://0.0.0.0:7650              | 1                   | 1           | 1          | 0          | 0          |
 http://0.0.0.0:7650/index        | 0                   | 1           | 1          | 3          | 0          |
 http://0.0.0.0:7650/page1        | 1                   | 1           | 1          | 2          | 0          |
 http://0.0.0.0:7650/bad          | 0                   | 1           | 0          | 1          | 1          |
 http://0.0.0.0:7650/page2        | 0                   | 1           | 1          | 0          | 0          |

 Totals   | Duration (ms)   |
          | 9560            |
```

### Tests

Tests can be run using the following command, it also includes a series of
benchmarking tests:

```
 go test -v -bench=. $(glide nv)
```

### Improvements

Possible improvements:

 - Store the urls in a KVS so that a crawler can truly work distributed, esp. if
 the host is large or if it's allowed to crawl beyond the host.
 - Potentially better strategies to walk assets at a later date to back fill the
 metrics.
