// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestUnits(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	if !strings.Contains(out, "metric   temperature median -2.0 °C") || !strings.Contains(out, "imperial temperature median 28.4 °F") {
		t.Errorf("unexpected output:\n%s", out)
	}
}
