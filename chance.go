// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// Variable is a forecast variable a probability or window total can ask about.
type Variable string

// The variables of probability terms ([Chance]) and window totals ([Total]).
const (
	Temperature   Variable = "temperature"
	WindSpeed     Variable = "wind_speed"
	SnowLevel     Variable = "snow_level"
	Precipitation Variable = "precipitation"
	Snowfall      Variable = "snowfall"
)

// IsKnown reports whether v is a variable this SDK knows.
func (v Variable) IsKnown() bool {
	switch v {
	case Temperature, WindSpeed, SnowLevel, Precipitation, Snowfall:
		return true
	}
	return false
}

func (v Variable) isAmount() bool { return v == Precipitation || v == Snowfall }

// Op is a probability comparison.
type Op string

// The comparisons.
const (
	// AtLeast asks for the chance the value is at least the threshold (ge).
	AtLeast Op = "ge"
	// AtMost asks for the chance the value is at most the threshold (le).
	AtMost Op = "le"
)

// IsKnown reports whether o is a comparison this SDK knows.
func (o Op) IsKnown() bool { return o == AtLeast || o == AtMost }

// Aggregation says how a probability term or a window total combines the
// hours of its window.
type Aggregation string

// The aggregations. In a [Chance], precipitation and snowfall take only
// [AggregationTotal]; the other variables take [AggregationAny] (their
// default) or [AggregationAll]. [AggregationMin] and [AggregationMax] appear
// only in a [WindowTotal] the server answered.
const (
	// AggregationTotal: the member's total over the window crosses the threshold.
	AggregationTotal Aggregation = "total"
	// AggregationAny: any hour of the window crosses the threshold.
	AggregationAny Aggregation = "any"
	// AggregationAll: every hour of the window crosses the threshold.
	AggregationAll Aggregation = "all"
	// AggregationMin: each member's lowest value over the window.
	AggregationMin Aggregation = "min"
	// AggregationMax: each member's highest value over the window.
	AggregationMax Aggregation = "max"
)

// IsKnown reports whether a is an aggregation this SDK knows.
func (a Aggregation) IsKnown() bool {
	switch a {
	case AggregationTotal, AggregationAny, AggregationAll, AggregationMin, AggregationMax:
		return true
	}
	return false
}

var aggregationOrder = map[Aggregation]int{AggregationTotal: 0, AggregationAny: 1, AggregationAll: 2}

// MaxWindowHours is the longest window the API answers.
const MaxWindowHours = 360

// Window is a span of forecast hours counted from the first served hour:
// hours Start..End. The zero Window means the API default, the first 24 h.
type Window struct {
	Start, End int
}

// Hours returns the window of the first n hours (0..n).
func Hours(n int) Window { return Window{0, n} }

// HoursBetween returns the window from hour start to hour end.
func HoursBetween(start, end int) Window { return Window{start, end} }

func (w Window) orDefault() Window {
	if w == (Window{}) {
		return Hours(24)
	}
	return w
}

func (w Window) valid() bool {
	return w.Start >= 0 && w.End > w.Start && w.End <= MaxWindowHours
}

// String returns the window in the API's syntax: "24h" or "12-36h".
func (w Window) String() string {
	w = w.orDefault()
	if w.Start == 0 {
		return strconv.Itoa(w.End) + "h"
	}
	return strconv.Itoa(w.Start) + "-" + strconv.Itoa(w.End) + "h"
}

// Chance is one probability term: the chance that Variable is at least (or at
// most) Threshold over Window. The server computes it from every ensemble
// member and answers it in [ForecastResponse.Probabilities], in the same
// order as [ForecastParams.Chances].
type Chance struct {
	Variable  Variable
	Op        Op
	Threshold Threshold
	// Window defaults to the first 24 h.
	Window Window
	// Aggregation defaults to total for precipitation and snowfall and to
	// any for the other variables.
	Aggregation Aggregation
}

// Total is one window total: the distribution of each ensemble member's total
// precipitation or snowfall over Window, answered in
// [ForecastResponse.WindowTotals] in the same order as [ForecastParams.Totals].
// This is the only storm total the SDK offers: it never sums hourly values.
type Total struct {
	Variable Variable
	// Window defaults to the first 24 h.
	Window Window
}

// precision is the number of decimals the server serves for v in units u.
func precision(v Variable, u Units) int {
	switch v {
	case Precipitation, Snowfall:
		if u == UnitsImperial {
			return 3
		}
		return 2
	case SnowLevel:
		return 0
	}
	return 1
}

// roundThreshold rounds v exactly as the server does: half away from zero on
// v×10^d in IEEE-754 doubles, with -0 folded to 0.
func roundThreshold(v float64, d int) float64 {
	p := math.Pow10(d)
	r := math.Round(v*p) / p
	if r == 0 {
		r = 0 // fold -0
	}
	return r
}

func formatThreshold(v float64, d int) string {
	return strconv.FormatFloat(roundThreshold(v, d), 'f', d, 64)
}

// chanceTerm is a validated, canonical term.
type chanceTerm struct {
	variable    Variable
	op          Op
	value       float64 // rounded, in request units
	text        string  // value printed with the served precision
	window      Window
	aggregation Aggregation // effective (never empty)
}

// wire is the term's canonical spelling: what the SDK sends and the server
// echoes in ProbabilityResult.Term.
func (t chanceTerm) wire() string {
	s := string(t.variable) + ":" + string(t.op) + ":" + t.text + ":" + t.window.String()
	if t.aggregation != defaultAggregation(t.variable) {
		s += ":" + string(t.aggregation)
	}
	return s
}

func defaultAggregation(v Variable) Aggregation {
	if v.isAmount() {
		return AggregationTotal
	}
	return AggregationAny
}

// prepareChance validates c and converts its threshold to the request units.
func prepareChance(op string, c Chance, units Units) (chanceTerm, error) {
	if !c.Variable.IsKnown() {
		return chanceTerm{}, usageError(op, "chance: unknown variable %q", c.Variable)
	}
	if !c.Op.IsKnown() {
		return chanceTerm{}, usageError(op, "chance: unknown op %q (use AtLeast or AtMost)", c.Op)
	}
	w := c.Window.orDefault()
	if !w.valid() {
		return chanceTerm{}, usageError(op, "chance: window %d-%dh must satisfy 0 <= start < end <= %d", w.Start, w.End, MaxWindowHours)
	}
	agg := c.Aggregation
	if agg == "" {
		agg = defaultAggregation(c.Variable)
	}
	switch {
	case agg != AggregationTotal && agg != AggregationAny && agg != AggregationAll:
		return chanceTerm{}, usageError(op, "chance: unknown aggregation %q", agg)
	case c.Variable.isAmount() && agg != AggregationTotal:
		return chanceTerm{}, usageError(op, "chance: %s takes only the total aggregation, not %q", c.Variable, agg)
	case !c.Variable.isAmount() && agg == AggregationTotal:
		return chanceTerm{}, usageError(op, "chance: %s takes any or all, not total", c.Variable)
	}
	v := c.Threshold.Value
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return chanceTerm{}, usageError(op, "chance: threshold must be a finite number")
	}
	if c.Threshold.Unit != "" {
		target := units.For(c.Variable)
		if c.Threshold.Unit.dimension() != target.dimension() {
			return chanceTerm{}, usageError(op, "chance: a %s threshold cannot be given in %q", c.Variable, c.Threshold.Unit)
		}
		v = convert(v, c.Threshold.Unit, target)
	}
	d := precision(c.Variable, units)
	return chanceTerm{
		variable:    c.Variable,
		op:          c.Op,
		value:       roundThreshold(v, d),
		text:        formatThreshold(v, d),
		window:      w,
		aggregation: agg,
	}, nil
}

// chanceQuery is the canonical prob= value plus the canonical term of each
// caller term, in the caller's order.
type chanceQuery struct {
	value string
	keys  []string
}

func buildChances(op string, cs []Chance, units Units) (chanceQuery, error) {
	if len(cs) == 0 {
		return chanceQuery{}, nil
	}
	keys := make([]string, len(cs))
	uniq := map[string]chanceTerm{}
	for i, c := range cs {
		t, err := prepareChance(op, c, units)
		if err != nil {
			return chanceQuery{}, err
		}
		keys[i] = t.wire()
		uniq[keys[i]] = t
	}
	terms := make([]chanceTerm, 0, len(uniq))
	for _, t := range uniq {
		terms = append(terms, t)
	}
	sort.Slice(terms, func(i, j int) bool {
		a, b := terms[i], terms[j]
		switch {
		case a.variable != b.variable:
			return a.variable < b.variable
		case a.op != b.op:
			return a.op < b.op
		case a.window.Start != b.window.Start:
			return a.window.Start < b.window.Start
		case a.window.End != b.window.End:
			return a.window.End < b.window.End
		case a.aggregation != b.aggregation:
			return aggregationOrder[a.aggregation] < aggregationOrder[b.aggregation]
		}
		return a.value < b.value
	})
	wire := make([]string, len(terms))
	for i, t := range terms {
		wire[i] = t.wire()
	}
	return chanceQuery{value: strings.Join(wire, ","), keys: keys}, nil
}

// mapChances matches the server's answers to the caller's terms by their
// echoed canonical term (never by position), in the caller's order,
// duplicates sharing one answer.
//
// A response without any answers leaves Probabilities absent (nil), never
// filled in; a response that answers some terms but not all is an error.
func mapChances(op string, q chanceQuery, results []ProbabilityResult, units Units) ([]ProbabilityResult, error) {
	if len(q.keys) == 0 || len(results) == 0 {
		return results, nil
	}
	byKey := make(map[string]ProbabilityResult, len(results))
	for _, r := range results {
		r.ThresholdUnit = units.For(r.Variable)
		byKey[r.Term] = r
	}
	out := make([]ProbabilityResult, len(q.keys))
	for i, k := range q.keys {
		r, ok := byKey[k]
		if !ok {
			return nil, &Error{Category: CategoryUnavailable, Operation: op, Status: 200,
				Message: "the response has no answer for probability term " + k}
		}
		out[i] = r
	}
	return out, nil
}

type totalTerm struct {
	variable Variable
	window   Window
}

func (t totalTerm) key() string { return string(t.variable) + ":" + t.window.String() }

type totalsQuery struct {
	value string
	keys  []string
}

func buildTotals(op string, ts []Total) (totalsQuery, error) {
	if len(ts) == 0 {
		return totalsQuery{}, nil
	}
	keys := make([]string, len(ts))
	uniq := map[string]totalTerm{}
	for i, t := range ts {
		if !t.Variable.isAmount() {
			return totalsQuery{}, usageError(op, "total: variable must be precipitation or snowfall, not %q", t.Variable)
		}
		w := t.Window.orDefault()
		if !w.valid() {
			return totalsQuery{}, usageError(op, "total: window %d-%dh must satisfy 0 <= start < end <= %d", w.Start, w.End, MaxWindowHours)
		}
		tt := totalTerm{t.Variable, w}
		keys[i] = tt.key()
		uniq[keys[i]] = tt
	}
	terms := make([]totalTerm, 0, len(uniq))
	for _, t := range uniq {
		terms = append(terms, t)
	}
	sort.Slice(terms, func(i, j int) bool {
		a, b := terms[i], terms[j]
		switch {
		case a.variable != b.variable:
			return a.variable < b.variable
		case a.window.Start != b.window.Start:
			return a.window.Start < b.window.Start
		}
		return a.window.End < b.window.End
	})
	wire := make([]string, len(terms))
	for i, t := range terms {
		wire[i] = t.key()
	}
	return totalsQuery{value: strings.Join(wire, ","), keys: keys}, nil
}

// mapTotals matches window totals to the caller's terms by their echoed
// canonical term, as mapChances.
func mapTotals(op string, q totalsQuery, results []WindowTotal, units Units) ([]WindowTotal, error) {
	if len(q.keys) == 0 || len(results) == 0 {
		return results, nil
	}
	byKey := make(map[string]WindowTotal, len(results))
	for _, r := range results {
		r.Unit = units.For(r.Variable)
		byKey[r.Term] = r
	}
	out := make([]WindowTotal, len(q.keys))
	for i, k := range q.keys {
		r, ok := byKey[k]
		if !ok {
			return nil, &Error{Category: CategoryUnavailable, Operation: op, Status: 200,
				Message: "the response has no answer for window total " + k}
		}
		out[i] = r
	}
	return out, nil
}
