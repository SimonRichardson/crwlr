package crawler

import (
	"fmt"
	"net/url"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/crwlr/pkg/test"
)

func TestFilterDomain(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			var (
				h      = fmt.Sprintf("%s.com", a.String())
				u, err = url.Parse(fmt.Sprintf("http://%s", h))
			)
			if err != nil {
				t.Error(err)
			}
			return Domain(h).Valid(u)
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		fn := func(a, b test.ASCII) bool {
			var (
				h      = fmt.Sprintf("%s.com", a.String())
				u, err = url.Parse(fmt.Sprintf("http://%s.com", b.String()))
			)
			if err != nil {
				t.Error(err)
			}
			return !Domain(h).Valid(u)
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
			return Domain(h).Valid(u)
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}
