// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestPointForecast(t *testing.T) {
	if out := stub.RunMain(t, main, nil); !strings.Contains(out, "set OUREA_API_KEY") {
		t.Errorf("without a key: %q", out)
	}
	out := stub.RunMain(t, main, map[string]string{"OUREA_API_KEY": "ow_example"})
	if !strings.Contains(out, "wind median") || !strings.Contains(out, "rate limit: 599 of 600 left") {
		t.Errorf("with a key:\n%s", out)
	}
}
