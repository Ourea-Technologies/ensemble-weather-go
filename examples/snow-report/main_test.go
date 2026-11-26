// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestSnowReportAnonymous(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	for _, want := range []string{"Most snow in the next 5 days (in):", "1. Brundage Mountain, id: 11.4", "Water year 2027 at Mt Hood Meadows, through 2026-11-23:", "precipitation 210.69 mm, 41% of normal", "part of Plus and Pro"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestSnowReportWithKey(t *testing.T) {
	out := stub.RunMain(t, main, map[string]string{"OUREA_API_KEY": "ow_plus"})
	if want := "Snow trend (ourea), newest run Fri 06:00Z:\n  2026-11-20: none"; !strings.Contains(out, want) {
		t.Errorf("missing %q in:\n%s", want, out)
	}
}
