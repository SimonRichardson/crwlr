package crawler_test

import (
	"testing"

	"github.com/SimonRichardson/crwlr/pkg/crawler"
	"github.com/go-kit/kit/log"
)

func BenchmarkCacheExistsEmpty(b *testing.B) {
	cache := crawler.NewCache(log.NewNopLogger())

	b.ResetTimer()

	var valid bool
	for i := 0; i < b.N; i++ {
		valid = cache.Exists("empty")
	}

	result = valid
}

func BenchmarkCacheExistsNonEmpty(b *testing.B) {
	cache := crawler.NewCache(log.NewNopLogger())
	cache.Set("nonempty", &crawler.Metric{})

	b.ResetTimer()

	var valid bool
	for i := 0; i < b.N; i++ {
		valid = cache.Exists("nonempty")
	}

	result = valid
}

func BenchmarkCacheGetEmpty(b *testing.B) {
	cache := crawler.NewCache(log.NewNopLogger())

	b.ResetTimer()

	var err error
	for i := 0; i < b.N; i++ {
		_, err = cache.Get("empty")
	}

	result = err == nil
}

func BenchmarkCacheGetNonEmpty(b *testing.B) {
	cache := crawler.NewCache(log.NewNopLogger())
	cache.Set("nonempty", &crawler.Metric{})

	b.ResetTimer()

	var err error
	for i := 0; i < b.N; i++ {
		_, err = cache.Get("nonempty")
	}

	result = err == nil
}

func BenchmarkClock(b *testing.B) {
	clock := crawler.NewClock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clock.Increment()
	}

	result = clock.Time() == int64(b.N)
}
