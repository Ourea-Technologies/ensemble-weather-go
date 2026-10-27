// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRetries(t *testing.T) {
	connReset := errors.New("connection reset by peer")
	tests := []struct {
		name     string
		retries  int
		script   []func() (*http.Response, error)
		wantErr  error
		requests int
		delays   []float64
	}{
		{"connection error retried", 2, []func() (*http.Response, error){
			func() (*http.Response, error) { return nil, connReset },
			func() (*http.Response, error) { return respond(200, minimalForecast), nil },
		}, nil, 2, []float64{0.5}},
		{"TLS verification never retried", 2, []func() (*http.Response, error){
			func() (*http.Response, error) {
				return nil, &tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}}
			},
		}, ErrConnection, 1, nil},
		{"504 retried, backoff doubles and caps", 5, []func() (*http.Response, error){
			func() (*http.Response, error) { return respond(504, ""), nil },
			func() (*http.Response, error) { return respond(504, ""), nil },
			func() (*http.Response, error) { return respond(504, ""), nil },
			func() (*http.Response, error) { return respond(504, ""), nil },
			func() (*http.Response, error) { return respond(504, ""), nil },
			func() (*http.Response, error) { return respond(504, ""), nil },
		}, ErrUnavailable, 6, []float64{0.5, 1, 2, 4, 8}},
		{"500 not retried", 2, []func() (*http.Response, error){
			func() (*http.Response, error) { return respond(500, `{"error":"INTERNAL_ERROR","message":"x"}`), nil },
		}, ErrUnavailable, 1, nil},
		{"503 Retry-After honoured", 2, []func() (*http.Response, error){
			func() (*http.Response, error) { return respond(503, "", "Retry-After", "7"), nil },
			func() (*http.Response, error) { return respond(200, minimalForecast), nil },
		}, nil, 2, []float64{7}},
		{"502 Retry-After ignored", 2, []func() (*http.Response, error){
			func() (*http.Response, error) { return respond(502, "", "Retry-After", "7"), nil },
			func() (*http.Response, error) { return respond(200, minimalForecast), nil },
		}, nil, 2, []float64{0.5}},
		{"429 HTTP-date Retry-After", 2, []func() (*http.Response, error){
			func() (*http.Response, error) {
				return respond(429, "", "Retry-After", "Sun, 15 Nov 2026 00:00:30 GMT"), nil
			},
			func() (*http.Response, error) { return respond(200, minimalForecast), nil },
		}, nil, 2, []float64{30}},
		{"404 not retried", 2, []func() (*http.Response, error){
			func() (*http.Response, error) {
				return respond(404, `{"error":"LOCATION_NOT_FOUND","message":"no"}`), nil
			},
		}, ErrNotFound, 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rt, delays := testClient(t, func(n int, _ *http.Request) (*http.Response, error) {
				return tt.script[n-1]()
			}, WithMaxRetries(tt.retries))
			_, err := c.Forecast(context.Background(), "or/x", nil)
			if tt.wantErr == nil && err != nil || tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if len(rt.reqs) != tt.requests {
				t.Errorf("requests = %d, want %d", len(rt.reqs), tt.requests)
			}
			if fmt.Sprint(*delays) != fmt.Sprint(tt.delays) && (len(*delays) != 0 || len(tt.delays) != 0) {
				t.Errorf("delays = %v, want %v", *delays, tt.delays)
			}
		})
	}
}

func TestCanceledContextNotRetried(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, rt, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		cancel()
		return nil, context.Canceled
	})
	_, err := c.Forecast(ctx, "or/x", nil)
	if !errors.Is(err, ErrConnection) || !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want a connection error wrapping context.Canceled", err)
	}
	if len(rt.reqs) != 1 {
		t.Errorf("requests = %d, want 1", len(rt.reqs))
	}
}

func TestCategorize(t *testing.T) {
	tests := []struct {
		status int
		code   ErrorCode
		want   Category
	}{
		{401, "", CategoryAuthentication}, {400, CodeUnauthorized, CategoryAuthentication},
		{403, "", CategoryPermission}, {403, CodeProRequired, CategoryPermission},
		{404, "", CategoryNotFound}, {410, "", CategoryNotFound}, {404, CodeModelRunNotFound, CategoryNotFound},
		{400, "", CategoryInvalidRequest}, {400, CodeAmbiguous, CategoryInvalidRequest}, {422, "SOMETHING", CategoryInvalidRequest},
		{400, CodeOutsideCoverage, CategoryOutsideCoverage}, {404, CodeOutsideCoverage, CategoryOutsideCoverage},
		{429, "", CategoryRateLimited}, {500, "", CategoryUnavailable}, {504, CodeTimeout, CategoryUnavailable},
		{400, CodeTooManyRequests, CategoryRateLimited},
	}
	for _, tt := range tests {
		if got := categorize(tt.status, tt.code); got != tt.want {
			t.Errorf("categorize(%d, %q) = %s, want %s", tt.status, tt.code, got, tt.want)
		}
	}
	for c := CategoryAuthentication; c <= CategoryUsage; c++ {
		e := &Error{Category: c}
		for c2, s := range sentinels {
			if errors.Is(e, s) != (c == c2) {
				t.Errorf("errors.Is(%s error, %v) = %v", c, s, c != c2)
			}
		}
	}
	if Category(99).String() != "Category(99)" {
		t.Error("unknown category String")
	}
}

func TestErrorFields(t *testing.T) {
	c, _, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		return respond(429, `{"error":"TOO_MANY_REQUESTS","message":"slow down"}`,
			"Retry-After", "120", "X-Request-Id", "req-1", "RateLimit-Limit", "60", "RateLimit-Remaining", "0"), nil
	})
	_, err := c.Forecast(context.Background(), "or/x", nil)
	var e *Error
	if !errors.As(err, &e) {
		t.Fatal(err)
	}
	if e.Status != 429 || e.Code != CodeTooManyRequests || e.RetryAfter != 120*time.Second || e.RequestID != "req-1" ||
		e.RateLimit == nil || e.RateLimit.Remaining != 0 || e.RateLimit.Limit != 60 || e.Operation != "getLocationForecast" {
		t.Errorf("error = %+v", e)
	}
	if want := "ensembleweather: getLocationForecast: 429 TOO_MANY_REQUESTS: slow down (request ID req-1)"; e.Error() != want {
		t.Errorf("Error() = %q, want %q", e.Error(), want)
	}
}

func TestCacheRevalidation(t *testing.T) {
	c, rt, _ := testClient(t, func(_ int, r *http.Request) (*http.Response, error) {
		if r.Header.Get("If-None-Match") == `"v1"` {
			return respond(304, "", "ETag", `"v1"`, "RateLimit-Remaining", "8"), nil
		}
		return respond(200, minimalForecast, "ETag", `"v1"`, "Cache-Control", "private, max-age=0", "RateLimit-Remaining", "9"), nil
	}, WithAPIKey("ow_a"))
	ctx := context.Background()
	first, err := c.Forecast(ctx, "or/x", nil)
	if err != nil || first.Revalidated {
		t.Fatalf("first: %v revalidated=%v", err, first.Revalidated)
	}
	second, err := c.Forecast(ctx, "or/x", nil)
	if err != nil || !second.Revalidated || second.Location.Slug != "mt-hood-meadows" || second.RateLimit.Remaining != 8 {
		t.Fatalf("second: %v %+v", err, second.ResponseMeta)
	}
	// Another key, or keyless, never sees the entry.
	_, _ = c.With(WithAPIKey("ow_b")).Forecast(ctx, "or/x", nil)
	_, _ = c.With(WithAPIKey("")).Forecast(ctx, "or/x", nil)
	for i, r := range rt.reqs[2:] {
		if r.Header.Get("If-None-Match") != "" {
			t.Errorf("request %d under another credential sent If-None-Match", i+3)
		}
	}
}

func TestCacheHonoursMaxAge(t *testing.T) {
	c, rt, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		return respond(200, minimalForecast, "Cache-Control", "public, max-age=60, s-maxage=3600, stale-while-revalidate=300"), nil
	})
	for i := 0; i < 3; i++ {
		if _, err := c.Forecast(context.Background(), "or/x", nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(rt.reqs) != 1 {
		t.Errorf("requests = %d, want 1 (fresh for max-age=60, the clock is fixed)", len(rt.reqs))
	}
	if cc := parseCacheControl("public, max-age=0, s-maxage=3600"); cc.maxAge != 0 {
		t.Errorf("s-maxage leaked into the private max-age: %v", cc.maxAge)
	}
	if cc := parseCacheControl("max-age=60, no-cache"); cc.maxAge != 0 || !cc.noCache {
		t.Errorf("no-cache: %+v", cc)
	}
}

func TestMemoryCacheBounds(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(2, 1<<20)
	c.Set(ctx, "a", &CacheEntry{Body: []byte("1")})
	c.Set(ctx, "b", &CacheEntry{Body: []byte("2")})
	c.Get(ctx, "a") // a is now the most recent
	c.Set(ctx, "c", &CacheEntry{Body: []byte("3")})
	if _, ok := c.Get(ctx, "b"); ok {
		t.Error("b should have been evicted")
	}
	if _, ok := c.Get(ctx, "a"); !ok {
		t.Error("a was evicted")
	}
	small := NewMemoryCache(10, 100)
	small.Set(ctx, "big", &CacheEntry{Body: make([]byte, 200)})
	if small.Len() != 0 {
		t.Error("an entry over the byte bound was stored")
	}
	for i := 0; i < 5; i++ {
		small.Set(ctx, fmt.Sprint(i), &CacheEntry{Body: make([]byte, 30)})
	}
	if small.Len() != 3 {
		t.Errorf("Len = %d, want 3 under a 100-byte bound", small.Len())
	}
	small.Delete(ctx, "4")
	if small.Len() != 2 {
		t.Errorf("Len after Delete = %d", small.Len())
	}
	if NewMemoryCache(0, 0).maxEntries != DefaultCacheEntries {
		t.Error("default bounds")
	}
}

func TestDeprecationOncePerOperation(t *testing.T) {
	deprecationWarned.Range(func(k, _ any) bool { deprecationWarned.Delete(k); return true })
	logs := &logCapture{}
	c, _, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		return respond(200, minimalForecast, "Sunset", "Wed, 01 Dec 2027 00:00:00 GMT",
			"Link", `</api/v1/attribution>; rel="license", <https://docs.oureawx.com/x>; rel="deprecation"`), nil
	}, WithLogger(slog.New(logs)))
	ctx := context.Background()
	_, _ = c.Forecast(ctx, "or/x", nil)
	_, _ = c.Forecast(ctx, "or/y", nil)
	_, _ = c.Current(ctx, "or/x", nil)
	if len(logs.msgs) != 2 || !strings.Contains(logs.msgs[0], "https://docs.oureawx.com/x") {
		t.Errorf("warnings = %q", logs.msgs)
	}
}

// SDK-side security guarantees (SDK-STANDARD "Security").

func TestKeyNeverPrinted(t *testing.T) {
	const key = "ow_secret_0123456789"
	logs := &logCapture{}
	c, _, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		// A server that echoes the key in its message must not leak it either.
		return respond(401, `{"error":"INVALID_API_KEY","message":"bad key ow_secret_0123456789"}`, "Deprecation", "true"), nil
	}, WithAPIKey(key), WithLogger(slog.New(logs)))
	deprecationWarned.Range(func(k, _ any) bool { deprecationWarned.Delete(k); return true })
	_, err := c.Forecast(context.Background(), "or/x", nil)
	forms := []string{
		err.Error(), fmt.Sprintf("%v", err), fmt.Sprintf("%+v", err), fmt.Sprintf("%#v", err),
		c.String(), fmt.Sprintf("%v", c), fmt.Sprintf("%+v", c), fmt.Sprintf("%#v", c),
		fmt.Sprintf("%v", c.cfg), fmt.Sprintf("%+v", c.cfg), fmt.Sprintf("%#v", c.cfg), fmt.Sprintf("%s", c.cfg.apiKey), fmt.Sprintf("%v", *c), fmt.Sprintf("%+v", *c), fmt.Sprintf("%#v", *c),
	}
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("client", "key", c.cfg.apiKey, "config", c.cfg)
	forms = append(forms, buf.String())
	forms = append(forms, logs.msgs...)
	for _, f := range forms {
		if strings.Contains(f, key) {
			t.Errorf("key leaked: %s", f)
		}
	}
}

func TestKeyDroppedOnCrossHostRedirect(t *testing.T) {
	tests := []struct {
		location string
		keep     bool
	}{
		{"/api/v1/locations/or/x", true},
		{"https://api.oureawx.com/api/v1/locations/or/x", true},
		{"https://evil.example/api/v1/locations/or/x", false},
		{"http://api.oureawx.com/api/v1/locations/or/x", false}, // scheme downgrade
		{"https://api.oureawx.com:8443/x", false},
	}
	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			c, rt, _ := testClient(t, func(n int, _ *http.Request) (*http.Response, error) {
				if n == 1 {
					return respond(307, "", "Location", tt.location), nil
				}
				return respond(200, minimalForecast), nil
			}, WithAPIKey("ow_key"), WithBaseURL(DefaultBaseURL))
			if _, err := c.Forecast(context.Background(), "or/x", nil); err != nil {
				t.Fatal(err)
			}
			if got := rt.last().Header.Get("X-API-Key") != ""; got != tt.keep {
				t.Errorf("key sent after redirect = %v, want %v", got, tt.keep)
			}
		})
	}
}

func TestRedirectLimit(t *testing.T) {
	c, rt, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		return respond(302, "", "Location", "/loop"), nil
	})
	_, err := c.Forecast(context.Background(), "or/x", nil)
	if !errors.Is(err, ErrConnection) || len(rt.reqs) != maxRedirects+1 {
		t.Errorf("err = %v after %d requests", err, len(rt.reqs))
	}
}

func TestNoTLSOffSwitch(t *testing.T) {
	// No Option touches TLS: the exported options are exactly these.
	want := []string{"WithAPIKey", "WithBaseURL", "WithCache", "WithHTTPClient", "WithKeyOnOpenRoutes",
		"WithLogger", "WithMaxRetries", "WithTimeout", "WithUnits", "WithUserAgent"}
	opts := []any{WithAPIKey, WithBaseURL, WithCache, WithHTTPClient, WithKeyOnOpenRoutes, WithLogger, WithMaxRetries, WithTimeout, WithUnits, WithUserAgent}
	for i, o := range opts {
		name := runtime.FuncForPC(reflect.ValueOf(o).Pointer()).Name()
		if !strings.HasSuffix(name, "."+want[i]) {
			t.Errorf("option %d is %s", i, name)
		}
	}
	for _, f := range reflect.VisibleFields(reflect.TypeOf(config{})) {
		if n := strings.ToLower(f.Name); strings.Contains(n, "tls") || strings.Contains(n, "insecure") || strings.Contains(n, "verify") {
			t.Errorf("config field %s looks like a TLS switch", f.Name)
		}
	}
	// The default client verifies certificates.
	c := NewClient()
	if c.http.Transport != nil {
		t.Errorf("default transport = %T, want nil (http.DefaultTransport)", c.http.Transport)
	}
	if tr, ok := http.DefaultTransport.(*http.Transport); ok && tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("http.DefaultTransport skips verification")
	}
}

// endless is an unbounded body that never allocates.
type endless struct{}

func (endless) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	return len(p), nil
}

func TestResponseBodyBounded(t *testing.T) {
	c, rt, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(endless{})}, nil
	})
	_, err := c.Forecast(context.Background(), "or/x", nil)
	if !errors.Is(err, ErrConnection) || !errors.Is(err, errBodyTooLarge) {
		t.Fatalf("err = %v, want the body-size error", err)
	}
	if len(rt.reqs) != 1 {
		t.Errorf("an oversized body was retried (%d requests)", len(rt.reqs))
	}
}

func TestDecodeErrorIsTyped(t *testing.T) {
	c, _, _ := testClient(t, func(int, *http.Request) (*http.Response, error) { return respond(200, "<html>"), nil })
	if _, err := c.Forecast(context.Background(), "or/x", nil); !errors.Is(err, ErrUnavailable) {
		t.Errorf("err = %v, want unavailable", err)
	}
}
