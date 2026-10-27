// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestCurrentAndAlerts(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	for _, want := range []string{"Mt. Hood Meadows at 00:00 UTC: Snow, -2.0 °C (from nowcast)", "1 active alert(s)", "Winter Storm Warning [Severe]"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}
