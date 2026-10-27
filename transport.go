// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimit is a response's rate-limit state, from its RateLimit-* headers.
// Read it rather than hard-coding limits. Responses a shared cache may store
// (anonymous forecasts, for example) carry none.
type RateLimit struct {
	// Limit is the request budget of the current window (RateLimit-Limit).
	Limit int
	// Remaining is what is left of it (RateLimit-Remaining).
	Remaining int
	// Reset is the number of seconds until the window resets (RateLimit-Reset).
	Reset int
	// Policy is the raw RateLimit-Policy, for example "60;w=60".
	Policy string
}

func parseRateLimit(h http.Header) *RateLimit {
	l, r, s, p := h.Get("RateLimit-Limit"), h.Get("RateLimit-Remaining"), h.Get("RateLimit-Reset"), h.Get("RateLimit-Policy")
	if l == "" && r == "" && s == "" && p == "" {
		return nil
	}
	atoi := func(v string) int { n, _ := strconv.Atoi(strings.TrimSpace(v)); return n }
	return &RateLimit{Limit: atoi(l), Remaining: atoi(r), Reset: atoi(s), Policy: p}
}

// ResponseMeta describes how a response was obtained. Every top-level result
// type embeds it.
type ResponseMeta struct {
	// Header is the response's HTTP header (for a revalidated response, the
	// fresh header of the 304).
	Header http.Header
	// RateLimit is the response's rate-limit state, or nil when it had none.
	RateLimit *RateLimit
	// RequestID is the response's X-Request-Id, when present.
	RequestID string
	// Revalidated reports that the server answered 304 Not Modified and the
	// body came from the local cache.
	Revalidated bool
}

func (m *ResponseMeta) setMeta(v ResponseMeta) { *m = v }

type metaSetter interface{ setMeta(ResponseMeta) }

// operation describes one API operation.
type operation struct {
	id   string // the OpenAPI operationId
	open bool   // security: [] — sent without credentials by default
}

const maxRedirects = 5

// get performs one operation: the canonical URL, credentials per route,
// redirects, retries, the ETag cache and deprecation warnings. It decodes the
// JSON body into out.
func (c *Client) get(ctx context.Context, op operation, segments []string, q *query, out metaSetter) error {
	if c.baseErr != nil {
		return usageError(op.id, "%v", c.baseErr)
	}
	reqURL := c.buildURL(segments, q)
	key := c.cfg.apiKey.value()
	if op.open && !c.cfg.keyOnOpenRoutes {
		key = ""
	}

	ck := cacheKey(reqURL, key)
	var cached *CacheEntry
	if c.cfg.cache != nil {
		if e, ok := c.cfg.cache.Get(ctx, ck); ok && e != nil {
			cached = e
		}
	}
	if cached != nil && cached.MaxAge > 0 && c.cfg.now().Sub(cached.StoredAt) < cached.MaxAge {
		return decode(op.id, cached.Body, ResponseMeta{Header: cached.Header, RateLimit: parseRateLimit(cached.Header)}, out)
	}
	etag := ""
	if cached != nil {
		etag = cached.ETag
	}

	resp, body, err := c.retrying(ctx, op.id, reqURL, key, etag)
	if err != nil {
		return err
	}
	c.warnDeprecation(op.id, resp.Header)
	meta := ResponseMeta{Header: resp.Header, RateLimit: parseRateLimit(resp.Header), RequestID: resp.Header.Get("X-Request-Id")}

	if resp.StatusCode == http.StatusNotModified {
		if cached == nil {
			return &Error{Category: CategoryUnavailable, Operation: op.id, Status: resp.StatusCode,
				Message: "304 Not Modified for a request the client has no cached copy of", RequestID: meta.RequestID}
		}
		header := cached.Header.Clone()
		for k, v := range resp.Header {
			header[k] = v
		}
		cc := parseCacheControl(header.Get("Cache-Control"))
		newTag := resp.Header.Get("ETag")
		if newTag == "" {
			newTag = cached.ETag
		}
		c.cfg.cache.Set(ctx, ck, &CacheEntry{ETag: newTag, Body: cached.Body, Header: header, StoredAt: c.cfg.now(), MaxAge: cc.maxAge})
		meta.Revalidated = true
		return decode(op.id, cached.Body, meta, out)
	}

	if c.cfg.cache != nil {
		cc := parseCacheControl(resp.Header.Get("Cache-Control"))
		etag := resp.Header.Get("ETag")
		if !cc.noStore && (etag != "" || cc.maxAge > 0) {
			c.cfg.cache.Set(ctx, ck, &CacheEntry{ETag: etag, Body: body, Header: resp.Header, StoredAt: c.cfg.now(), MaxAge: cc.maxAge})
		} else if cached != nil {
			c.cfg.cache.Delete(ctx, ck)
		}
	}
	return decode(op.id, body, meta, out)
}

func decode(op string, body []byte, meta ResponseMeta, out metaSetter) error {
	if err := json.Unmarshal(body, out); err != nil {
		return &Error{Category: CategoryUnavailable, Operation: op, Status: http.StatusOK,
			Message: "the response is not valid JSON for this operation: " + err.Error(), RequestID: meta.RequestID, Err: err}
	}
	out.setMeta(meta)
	return nil
}

func (c *Client) buildURL(segments []string, q *query) string {
	var b strings.Builder
	b.WriteString(c.base.Scheme)
	b.WriteString("://")
	b.WriteString(c.base.Host)
	b.WriteString(c.base.EscapedPath())
	b.WriteString("/api/v1")
	for _, s := range segments {
		b.WriteByte('/')
		b.WriteString(escape(s, ""))
	}
	if q != nil {
		if qs := q.encode(); qs != "" {
			b.WriteByte('?')
			b.WriteString(qs)
		}
	}
	return b.String()
}

// retrying sends the request, retrying connection errors (except TLS
// verification failures), 429, 502, 503 and 504 up to MaxRetries times. It
// returns the final 2xx or 304 response with its body, or an *Error.
func (c *Client) retrying(ctx context.Context, op, reqURL, key, etag string) (*http.Response, []byte, error) {
	for attempt := 0; ; attempt++ {
		resp, body, err := c.follow(ctx, reqURL, key, etag)
		if err != nil {
			cerr := connectionError(op, err)
			if ctx.Err() != nil || !retryableConnErr(err) || attempt >= c.cfg.maxRetries {
				return nil, nil, cerr
			}
			if err := c.cfg.sleep(ctx, c.backoff(attempt+1)); err != nil {
				return nil, nil, connectionError(op, err)
			}
			continue
		}
		if resp.StatusCode/100 == 2 || resp.StatusCode == http.StatusNotModified {
			return resp, body, nil
		}
		apiErr := c.responseError(op, resp, body, key)
		if !retryableStatus(resp.StatusCode) || attempt >= c.cfg.maxRetries {
			return nil, nil, apiErr
		}
		delay := c.backoff(attempt + 1)
		if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable) &&
			resp.Header.Get("Retry-After") != "" {
			if apiErr.RetryAfter > 60*time.Second {
				return nil, nil, apiErr // not waited out; the caller decides
			}
			delay = apiErr.RetryAfter
		}
		if err := c.cfg.sleep(ctx, delay); err != nil {
			return nil, nil, connectionError(op, err)
		}
	}
}

func retryableStatus(s int) bool {
	return s == http.StatusTooManyRequests || s == http.StatusBadGateway ||
		s == http.StatusServiceUnavailable || s == http.StatusGatewayTimeout
}

func retryableConnErr(err error) bool {
	var (
		verr  *tls.CertificateVerificationError
		uaerr x509.UnknownAuthorityError
		herr  x509.HostnameError
		cierr x509.CertificateInvalidError
	)
	return !errors.As(err, &verr) && !errors.As(err, &uaerr) && !errors.As(err, &herr) && !errors.As(err, &cierr) &&
		!errors.Is(err, errBodyTooLarge) && !errors.Is(err, errTooManyRedirects)
}

// backoff is full jitter: a random delay in [0, min(8 s, 0.5 s × 2^(n-1))].
func (c *Client) backoff(n int) time.Duration {
	ceiling := 0.5 * float64(int(1)<<min(n-1, 10))
	ceiling = min(ceiling, 8)
	return time.Duration(c.cfg.jitter() * ceiling * float64(time.Second))
}

func connectionError(op string, err error) *Error {
	return &Error{Category: CategoryConnection, Operation: op, Message: err.Error(), Err: err}
}

var (
	errBodyTooLarge     = fmt.Errorf("response body exceeds %d bytes", MaxResponseBytes)
	errTooManyRedirects = fmt.Errorf("stopped after %d redirects", maxRedirects)
)

// follow sends one attempt and follows 301, 302, 307 and 308 redirects (at
// most 5). The key is kept only while the target stays on the base URL's
// scheme and host.
func (c *Client) follow(ctx context.Context, reqURL, key, etag string) (*http.Response, []byte, error) {
	actx, cancel := context.WithTimeout(ctx, c.cfg.timeout)
	defer cancel()
	target := reqURL
	for hops := 0; ; hops++ {
		req, err := http.NewRequestWithContext(actx, http.MethodGet, target, nil)
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent(c.cfg.appToken))
		if key != "" {
			req.Header.Set("X-API-Key", key)
		}
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, nil, err
		}
		if isRedirect(resp.StatusCode) && resp.Header.Get("Location") != "" {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
			_ = resp.Body.Close()
			if hops >= maxRedirects {
				return nil, nil, errTooManyRedirects
			}
			next, err := req.URL.Parse(resp.Header.Get("Location"))
			if err != nil {
				return nil, nil, fmt.Errorf("invalid redirect Location: %w", err)
			}
			if !sameOrigin(next, c.base) {
				key = ""
			}
			target = next.String()
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes+1))
		_ = resp.Body.Close()
		if err != nil {
			return nil, nil, err
		}
		if len(body) > MaxResponseBytes {
			return nil, nil, errBodyTooLarge
		}
		return resp, body, nil
	}
}

func isRedirect(s int) bool {
	return s == http.StatusMovedPermanently || s == http.StatusFound ||
		s == http.StatusTemporaryRedirect || s == http.StatusPermanentRedirect
}

func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

type errorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (c *Client) responseError(op string, resp *http.Response, body []byte, key string) *Error {
	e := &Error{
		Operation: op,
		Status:    resp.StatusCode,
		RequestID: resp.Header.Get("X-Request-Id"),
		RateLimit: parseRateLimit(resp.Header),
	}
	var eb errorBody
	if json.Unmarshal(body, &eb) == nil && eb.Error != "" {
		e.Code, e.Message = ErrorCode(eb.Error), eb.Message
	} else {
		e.Message = strings.TrimSpace(string(body[:min(len(body), 200)]))
		if e.Message == "" {
			e.Message = http.StatusText(resp.StatusCode)
		}
	}
	if key != "" {
		e.Message = strings.ReplaceAll(e.Message, key, "[redacted]")
	}
	e.Category = categorize(e.Status, e.Code)
	if e.Status == http.StatusTooManyRequests || e.Status == http.StatusServiceUnavailable {
		e.RetryAfter = parseRetryAfter(resp.Header.Get("Retry-After"), c.cfg.now())
	}
	return e
}

// parseRetryAfter reads delta-seconds or an HTTP-date.
func parseRetryAfter(v string, now time.Time) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(max(n, 0)) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return max(t.Sub(now), 0)
	}
	return 0
}

// deprecationWarned records the operations already warned about, so each is
// reported once per process.
var deprecationWarned sync.Map

func (c *Client) warnDeprecation(op string, h http.Header) {
	dep, sunset := h.Get("Deprecation"), h.Get("Sunset")
	if dep == "" && sunset == "" {
		return
	}
	if _, seen := deprecationWarned.LoadOrStore(op, true); seen {
		return
	}
	link := deprecationLink(h.Values("Link"))
	msg := "ensembleweather: " + op + " is deprecated"
	if sunset != "" {
		msg += " (sunset " + sunset + ")"
	}
	if link != "" {
		msg += "; see " + link
	}
	c.cfg.logger.Warn(msg, "operation", op, "deprecation", dep, "sunset", sunset, "link", link)
}

// deprecationLink returns the target of a Link with rel="deprecation".
func deprecationLink(values []string) string {
	for _, v := range values {
		for _, link := range strings.Split(v, ",") {
			target, params, ok := strings.Cut(link, ";")
			if !ok {
				continue
			}
			for _, p := range strings.Split(params, ";") {
				name, val, _ := strings.Cut(strings.TrimSpace(p), "=")
				if strings.EqualFold(name, "rel") && strings.EqualFold(strings.Trim(val, `"`), "deprecation") {
					return strings.Trim(strings.TrimSpace(target), "<>")
				}
			}
		}
	}
	return ""
}
