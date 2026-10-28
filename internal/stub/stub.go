// SPDX-License-Identifier: BSD-2-Clause

// Package stub is a local stand-in for the Ourea API, used to run the
// examples in CI without network access or a key. It serves the shared
// vector bodies from spec/vectors/bodies and echoes probability and totals
// requests back with fixed answers. It is a test fixture: its numbers are
// placeholders, not forecasts.
package stub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func root() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

// read loads a JSON fixture; nil means it is missing or broken, which the
// handler answers with a 500 so the example under test fails loudly.
func read(rel string) map[string]any {
	raw, err := os.ReadFile(filepath.Join(root(), rel))
	if err != nil {
		return nil
	}
	var v map[string]any
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return v
}

func body(name string) map[string]any { return read(filepath.Join("spec", "vectors", "bodies", name)) }

func fixture(name string) map[string]any {
	return read(filepath.Join("internal", "stub", "testdata", name))
}

// Start starts the stub API on a local port. Close it when done.
func Start() *httptest.Server { return httptest.NewServer(http.HandlerFunc(serve)) }

var origin = time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

func serve(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	keyed := r.Header.Get("X-API-Key") != ""
	path := strings.TrimPrefix(r.URL.EscapedPath(), "/api/v1/")
	parts := strings.Split(path, "/")
	switch {
	case path == "forecast":
		if !keyed {
			writeError(w, 403, "PRO_REQUIRED", "Forecasts at arbitrary coordinates need a Pro API key.")
			return
		}
		lat, _ := strconv.ParseFloat(q.Get("lat"), 64)
		lon, _ := strconv.ParseFloat(q.Get("lon"), 64)
		if lon > -60 || lat < 20 {
			writeError(w, 404, "OUTSIDE_COVERAGE", fmt.Sprintf("%s,%s is outside the forecast domain (CONUS).", q.Get("lat"), q.Get("lon")))
			return
		}
		write(w, r, keyed, forecast(q))
	case len(parts) == 3 && parts[0] == "locations":
		write(w, r, keyed, forecast(q))
	case len(parts) == 4 && parts[0] == "locations" && parts[3] == "current":
		write(w, r, keyed, fixture("current.json"))
	case len(parts) == 4 && parts[0] == "locations" && parts[3] == "alerts":
		write(w, r, keyed, fixture("alerts.json"))
	case path == "search/locations":
		write(w, r, false, body("search.json"))
	case path == "catalog/locations":
		page := "catalog-page-1.json"
		if q.Get("cursor") == "c2" {
			page = "catalog-page-2.json"
		}
		write(w, r, keyed, body(page))
	case len(parts) == 3 && parts[0] == "catalog" && parts[1] == "locations":
		write(w, r, false, body("catalog-detail.json"))
	case path == "attribution":
		write(w, r, false, body("attribution.json"))
	default:
		writeError(w, 404, "NOT_FOUND", "No route "+r.URL.Path+".")
	}
}

func forecast(q map[string][]string) map[string]any {
	get := func(k string) string {
		if v := q[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	units := get("units")
	f := body("forecast-metric.json")
	if units == "imperial" {
		f = body("forecast-imperial.json")
	}
	if f == nil {
		return nil
	}
	if p := get("prob"); p != "" {
		var out []any
		for i, term := range strings.Split(p, ",") {
			out = append(out, probability(term, i))
		}
		f["probabilities"] = out
	}
	if t := get("totals"); t != "" {
		var out []any
		for _, term := range strings.Split(t, ",") {
			out = append(out, total(term))
		}
		f["window_totals"] = out
	}
	return f
}

func window(s string) (int, int) {
	s = strings.TrimSuffix(s, "h")
	if a, b, ok := strings.Cut(s, "-"); ok {
		x, _ := strconv.Atoi(a)
		y, _ := strconv.Atoi(b)
		return x, y
	}
	y, _ := strconv.Atoi(s)
	return 0, y
}

func probability(term string, i int) map[string]any {
	f := strings.Split(term, ":")
	threshold, _ := strconv.ParseFloat(f[2], 64)
	start, end := window(f[3])
	agg := "any"
	if f[0] == "precipitation" || f[0] == "snowfall" {
		agg = "total"
	}
	if len(f) > 4 {
		agg = f[4]
	}
	return map[string]any{
		"term": term, "variable": f[0], "op": f[1], "threshold": threshold, "aggregation": agg,
		"window_start_hours": start, "window_end_hours": end,
		"from": origin.Add(time.Duration(start) * time.Hour), "to": origin.Add(time.Duration(end) * time.Hour),
		"probability": []float64{0.42, 0.17, 0.08, 0.63}[i%4], "members": 131, "null_reason": nil,
	}
}

func total(term string) map[string]any {
	v, w, _ := strings.Cut(term, ":")
	start, end := window(w)
	return map[string]any{
		"term": term, "variable": v, "window_start_hours": start, "window_end_hours": end, "aggregation": "total",
		"from": origin.Add(time.Duration(start) * time.Hour), "to": origin.Add(time.Duration(end) * time.Hour),
		"quantiles":   map[string]any{"mean": 31.0, "p10": 12.0, "p25": 20.0, "p50": 29.0, "p75": 40.0, "p90": 52.0, "min": 4.0, "max": 88.0},
		"members":     131,
		"null_reason": nil,
	}
}

func write(w http.ResponseWriter, r *http.Request, keyed bool, v map[string]any) {
	if v == nil {
		writeError(w, 500, "INTERNAL_ERROR", "stub fixture missing")
		return
	}
	raw, _ := json.Marshal(v)
	etag := fmt.Sprintf(`"stub-%x"`, len(raw))
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("ETag", etag)
	if keyed {
		h.Set("Cache-Control", "private, max-age=0")
		h.Set("RateLimit-Limit", "600")
		h.Set("RateLimit-Remaining", "599")
		h.Set("RateLimit-Reset", "60")
		h.Set("RateLimit-Policy", "600;w=60")
	} else {
		h.Set("Cache-Control", "public, max-age=0, s-maxage=600, stale-while-revalidate=60")
	}
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	_, _ = w.Write(raw)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-Id", "stub-request")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": code, "message": msg, "request_id": "stub-request",
		"doc_url": "https://docs.oureawx.com/features/errors/#" + strings.ToLower(code),
	})
}

var stdoutMu sync.Mutex

// RunMain runs an example's main with the given environment (OUREA_BASE_URL
// pointing at a fresh stub) and returns what it printed.
func RunMain(t *testing.T, main func(), env map[string]string) string {
	t.Helper()
	srv := Start()
	defer srv.Close()
	t.Setenv("OUREA_BASE_URL", srv.URL)
	t.Setenv("OUREA_API_KEY", "")
	for k, v := range env {
		t.Setenv(k, v)
	}
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	rd, wr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = wr
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(rd)
		done <- string(b)
	}()
	func() {
		defer func() { os.Stdout = saved }()
		main()
	}()
	_ = wr.Close()
	out := <-done
	t.Log(out)
	return out
}
