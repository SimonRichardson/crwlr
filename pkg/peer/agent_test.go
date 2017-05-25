package peer

import (
	"testing"
	"testing/quick"

	"net/http"
	"net/http/httptest"

	"net/url"

	"github.com/SimonRichardson/crwlr/pkg/test"
	"github.com/go-kit/kit/log"
)

func TestUserAgent(t *testing.T) {
	t.Parallel()

	t.Run("type", func(t *testing.T) {
		fn := func(a, b test.ASCII) bool {
			host, robot := a.String(), b.String()
			agent := NewUserAgent(host, robot)

			if expected, actual := host, agent.Type(Host); expected != actual {
				t.Errorf("expected: %s, actual: %s", expected, actual)
				return false
			}

			if expected, actual := robot, agent.Type(Robot); expected != actual {
				t.Errorf("expected: %s, actual: %s", expected, actual)
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestRequest(t *testing.T) {
	t.Parallel()

	var header string

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("User-Agent")
	})

	var (
		server    = httptest.NewServer(mux)
		client    = http.DefaultClient
		userAgent = NewUserAgent("host", "robot")
	)
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	agent := NewAgent(client, userAgent, log.NewNopLogger())
	if _, err := agent.Request(NewAgentContext(u), Host); err != nil {
		t.Fatal(err)
	}

	if expected, actual := "host", header; expected != actual {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}

	if _, err := agent.Request(NewAgentContext(u), Robot); err != nil {
		t.Fatal(err)
	}

	if expected, actual := "robot", header; expected != actual {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}
