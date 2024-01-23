package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

// GraphqlEndpoint is the GitHub GraphQL API endpoint.
const GraphqlEndpoint string = "https://api.github.com/graphql"

var (
	// ErrMissingToken is returned when the GitHub API token is missing.
	ErrMissingToken = errors.New("missing GitHub API token")

	// ErrMissingEndpoint is returned when the GitHub GraphQL API endpoint is missing.
	ErrMissingEndpoint = errors.New("missing GitHub GraphQL API endpoint")
)

//go:generate go run github.com/vburenin/ifacemaker --file $GOFILE --struct client --iface Client --pkg github --output github_iface.go
type client struct {
	c      *graphql.Client
	logger *slog.Logger
}

type config struct {
	// Token is the GitHub API token.
	Token string

	// HTTPClient is the HTTP client to use for requests.
	HTTPClient *http.Client

	// GraphqlEndpoint is the GitHub GraphQL API endpoint.
	Endpoint string

	// Logger is the logger to use.
	Logger *slog.Logger
}

func defaultEnvs(values []string, def string) string {
	for _, v := range values {
		if vv := os.Getenv(v); vv != "" {
			return vv
		}
	}

	return def
}

func newDefaultConfig() *config {
	return &config{
		Token:      defaultEnvs([]string{"GITHUB_TOKEN", "GH_TOKEN"}, ""),
		HTTPClient: http.DefaultClient,
		Endpoint:   defaultEnvs([]string{"GITHUB_GRAPHQL_ENDPOINT", "GH_GRAPHQL_ENDPOINT"}, GraphqlEndpoint),
		Logger:     slog.Default(),
	}
}

func (c *config) validate() error {
	if c.Token == "" {
		return ErrMissingToken
	}

	if c.Endpoint == "" {
		return ErrMissingEndpoint
	}

	return nil
}

// Option is a client option.
type Option func(*config)

// WithToken sets the GitHub API token.
func WithToken(token string) Option {
	return func(c *config) {
		c.Token = token
	}
}

// WithHTTPClient sets the HTTP client to use for requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *config) {
		c.HTTPClient = httpClient
	}
}

// WithEndpoint sets the GitHub GraphQL API endpoint.
func WithEndpoint(endpoint string) Option {
	return func(c *config) {
		c.Endpoint = endpoint
	}
}

// WithLogger sets the logger to use.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.Logger = logger
	}
}

// New creates a new GitHub API client.
// It uses the GITHUB_TOKEN environment variable for authentication.
func New(ctx context.Context, opts ...Option) (Client, error) {
	conf := newDefaultConfig()
	for _, opt := range opts {
		opt(conf)
	}

	if err := conf.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Token})
	httpClient := oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, conf.HTTPClient), src)

	return &client{
		c:      graphql.NewClient(GraphqlEndpoint, httpClient),
		logger: conf.Logger,
	}, nil
}

// GetStars returns the list of repositories starred by the user.
func (c *client) GetStars(ctx context.Context) <-chan *StarredRepository {
	out := make(chan *StarredRepository)

	vars := map[string]any{
		"count":  100,
		"cursor": "",
	}

	go func() {
		defer close(out)

		for {
			var q query
			if err := c.c.Query(ctx, &q, vars); err != nil {
				c.logger.With(slogx.Err(err)).Error("failed to query")
				continue
			}

			for _, repo := range q.Viewer.StarredRepositories.Repositories {
				c.parseReadme(repo)
				out <- repo
			}

			if !q.Viewer.StarredRepositories.PageInfo.HasNextPage {
				break
			}

			// handle rate limit
			if q.RateLimit.Remaining <= q.RateLimit.Used+10 {
				wait := time.Until(q.RateLimit.ResetAt)
				c.logger.
					With(slog.Any("rate_limit", q.RateLimit)).
					Warn(fmt.Sprintf("rate limit reached (with 10 units buffer), will be reset in %s", wait))
				return
			}

			vars["cursor"] = q.Viewer.StarredRepositories.PageInfo.EndCursor
		}
	}()

	return out
}

// parseReadme parses the readme from the repository.
// Tries to parse README from all possible locations.
func (c *client) parseReadme(repo *StarredRepository) {
	readmes := []*readme{repo.Repository.R1, repo.Repository.R2}
	for _, r := range readmes {
		if r == nil {
			continue
		}

		repo.Repository.Readme = r.Blob.Text
	}
}
