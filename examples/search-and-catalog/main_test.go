// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestSearchAndCatalog(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	for _, want := range []string{"found Mt. Hood Meadows (or/mt-hood-meadows, Hood River County)", "page 1: Mt. Hood Meadows", "page 2: Timberline Lodge"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}
