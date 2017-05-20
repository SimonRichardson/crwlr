package crawler

import (
	"net/url"
	"strings"
)

// Filter describes a way of filtering what urls should and should not be
// crawled.
type Filter interface {
	// Valid takes a url that should be filtered when crawling.
	Valid(*url.URL) bool
}

type domain struct {
	d string
}

func (d domain) Valid(u *url.URL) bool {
	return strings.HasPrefix(d.d, u.Host)
}

// Domain returns a Filter that filters out urls that should be crawled by
// the domain name of a url host (accepts host:port)
func Domain(d string) Filter {
	return domain{d}
}

type fn struct {
	f func(*url.URL) bool
}

func (f fn) Valid(u *url.URL) bool {
	return f.f(u)
}

// Func returns a Filter that filters out urls that should be crawled by a func
func Func(f func(*url.URL) bool) Filter {
	return fn{f}
}
