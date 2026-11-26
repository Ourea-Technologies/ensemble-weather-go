// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// escape percent-encodes s with RFC 3986 rules: unreserved characters
// (A-Z a-z 0-9 - . _ ~) stay as they are and every other byte becomes %XX in
// uppercase hex (space is %20, never +). Bytes in keep are also left literal.
//
// net/url is not used on purpose: url.Values.Encode writes space as '+' and
// url.PathEscape leaves '&', '=' and others unescaped, and neither gives the
// one canonical spelling the edge cache needs (BEHAVIOR 3.1).
func escape(s, keep string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if unreserved(c) || (keep != "" && strings.IndexByte(keep, c) >= 0) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&15])
	}
	return b.String()
}

func unreserved(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

// formatNumber prints the shortest decimal that round-trips, never an exponent.
func formatNumber(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

type param struct {
	name, value string
	keep        string   // bytes sent literally (":," for prob and totals)
	items       []string // a comma list: each item encoded, the commas literal
}

// query builds a canonical query string: each parameter at most once, none
// at its default, sorted by name.
type query struct {
	params []param
}

func (q *query) set(name, value string) {
	q.params = append(q.params, param{name: name, value: value})
}

func (q *query) str(name, value string) {
	if value != "" {
		q.set(name, value)
	}
}

// strDefault sets a string unless it is empty or the route's default.
func (q *query) strDefault(name, value, def string) {
	if value != "" && value != def {
		q.set(name, value)
	}
}

// list sets a comma list in the server's canonical form (BEHAVIOR 3.1 rule
// 5) unless it is empty: sorted, de-duplicated, each item percent-encoded
// and the commas between items literal (%2C would be another edge-cache
// entry).
func (q *query) list(name string, values []string) {
	if len(values) == 0 {
		return
	}
	vs := append([]string(nil), values...)
	sort.Strings(vs)
	q.params = append(q.params, param{name: name, items: slices.Compact(vs)})
}

// numbers sets a fixed-order tuple of numbers, such as near (lat,lon) or
// bbox (west,south,east,north): each number by the shortest-decimal rule,
// joined by literal commas.
func (q *query) numbers(name string, vs ...float64) {
	items := make([]string, len(vs))
	for i, v := range vs {
		items[i] = formatNumber(v)
	}
	q.params = append(q.params, param{name: name, items: items})
}

// intDefault sets an integer unless it is 0 (unset) or the route's default.
func (q *query) intDefault(name string, v, def int) {
	if v != 0 && v != def {
		q.set(name, strconv.Itoa(v))
	}
}

func (q *query) num(name string, v float64) { q.set(name, formatNumber(v)) }

// numDefault sets a number unless it is 0 (unset) or the route's default.
func (q *query) numDefault(name string, v, def float64) {
	if v != 0 && v != def {
		q.num(name, v)
	}
}

// near sets near=lat,lon (literal comma) from a point, refusing an
// elevation or an out-of-range coordinate.
func (q *query) near(op string, p *Point) error {
	if p == nil {
		return nil
	}
	if p.ElevationM != nil {
		return usageError(op, "near takes no elevation")
	}
	if !checkLatLon(p.Lat, p.Lon) {
		return usageError(op, "near (%v, %v) is outside -90..90 latitude, -180..180 longitude", p.Lat, p.Lon)
	}
	q.numbers("near", p.Lat, p.Lon)
	return nil
}

func (q *query) numPtr(name string, v *float64) {
	if v != nil {
		q.num(name, *v)
	}
}

// flag sets a boolean whose default is false.
func (q *query) flag(name string, v bool) {
	if v {
		q.set(name, "true")
	}
}

// flagDefault sets an optional boolean unless it is nil or the default.
func (q *query) flagDefault(name string, v *bool, def bool) {
	if v != nil && *v != def {
		q.set(name, strconv.FormatBool(*v))
	}
}

func (q *query) time(name string, t time.Time) {
	if !t.IsZero() {
		q.set(name, t.UTC().Format(time.RFC3339))
	}
}

// literal sets a value whose ':' and ',' are sent as they are (prob, totals).
func (q *query) literal(name, value string) {
	if value != "" {
		q.params = append(q.params, param{name: name, value: value, keep: ":,"})
	}
}

// units sets units= unless it is the API default (metric).
func (q *query) units(u Units) {
	q.strDefault("units", string(u), string(UnitsMetric))
}

func (q *query) encode() string {
	ps := append([]param(nil), q.params...)
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].name < ps[j].name })
	var b strings.Builder
	for i, p := range ps {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(escape(p.name, ""))
		b.WriteByte('=')
		if p.items != nil {
			for j, it := range p.items {
				if j > 0 {
					b.WriteByte(',')
				}
				b.WriteString(escape(it, ""))
			}
			continue
		}
		b.WriteString(escape(p.value, p.keep))
	}
	return b.String()
}
