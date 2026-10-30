// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

// The shared vectors runner (BEHAVIOR 13). Every case in spec/vectors runs
// through a mock RoundTripper with a fake clock and a jitter source pinned to
// its upper bound. The only per-language code is runCall, which maps a case's
// semantic arguments onto this SDK's API; a case it cannot map fails.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const vectorsDir = "spec/vectors"

type vectorCase struct {
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Client         vectorClient      `json:"client"`
	Call           *vectorCall       `json:"call"`
	Calls          []vectorCall      `json:"calls"`
	ExpectRequest  *json.RawMessage  `json:"expect_request"`
	ExpectRequests []expectedRequest `json:"expect_requests"`
	Responses      []vectorResponse  `json:"responses"`
	Expect         vectorExpect      `json:"expect"`
}

type vectorClient struct {
	APIKey              *string           `json:"api_key"`
	BaseURL             *string           `json:"base_url"`
	MaxRetries          *int              `json:"max_retries"`
	Cache               *bool             `json:"cache"`
	SendKeyOnOpenRoutes bool              `json:"send_key_on_open_routes"`
	Env                 map[string]string `json:"env"`
}

type vectorCall struct {
	Operation  string                     `json:"operation"`
	Args       map[string]json.RawMessage `json:"args"`
	WithAPIKey *string                    `json:"with_api_key"`
}

type expectedRequest struct {
	Method       string            `json:"method"`
	Path         *string           `json:"path"`
	Query        map[string]string `json:"query"`
	QueryString  *string           `json:"query_string"`
	QueryAbsent  []string          `json:"query_absent"`
	Headers      map[string]string `json:"headers"`
	HeadersAbs   []string          `json:"headers_absent"`
	HeadersMatch map[string]string `json:"headers_match"`
	URLPrefix    *string           `json:"url_prefix"`
}

type vectorResponse struct {
	Status   int               `json:"status"`
	Headers  map[string]string `json:"headers"`
	BodyFile string            `json:"body_file"`
	Body     *string           `json:"body"`
}

type vectorExpect struct {
	Result   map[string]json.RawMessage   `json:"result"`
	Results  []map[string]json.RawMessage `json:"results"`
	Error    *expectedError               `json:"error"`
	Requests *int                         `json:"requests"`
	DelaysS  *[]float64                   `json:"delays_s"`
	Items    []string                     `json:"items"`
	Warnings *[]struct {
		Contains string `json:"contains"`
	} `json:"warnings"`
}

type expectedError struct {
	Category       string          `json:"category"`
	Code           json.RawMessage `json:"code"`
	Status         *int            `json:"status"`
	RetryAfter     *float64        `json:"retry_after"`
	StringExcludes string          `json:"string_excludes"`
}

// mockTransport serves a case's responses in order and records requests.
type mockTransport struct {
	mu        sync.Mutex
	responses []vectorResponse
	requests  []*http.Request
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req.Clone(context.Background()))
	n := len(m.requests) - 1
	if n >= len(m.responses) {
		return nil, fmt.Errorf("mock: request %d has no scripted response", n+1)
	}
	r := m.responses[n]
	var body []byte
	switch {
	case r.Body != nil:
		body = []byte(*r.Body)
	case r.BodyFile != "":
		var err error
		if body, err = os.ReadFile(filepath.Join(vectorsDir, r.BodyFile)); err != nil {
			return nil, err
		}
	}
	h := http.Header{}
	for k, v := range r.Headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: r.Status, Status: strconv.Itoa(r.Status) + " " + http.StatusText(r.Status),
		Header: h, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body)),
		Request: req, Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
	}, nil
}

// logCapture records slog warnings.
type logCapture struct {
	mu   sync.Mutex
	msgs []string
}

func (l *logCapture) Enabled(context.Context, slog.Level) bool { return true }
func (l *logCapture) WithAttrs([]slog.Attr) slog.Handler       { return l }
func (l *logCapture) WithGroup(string) slog.Handler            { return l }
func (l *logCapture) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Message)
	r.Attrs(func(a slog.Attr) bool { fmt.Fprintf(&b, " %s=%v", a.Key, a.Value); return true })
	l.mu.Lock()
	l.msgs = append(l.msgs, b.String())
	l.mu.Unlock()
	return nil
}

// withTestSeams pins the clock, makes jitter its upper bound and records
// retry delays instead of sleeping.
func withTestSeams(delays *[]float64) Option {
	return func(c *config) {
		fixed := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
		c.now = func() time.Time { return fixed }
		c.jitter = func() float64 { return 1.0 }
		c.sleep = func(_ context.Context, d time.Duration) error {
			*delays = append(*delays, d.Seconds())
			return nil
		}
	}
}

func loadVectors(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(vectorsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".json") && !strings.Contains(path, string(filepath.Separator)+"bodies"+string(filepath.Separator)) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return files
}

// TestVectors runs every shared vector. The count is pinned so a vector that
// fails to load can never silently disappear.
func TestVectors(t *testing.T) {
	files := loadVectors(t)
	if len(files) != 43 {
		t.Errorf("found %d vectors, want 43", len(files))
	}
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var vc vectorCase
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&vc); err != nil {
			t.Errorf("%s: cannot parse vector: %v", file, err)
			continue
		}
		t.Run(vc.Name, func(t *testing.T) { runVector(t, vc) })
	}
}

func runVector(t *testing.T, vc vectorCase) {
	deprecationWarned.Range(func(k, _ any) bool { deprecationWarned.Delete(k); return true })
	t.Setenv(EnvAPIKey, "")
	t.Setenv(EnvBaseURL, "")
	for k, v := range vc.Client.Env {
		t.Setenv(k, v)
	}
	mock := &mockTransport{responses: vc.Responses}
	logs := &logCapture{}
	var delays []float64
	opts := []Option{
		WithHTTPClient(&http.Client{Transport: mock}),
		WithLogger(slog.New(logs)),
		withTestSeams(&delays),
	}
	if vc.Client.APIKey != nil {
		opts = append(opts, WithAPIKey(*vc.Client.APIKey))
	}
	if vc.Client.BaseURL != nil {
		opts = append(opts, WithBaseURL(*vc.Client.BaseURL))
	}
	if vc.Client.MaxRetries != nil {
		opts = append(opts, WithMaxRetries(*vc.Client.MaxRetries))
	}
	if vc.Client.Cache != nil && !*vc.Client.Cache {
		opts = append(opts, WithCache(nil))
	}
	if vc.Client.SendKeyOnOpenRoutes {
		opts = append(opts, WithKeyOnOpenRoutes())
	}
	client := NewClient(opts...)

	calls := vc.Calls
	if vc.Call != nil {
		calls = []vectorCall{*vc.Call}
	}
	var lastErr error
	for i, call := range calls {
		c := client
		if call.WithAPIKey != nil {
			c = client.With(WithAPIKey(*call.WithAPIKey))
		}
		result, items, err := runCall(context.Background(), c, call)
		if errors.Is(err, errUnmapped) {
			t.Fatalf("call %d: %v", i, err)
		}
		lastErr = err
		want := vc.Expect.Result
		if vc.Expect.Results != nil {
			want = vc.Expect.Results[i]
		}
		if want != nil {
			if err != nil {
				t.Fatalf("call %d: unexpected error: %v", i, err)
			}
			for path, exp := range want {
				checkPath(t, result, path, exp)
			}
		}
		if vc.Expect.Items != nil && !reflect.DeepEqual(items, vc.Expect.Items) {
			t.Errorf("items = %v, want %v", items, vc.Expect.Items)
		}
	}
	if vc.Expect.Error != nil {
		checkError(t, lastErr, *vc.Expect.Error)
	} else if lastErr != nil && vc.Expect.Result == nil && vc.Expect.Results == nil {
		t.Errorf("unexpected error: %v", lastErr)
	}

	reqs := mock.requests
	if vc.Expect.Requests != nil && len(reqs) != *vc.Expect.Requests {
		t.Errorf("made %d requests, want %d", len(reqs), *vc.Expect.Requests)
	}
	expected := vc.ExpectRequests
	if vc.ExpectRequest != nil {
		if string(*vc.ExpectRequest) == "null" {
			if len(reqs) != 0 {
				t.Errorf("made %d requests, want none", len(reqs))
			}
		} else {
			var er expectedRequest
			if err := json.Unmarshal(*vc.ExpectRequest, &er); err != nil {
				t.Fatal(err)
			}
			expected = []expectedRequest{er}
		}
	}
	for i, er := range expected {
		if i >= len(reqs) {
			t.Errorf("request %d was never sent", i+1)
			continue
		}
		checkRequest(t, i+1, reqs[i], er)
	}
	if vc.Expect.DelaysS != nil {
		if len(delays) != len(*vc.Expect.DelaysS) {
			t.Errorf("delays = %v, want %v", delays, *vc.Expect.DelaysS)
		} else {
			for i, d := range *vc.Expect.DelaysS {
				if math.Abs(delays[i]-d) > 1e-9 {
					t.Errorf("delay %d = %v s, want %v s", i, delays[i], d)
				}
			}
		}
	}
	if vc.Expect.Warnings != nil {
		ws := *vc.Expect.Warnings
		if len(logs.msgs) != len(ws) {
			t.Errorf("warnings = %q, want %d", logs.msgs, len(ws))
		}
		for i, w := range ws {
			if i < len(logs.msgs) && !strings.Contains(logs.msgs[i], w.Contains) {
				t.Errorf("warning %q does not contain %q", logs.msgs[i], w.Contains)
			}
		}
	}
}

func checkRequest(t *testing.T, n int, req *http.Request, er expectedRequest) {
	t.Helper()
	if er.Method != "" && req.Method != er.Method {
		t.Errorf("request %d: method %s, want %s", n, req.Method, er.Method)
	}
	if er.Path != nil && req.URL.EscapedPath() != *er.Path {
		t.Errorf("request %d: path %q, want %q", n, req.URL.EscapedPath(), *er.Path)
	}
	if er.QueryString != nil && req.URL.RawQuery != *er.QueryString {
		t.Errorf("request %d: query string\n got %q\nwant %q", n, req.URL.RawQuery, *er.QueryString)
	}
	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		t.Errorf("request %d: unparsable query %q: %v", n, req.URL.RawQuery, err)
	}
	for k, v := range er.Query {
		if got := q[k]; len(got) != 1 || got[0] != v {
			t.Errorf("request %d: query %s = %q, want %q", n, k, got, v)
		}
	}
	for _, k := range er.QueryAbsent {
		if _, ok := q[k]; ok {
			t.Errorf("request %d: query %s present, want absent", n, k)
		}
	}
	for k, v := range er.Headers {
		if got := req.Header.Get(k); got != v {
			t.Errorf("request %d: header %s = %q, want %q", n, k, got, v)
		}
	}
	for _, k := range er.HeadersAbs {
		if _, ok := req.Header[http.CanonicalHeaderKey(k)]; ok {
			t.Errorf("request %d: header %s present, want absent", n, k)
		}
	}
	for k, pattern := range er.HeadersMatch {
		if !regexp.MustCompile(pattern).MatchString(req.Header.Get(k)) {
			t.Errorf("request %d: header %s = %q, does not match %s", n, k, req.Header.Get(k), pattern)
		}
	}
	if er.URLPrefix != nil && !strings.HasPrefix(req.URL.String(), *er.URLPrefix) {
		t.Errorf("request %d: URL %q, want prefix %q", n, req.URL.String(), *er.URLPrefix)
	}
}

func checkError(t *testing.T, err error, want expectedError) {
	t.Helper()
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("error = %v, want an *Error of category %s", err, want.Category)
	}
	if e.Category.String() != want.Category {
		t.Errorf("category = %s, want %s", e.Category, want.Category)
	}
	if !errors.Is(err, sentinels[e.Category]) {
		t.Errorf("errors.Is(err, sentinel of %s) = false", e.Category)
	}
	if len(want.Code) > 0 {
		var code *string
		if err := json.Unmarshal(want.Code, &code); err != nil {
			t.Fatal(err)
		}
		switch {
		case code == nil && e.Code != "":
			t.Errorf("code = %q, want none", e.Code)
		case code != nil && string(e.Code) != *code:
			t.Errorf("code = %q, want %q", e.Code, *code)
		}
	}
	if want.Status != nil && e.Status != *want.Status {
		t.Errorf("status = %d, want %d", e.Status, *want.Status)
	}
	if want.RetryAfter != nil && e.RetryAfter.Seconds() != *want.RetryAfter {
		t.Errorf("retry_after = %v, want %vs", e.RetryAfter, *want.RetryAfter)
	}
	if s := want.StringExcludes; s != "" {
		for _, form := range []string{e.Error(), e.Message, fmt.Sprintf("%v", e), fmt.Sprintf("%+v", e), fmt.Sprintf("%#v", e)} {
			if strings.Contains(form, s) {
				t.Errorf("error string form contains %q: %s", s, form)
			}
		}
	}
}

var errUnmapped = errors.New("vector runner: unmapped")

// args decodes a call's arguments, refusing any it does not consume.
type args struct {
	raw  map[string]json.RawMessage
	used map[string]bool
	err  error
}

func (a *args) get(name string, v any) bool {
	raw, ok := a.raw[name]
	if !ok {
		return false
	}
	a.used[name] = true
	if err := json.Unmarshal(raw, v); err != nil && a.err == nil {
		a.err = fmt.Errorf("%w: arg %s: %w", errUnmapped, name, err)
	}
	return true
}

func (a *args) done() error {
	if a.err != nil {
		return a.err
	}
	for k := range a.raw {
		if !a.used[k] {
			return fmt.Errorf("%w: argument %q", errUnmapped, k)
		}
	}
	return nil
}

type vectorThreshold struct {
	Value float64 `json:"value"`
	Unit  Unit    `json:"unit"`
}

type vectorWindow struct {
	Hours      *int `json:"hours"`
	StartHours *int `json:"start_hours"`
	EndHours   *int `json:"end_hours"`
}

func (w *vectorWindow) window() Window {
	switch {
	case w == nil:
		return Window{}
	case w.Hours != nil:
		return Hours(*w.Hours)
	case w.StartHours != nil && w.EndHours != nil:
		return HoursBetween(*w.StartHours, *w.EndHours)
	}
	return Window{}
}

type vectorProb struct {
	Variable    Variable        `json:"variable"`
	Op          Op              `json:"op"`
	Threshold   vectorThreshold `json:"threshold"`
	Window      *vectorWindow   `json:"window"`
	Aggregation Aggregation     `json:"aggregation"`
}

type vectorTotal struct {
	Variable Variable      `json:"variable"`
	Window   *vectorWindow `json:"window"`
}

// runCall maps one vector call onto the SDK. It is the only per-language code.
func runCall(ctx context.Context, c *Client, call vectorCall) (any, []string, error) {
	a := &args{raw: call.Args, used: map[string]bool{}}
	location := func() string {
		var loc, state, slug string
		if a.get("location", &loc) {
			return loc
		}
		a.get("state", &state)
		a.get("slug", &slug)
		return state + "/" + slug
	}
	forecastParams := func() *ForecastParams {
		p := &ForecastParams{}
		var b bool
		a.get("units", &p.Units)
		a.get("max_hours", &p.MaxHours)
		a.get("max_days", &p.MaxDays)
		a.get("include_past", &p.IncludePast)
		a.get("members", &p.Members)
		a.get("models", &p.Models)
		if a.get("daily", &b) {
			p.Daily = Ptr(b)
		}
		if a.get("hourly", &b) {
			p.Hourly = Ptr(b)
		}
		if a.get("nowcast", &b) {
			p.Nowcast = Ptr(b)
		}
		var probs []vectorProb
		a.get("prob", &probs)
		for _, vp := range probs {
			p.Chances = append(p.Chances, Chance{
				Variable: vp.Variable, Op: vp.Op, Threshold: Threshold(vp.Threshold),
				Window: vp.Window.window(), Aggregation: vp.Aggregation,
			})
		}
		var totals []vectorTotal
		a.get("totals", &totals)
		for _, vt := range totals {
			p.Totals = append(p.Totals, Total{Variable: vt.Variable, Window: vt.Window.window()})
		}
		return p
	}
	point := func() Point {
		var p Point
		a.get("lat", &p.Lat)
		a.get("lon", &p.Lon)
		var e float64
		if a.get("elevation_m", &e) {
			p.ElevationM = Ptr(e)
		}
		return p
	}
	var iterate bool
	a.get("iterate", &iterate)

	switch call.Operation {
	case "getLocationForecast":
		loc, p := location(), forecastParams()
		if err := a.done(); err != nil {
			return nil, nil, err
		}
		r, err := c.Forecast(ctx, loc, p)
		return r, nil, err
	case "getPointForecast":
		pt, p := point(), forecastParams()
		if err := a.done(); err != nil {
			return nil, nil, err
		}
		r, err := c.GetPointForecast(ctx, pt, p)
		return r, nil, err
	case "searchLocations":
		var q string
		a.get("q", &q)
		if err := a.done(); err != nil {
			return nil, nil, err
		}
		r, err := c.Search(ctx, q, nil)
		return r, nil, err
	case "getCatalogLocation":
		var slug string
		a.get("slug", &slug)
		if err := a.done(); err != nil {
			return nil, nil, err
		}
		r, err := c.CatalogLocation(ctx, slug, nil)
		return r, nil, err
	case "getAttribution":
		if err := a.done(); err != nil {
			return nil, nil, err
		}
		r, err := c.Attribution(ctx)
		return r, nil, err
	case "listCatalogLocations":
		if err := a.done(); err != nil {
			return nil, nil, err
		}
		if iterate {
			var slugs []string
			err := collectCatalogLocations(ctx, c, func(l CatalogLocation) { slugs = append(slugs, l.Slug) })
			return nil, slugs, err
		}
		r, err := c.CatalogLocations(ctx, nil)
		return r, nil, err
	}
	return nil, nil, fmt.Errorf("%w: operation %q", errUnmapped, call.Operation)
}

// checkPath compares one expect.result path against the semantic model.
func checkPath(t *testing.T, result any, path string, raw json.RawMessage) {
	t.Helper()
	got, found := lookup(reflect.ValueOf(result), path)
	var want any
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if want == nil {
		if found && !isAbsent(got) {
			t.Errorf("%s = %v, want absent", path, got.Interface())
		}
		return
	}
	if !found || isAbsent(got) {
		t.Errorf("%s is absent, want %v", path, want)
		return
	}
	got = deref(got)
	switch w := want.(type) {
	case map[string]any:
		rawStr, _ := w["unknown"].(string)
		k, ok := got.Interface().(interface{ IsKnown() bool })
		if !ok || got.Kind() != reflect.String {
			t.Errorf("%s: %T has no unknown-enum case", path, got.Interface())
			return
		}
		if k.IsKnown() || got.String() != rawStr {
			t.Errorf("%s = %q (known=%v), want the unknown case keeping %q", path, got.String(), k.IsKnown(), rawStr)
		}
	case float64:
		var g float64
		switch got.Kind() {
		case reflect.Float32, reflect.Float64:
			g = got.Float()
		case reflect.Int, reflect.Int64, reflect.Int32:
			g = float64(got.Int())
		default:
			t.Errorf("%s: %v is not a number", path, got.Interface())
			return
		}
		if g != w {
			t.Errorf("%s = %v, want %v", path, g, w)
		}
	case string:
		if got.Kind() != reflect.String || got.String() != w {
			t.Errorf("%s = %v, want %q", path, got.Interface(), w)
		}
	case bool:
		if got.Kind() != reflect.Bool || got.Bool() != w {
			t.Errorf("%s = %v, want %v", path, got.Interface(), w)
		}
	default:
		t.Errorf("%s: unsupported expectation %v", path, want)
	}
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}
	return v
}

func isAbsent(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	}
	return false
}

var segmentRE = regexp.MustCompile(`^([a-z0-9_]+)((?:\[\d+\])*)$`)

func lookup(v reflect.Value, path string) (reflect.Value, bool) {
	for _, seg := range strings.Split(path, ".") {
		m := segmentRE.FindStringSubmatch(seg)
		if m == nil {
			return reflect.Value{}, false
		}
		v = deref(v)
		if v.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		f, ok := field(v, m[1])
		if !ok {
			return reflect.Value{}, false
		}
		v = f
		for _, idx := range regexp.MustCompile(`\d+`).FindAllString(m[2], -1) {
			i, _ := strconv.Atoi(idx)
			v = deref(v)
			if v.Kind() != reflect.Slice || i >= v.Len() {
				return reflect.Value{}, false
			}
			v = v.Index(i)
		}
	}
	return v, true
}

// field finds a struct field by its JSON name (or its snake_case Go name for
// untagged fields such as ResponseMeta's), searching embedded structs.
func field(v reflect.Value, name string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.Anonymous {
			if f, ok := field(v.Field(i), name); ok {
				return f, true
			}
			continue
		}
		tag, _, _ := strings.Cut(sf.Tag.Get("json"), ",")
		if tag == "" || tag == "-" {
			tag = snake(sf.Name)
		}
		if tag == name {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if 'A' <= r && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}
