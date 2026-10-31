// SPDX-License-Identifier: BSD-2-Clause
//
// The spec-drift tests live in their own module so that their YAML parser
// (a test-only dependency) never appears in the SDK's go.mod.
module oureawx.com/go/ensemble-weather/spectest

go 1.22

require oureawx.com/go/ensemble-weather v0.0.0

require go.yaml.in/yaml/v3 v3.0.5

replace oureawx.com/go/ensemble-weather => ../
