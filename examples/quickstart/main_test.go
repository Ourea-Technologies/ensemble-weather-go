// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestQuickstart(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	if !strings.Contains(out, "°C") || !strings.Contains(out, "Chance of >= 152.4 mm of snow in 0-24 h: 42%") {
		t.Errorf("unexpected output:\n%s", out)
	}
}
