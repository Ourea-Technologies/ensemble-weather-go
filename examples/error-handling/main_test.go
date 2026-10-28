// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"strings"
	"testing"

	"oureawx.com/go/ensemble-weather/internal/stub"
)

func TestErrorHandling(t *testing.T) {
	out := stub.RunMain(t, main, nil)
	if !strings.Contains(out, "need a Pro key") || !strings.Contains(out, "no rate-limit headers") {
		t.Errorf("anonymous:\n%s", out)
	}
	out = stub.RunMain(t, main, map[string]string{"OUREA_API_KEY": "ow_example"})
	if !strings.Contains(out, "outside coverage: 45,10 is outside") ||
		!strings.Contains(out, "status 404, request stub-request; see https://docs.oureawx.com/features/errors/#outside_coverage") || !strings.Contains(out, "rate limit: 599 of 600 left") {
		t.Errorf("with a key:\n%s", out)
	}
}
