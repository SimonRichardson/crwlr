package crawler_test

import (
	"net/url"
	"testing"

	"github.com/SimonRichardson/crwlr/pkg/crawler"
)

var result bool

func BenchmarkAddrValid(b *testing.B) {
	var (
		u, _   = url.Parse("http://url.com")
		filter = crawler.Addr(u)
	)

	b.ResetTimer()

	var valid bool
	for i := 0; i < b.N; i++ {
		valid = filter.Valid(u)
	}

	result = valid
}

func BenchmarkAddrInvalid(b *testing.B) {
	var (
		x, _ = url.Parse("http://url.com")
		y, _ = url.Parse("http://lru.com")

		filter = crawler.Addr(x)
	)

	b.ResetTimer()

	var valid bool
	for i := 0; i < b.N; i++ {
		valid = filter.Valid(y)
	}

	result = valid
}
