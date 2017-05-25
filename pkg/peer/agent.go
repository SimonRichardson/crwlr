package peer

import (
	"context"
	"net/http"

	"net/url"
	"time"

	"github.com/go-kit/kit/log"
)

// AgentType represents the type of user agent to send
type AgentType int

const (
	// Host agent
	Host AgentType = iota
	// Robot agent
	Robot
)

// UserAgent contains the different user agent options when contacting a host.
type UserAgent struct {
	Host, Robot string
}

// NewUserAgent creates a UserAgent from the host and robot agent strings
func NewUserAgent(host, robot string) *UserAgent {
	return &UserAgent{
		Host:  host,
		Robot: robot,
	}
}

// Type returns the user agent, depending on the type
func (u *UserAgent) Type(a AgentType) string {
	switch a {
	case Robot:
		return u.Robot
	default:
		return u.Host
	}
}

// Agent wraps the http.Client to allow connections to peers.
type Agent struct {
	client    *http.Client
	userAgent *UserAgent
	logger    log.Logger
}

// NewAgent creates a new Agent
func NewAgent(client *http.Client, userAgent *UserAgent, logger log.Logger) *Agent {
	return &Agent{
		client:    client,
		userAgent: userAgent,
		logger:    logger,
	}
}

// Request a url using the client and the given peer.
func (a *Agent) Request(ctx *AgentContext, t AgentType) (*http.Response, error) {
	req, err := http.NewRequest("GET", ctx.URL.String(), nil)
	if err != nil {
		return nil, err
	}

	ctx.With(context.WithTimeout(req.Context(), ctx.Timeout))

	req.Header.Set("User-Agent", a.userAgent.Type(t))
	return a.client.Do(req)
}

// AgentContext creates a wrapper for the agent. Allows wrapping contexts, so
// that it becomes possible to cancel a request.
type AgentContext struct {
	URL      *url.URL
	Context  context.Context
	Timeout  time.Duration
	cancelFn context.CancelFunc
}

// NewAgentContext creates a new AgentContext from a given url.URL
// Note: Default time out is set to 10 seconds.
func NewAgentContext(u *url.URL) *AgentContext {
	context, cancelFn := context.WithCancel(context.Background())

	return &AgentContext{
		URL:      u,
		Context:  context,
		Timeout:  10 * time.Second,
		cancelFn: cancelFn,
	}
}

// With takes a context and a cancelFn so it's possible to chain cancelling.
func (a *AgentContext) With(context context.Context, cancelFn context.CancelFunc) {
	a.Context = context
	a.cancelFn = cancelFn
}
