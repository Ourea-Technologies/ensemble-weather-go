// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"os"
	"strings"
	"testing"
)

// TestReadmeQuickstartIsTheExample keeps the README's quickstart identical to
// examples/quickstart/main.go (from its package clause on), which CI runs.
func TestReadmeQuickstartIsTheExample(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	example, err := os.ReadFile("examples/quickstart/main.go")
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(readme), "## Quickstart")
	if !ok {
		t.Fatal("README has no Quickstart section")
	}
	_, block, ok := strings.Cut(section, "```go\n")
	if !ok {
		t.Fatal("README Quickstart has no Go block")
	}
	block, _, _ = strings.Cut(block, "```")
	i := strings.Index(string(example), "package main")
	if i < 0 {
		t.Fatal("example has no package clause")
	}
	if block != string(example[i:]) {
		t.Errorf("README quickstart differs from examples/quickstart/main.go:\n--- README\n%s\n--- example\n%s", block, example[i:])
	}
}
