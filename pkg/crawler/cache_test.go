package crawler

import (
	"testing"

	"testing/quick"

	"github.com/SimonRichardson/crwlr/pkg/test"
	"github.com/go-kit/kit/log"
)

func TestCacheExists(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			cache := NewCache(log.NewNopLogger())
			return !cache.Exists(a.String())
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("NonEmpty", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			cache := NewCache(log.NewNopLogger())
			cache.Set(a.String(), &Metric{})

			return cache.Exists(a.String())
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestCacheGet(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			var (
				cache  = NewCache(log.NewNopLogger())
				_, err = cache.Get(a.String())
			)
			return err != nil
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("NonEmpty", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			cache := NewCache(log.NewNopLogger())
			cache.Set(a.String(), &Metric{})

			_, err := cache.Get(a.String())
			return err == nil
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestCacheSet(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			var (
				cache  = NewCache(log.NewNopLogger())
				metric = &Metric{}
			)
			cache.Set(a.String(), metric)
			return cache.Exists(a.String())
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Metric", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			var (
				cache  = NewCache(log.NewNopLogger())
				metric = &Metric{
					Received: NewClock(),
				}
			)
			metric.Received.Increment()
			cache.Set(a.String(), metric)
			metric.Received.Increment()

			m, err := cache.Get(a.String())
			if err != nil {
				t.Error(err)
			}
			return m.Received.Time() == metric.Received.Time()
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestClock(t *testing.T) {
	t.Parallel()

	t.Run("Time", func(t *testing.T) {
		fn := func(a uint) bool {
			n := a % 10000
			clock := NewClock()
			for i := 0; i < int(n); i++ {
				clock.Increment()
			}
			return clock.Time() == int64(n)
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func BenchmarkCacheExistsEmpty(b *testing.B) {
	cache := NewCache(log.NewNopLogger())

	b.ResetTimer()

	var valid bool
	for i := 0; i < b.N; i++ {
		valid = cache.Exists("empty")
	}

	result = valid
}

func BenchmarkCacheExistsNonEmpty(b *testing.B) {
	cache := NewCache(log.NewNopLogger())
	cache.Set("nonempty", &Metric{})

	b.ResetTimer()

	var valid bool
	for i := 0; i < b.N; i++ {
		valid = cache.Exists("nonempty")
	}

	result = valid
}

func BenchmarkCacheGetEmpty(b *testing.B) {
	cache := NewCache(log.NewNopLogger())

	b.ResetTimer()

	var err error
	for i := 0; i < b.N; i++ {
		_, err = cache.Get("empty")
	}

	result = err == nil
}

func BenchmarkCacheGetNonEmpty(b *testing.B) {
	cache := NewCache(log.NewNopLogger())
	cache.Set("nonempty", &Metric{})

	b.ResetTimer()

	var err error
	for i := 0; i < b.N; i++ {
		_, err = cache.Get("nonempty")
	}

	result = err == nil
}

func BenchmarkClock(b *testing.B) {
	clock := NewClock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clock.Increment()
	}

	result = clock.Time() == int64(b.N)
}
