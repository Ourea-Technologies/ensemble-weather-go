// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

// DefaultBaseURL is the API's address.
const DefaultBaseURL = "https://api.oureawx.com"

// Defaults of the client's settings.
const (
	DefaultTimeout    = 30 * time.Second
	DefaultMaxRetries = 2
)

// Environment variables the client reads in [NewClient] when the matching
// option is not given.
const (
	// EnvAPIKey holds the API key.
	EnvAPIKey = "OUREA_API_KEY" //nolint:gosec // G101: the variable's name, not a credential
	// EnvBaseURL overrides the base URL (for proxies and local test servers).
	EnvBaseURL = "OUREA_BASE_URL"
)

// MaxResponseBytes bounds a response body. A larger body is an error, never
// an unbounded allocation.
const MaxResponseBytes = 64 << 20

// Client calls the Ourea API. It is safe for concurrent use; create one and
// reuse it, since it holds the connection pool and the response cache.
//
// The zero Client is not usable; call [NewClient].
type Client struct {
	cfg     config
	base    *url.URL
	baseErr error
	http    *http.Client
}

// secret holds the API key behind a pointer, so printing a struct that holds
// it (even with %#v, which skips the methods of unexported fields) shows an
// address, never the key.
type secret struct{ key string }

func newSecret(key string) *secret {
	if key == "" {
		return nil
	}
	return &secret{key}
}

// value returns the key, or "" for none.
func (s *secret) value() string {
	if s == nil {
		return ""
	}
	return s.key
}

// Format prints a placeholder for every verb.
func (s *secret) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte("[redacted]")) }

// LogValue keeps the key out of slog output.
func (s *secret) LogValue() slog.Value { return slog.StringValue("[redacted]") }

type config struct {
	apiKey          *secret
	apiKeySet       bool
	baseURL         string
	baseURLSet      bool
	httpClient      *http.Client
	units           Units
	timeout         time.Duration
	maxRetries      int
	cache           Cache
	cacheSet        bool
	keyOnOpenRoutes bool
	appToken        string
	logger          *slog.Logger

	// Test seams: the clock, the jitter source (a value in [0, 1]) and the
	// sleep between retries.
	now    func() time.Time
	jitter func() float64
	sleep  func(ctx context.Context, d time.Duration) error
}

// Option configures a [Client].
type Option func(*config)

// WithAPIKey sets the API key, sent as X-API-Key. Without this option the
// client reads OUREA_API_KEY; with no key at all, calls are anonymous (the
// Free tier), which is valid. WithAPIKey("") forces anonymous calls.
func WithAPIKey(key string) Option {
	return func(c *config) { c.apiKey, c.apiKeySet = newSecret(key), true }
}

// WithBaseURL sets the API's base URL (default [DefaultBaseURL], or
// OUREA_BASE_URL when set). A trailing slash is ignored, and a path prefix is
// kept: "https://proxy.example/ourea/" calls ".../ourea/api/v1/...".
func WithBaseURL(u string) Option {
	return func(c *config) { c.baseURL, c.baseURLSet = u, true }
}

// WithHTTPClient sets the HTTP client used to send requests (default: a new
// client over [http.DefaultTransport]). The SDK follows redirects itself, so
// it uses a copy of hc with redirect-following turned off; hc is not changed.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) { c.httpClient = hc }
}

// WithUnits sets the unit system for every request (default [UnitsMetric]).
// A request's own Units field overrides it.
func WithUnits(u Units) Option {
	return func(c *config) { c.units = u }
}

// WithTimeout bounds each attempt (default 30 s). The whole call, with its
// retries, is bounded by the attempts allowed and by ctx.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithMaxRetries sets how many times a retryable failure is retried
// (default 2, so 3 attempts in all). 0 disables retrying.
func WithMaxRetries(n int) Option {
	return func(c *config) { c.maxRetries = max(n, 0) }
}

// WithCache replaces the response cache. nil disables caching.
func WithCache(cache Cache) Option {
	return func(c *config) { c.cache, c.cacheSet = cache, true }
}

// WithKeyOnOpenRoutes sends the API key on the routes that need no
// credentials too (search, catalog details, attribution). By default those
// are sent without the key so every caller shares one edge-cache entry; use
// this when every request must be attributed to the key.
func WithKeyOnOpenRoutes() Option {
	return func(c *config) { c.keyOnOpenRoutes = true }
}

// WithUserAgent appends an application product token, such as
// "my-app/1.2", to the SDK's User-Agent. It must not contain user data.
func WithUserAgent(product string) Option {
	return func(c *config) { c.appToken = strings.TrimSpace(product) }
}

// WithLogger sets the logger for deprecation warnings (default
// [slog.Default]). The API key is never logged.
func WithLogger(l *slog.Logger) Option {
	return func(c *config) { c.logger = l }
}

// NewClient returns a client configured by opts. A configuration error (an
// invalid base URL) is reported by every call rather than here.
func NewClient(opts ...Option) *Client {
	cfg := config{
		units:      UnitsMetric,
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
		now:        time.Now,
		jitter:     rand.Float64,
		sleep:      sleepContext,
	}
	for _, o := range opts {
		o(&cfg)
	}
	if !cfg.apiKeySet {
		cfg.apiKey, cfg.apiKeySet = newSecret(os.Getenv(EnvAPIKey)), true
	}
	if !cfg.baseURLSet {
		cfg.baseURL = os.Getenv(EnvBaseURL)
		if cfg.baseURL == "" {
			cfg.baseURL = DefaultBaseURL
		}
		cfg.baseURLSet = true
	}
	if !cfg.cacheSet {
		cfg.cache, cfg.cacheSet = NewMemoryCache(DefaultCacheEntries, DefaultCacheBytes), true
	}
	return build(cfg)
}

func build(cfg config) *Client {
	if cfg.units == "" {
		cfg.units = UnitsMetric
	}
	if cfg.logger == nil {
		cfg.logger = slog.Default()
	}
	hc := http.Client{}
	if cfg.httpClient != nil {
		hc = *cfg.httpClient
	}
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c := &Client{cfg: cfg, http: &hc}
	c.base, c.baseErr = parseBaseURL(cfg.baseURL)
	return c
}

// With returns a copy of c with opts applied on top of its configuration.
// The copy shares c's cache (unless opts replace it) and HTTP client, so it is
// cheap; use it for a per-request key or units:
//
//	other := client.With(ensembleweather.WithAPIKey(customerKey))
func (c *Client) With(opts ...Option) *Client {
	cfg := c.cfg
	for _, o := range opts {
		o(&cfg)
	}
	return build(cfg)
}

// String describes the client without its API key.
func (c Client) String() string {
	auth := "anonymous"
	if c.cfg.apiKey != nil {
		auth = "api key [redacted]"
	}
	return fmt.Sprintf("ensembleweather.Client{base: %s, %s}", c.cfg.baseURL, auth)
}

// GoString is String, so %#v never prints the key.
func (c Client) GoString() string { return c.String() }

func parseBaseURL(s string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimRight(s, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return nil, fmt.Errorf("invalid base URL %q: want http(s)://host[/prefix]", s)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("invalid base URL %q: no query or fragment allowed", s)
	}
	return u, nil
}

func userAgent(app string) string {
	ua := "ensemble-weather-go/" + Version + " (go/" + strings.TrimPrefix(runtime.Version(), "go") + ")"
	if app != "" {
		ua += " " + app
	}
	return ua
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Ptr returns a pointer to v, for the optional fields of request parameters:
//
//	&ensembleweather.ForecastParams{Nowcast: ensembleweather.Ptr(false)}
func Ptr[T any](v T) *T { return &v }
