package crawler

import "net/url"

// Filter describes a way of filtering what urls should and should not be
// crawled.
type Filter interface {
	// Valid takes a url that should be filtered when crawling.
	Valid(*url.URL) bool
}

type addr struct {
	u *url.URL
}

func (a addr) Valid(u *url.URL) bool {
	return a.u.Host == u.Host
}

// Addr returns a Filter that filters out urls that should be crawled by
// the addr name of a url host (accepts host:port)
func Addr(u *url.URL) Filter {
	return addr{u}
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
