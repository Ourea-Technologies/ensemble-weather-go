// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestStormTotal(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	for _, want := range []string{"snowfall, hours 0-72: median 29 mm", "precipitation, hours 0-72", "snowfall, hours 24-48"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}
