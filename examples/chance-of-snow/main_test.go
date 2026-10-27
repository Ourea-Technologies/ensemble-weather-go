// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestChanceOfSnow(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	for _, want := range []string{"snowfall ge 6 in, hours 0-24", "snowfall ge 12 in, hours 0-72", "snowfall ge 6 in, hours 24-48", "temperature le 32 °F, hours 0-12 (all)"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}
