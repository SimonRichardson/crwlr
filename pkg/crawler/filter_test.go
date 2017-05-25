package crawler

import (
	"fmt"
	"net/url"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/crwlr/pkg/test"
)

func TestFilterAddr(t *testing.T) {
	t.Parallel()

	t.Run("Valid", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			var (
				h      = fmt.Sprintf("%s.com", a.String())
				u, err = url.Parse(fmt.Sprintf("http://%s", h))
			)
			if err != nil {
				t.Error(err)
			}
			return Addr(u).Valid(u)
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		fn := func(a, b test.ASCII) bool {
			var (
				h, err0 = url.Parse(fmt.Sprintf("http://%s.com", a.String()))
				u, err1 = url.Parse(fmt.Sprintf("http://%s.com", b.String()))
			)
			if err0 != nil || err1 != nil {
				t.Errorf("errors %v %v", err0.Error(), err1.Error())
			}
			return !Addr(h).Valid(u)
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Valid Port", func(t *testing.T) {
		fn := func(a test.ASCII, b uint) bool {
			var (
				h      = fmt.Sprintf("%s.com:%d", a.String(), b%10000)
				u, err = url.Parse(fmt.Sprintf("http://%s", h))
			)
			if err != nil {
				t.Error(err)
			}
			return Addr(u).Valid(u)
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

var result bool

func BenchmarkAddrValid(b *testing.B) {
	var (
		u, _   = url.Parse("http://url.com")
		filter = Addr(u)
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

		filter = Addr(x)
	)

	b.ResetTimer()

	var valid bool
	for i := 0; i < b.N; i++ {
		valid = filter.Valid(y)
	}

	result = valid
}
