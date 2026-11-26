# Contributing to ensemble-weather-go

Thank you for helping. This is the official Go SDK for the
[Ourea](https://oureawx.com) API; its behaviour is fixed by
[ensemble-weather-spec](https://github.com/Ourea-Technologies/ensemble-weather-spec)
(pinned in [`spec/`](spec/)), and its shape follows
[Effective Go](https://go.dev/doc/effective_go) and the
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).

## Setup

You need Go 1.22 or later (CI also tests the two latest releases). Nothing
else: the SDK uses only the standard library, and the linters run through
`go run` at pinned versions.

```sh
git clone https://github.com/Ourea-Technologies/ensemble-weather-go
cd ensemble-weather-go
scripts/check        # every gate CI runs
```

## The gates

`scripts/check` runs these in order and stops at the first failure. CI runs
the same commands.

| Gate | Command |
|---|---|
| format | `gofmt -l .` (must print nothing) |
| vet | `go vet ./...` (and in `spectest/`) |
| staticcheck | `go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 -checks all ./...` |
| golangci-lint | `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...` (standard linters + revive, gosec, errorlint, misspell; see `.golangci.yml`) |
| vulnerabilities | `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...` |
| dependencies | `scripts/check-deps` (see below) |
| unit tests, vectors, examples | `go test -race ./...` |
| spec drift | `cd spectest && go test ./...` |
| docs | `go doc -all .` |
| build | `go build ./...` |

- **Vectors.** `vectors_test.go` runs every case in `spec/vectors` through a
  mock `http.RoundTripper` with a fake clock and a jitter source pinned to
  its upper bound. A case the runner cannot map is a failure, never a skip,
  and the count (43) is pinned.
- **Examples.** Each `examples/<name>/main.go` has a `main_test.go` that runs
  `main()` against `internal/stub`, a local stand-in for the API, with
  `OUREA_BASE_URL` pointing at it. To try an example by hand against the stub
  or a proxy, set `OUREA_BASE_URL`. Never point tests or examples at
  `api.oureawx.com`.
- **Spec drift.** `spectest/` is a separate module (so its YAML parser never
  enters the SDK's `go.mod`). It fails when the spec has an operation the SDK
  does not expose or a schema property the models lack, and it round-trips
  every response example.
- **Go 1.22.** Code outside `iter.go` must build on Go 1.22: no standard
  library API newer than 1.22 (CI's 1.22 job catches it).

## Updating the spec

```sh
scripts/sync-spec --ref <tag>   # rewrites spec/ and spec/VERSION
scripts/sync-spec --check       # what CI runs
```

`OUREA_SPEC_REPO=/path/to/ensemble-weather-spec` syncs from a local clone.
A spec change that changes behaviour arrives with new vectors; make them pass
rather than editing them.

## Dependencies

The SDK has **zero** runtime dependencies: `net/http`, `encoding/json` and
`log/slog` from the standard library cover everything. `scripts/check-deps`
fails CI if `go.mod` requires any module that is not listed below with its
justification, and a new entry needs Harrison's approval.

### Runtime dependencies

(none)

### Test-only dependencies

These live in `spectest/go.mod`, a separate module the SDK never imports:

- `go.yaml.in/yaml/v3`: parses `spec/openapi/v1/openapi.yaml` for the
  spec-drift tests (the standard library has no YAML parser).

Linters and `govulncheck` run through `go run <module>@<version>` and are
not dependencies of either module.

## Commits and pull requests

- One logical change per commit; an imperative summary line (`Add`, `Fix`,
  ...) under about 72 characters, then a short body saying what and why.
- Every behaviour change updates the README or docs, adds a test (a red-first
  regression test for a fix), and a line under `[Unreleased]` in
  `CHANGELOG.md`.
- `main` is protected: CI must pass and history stays linear (rebase, no
  merge commits).

## Releasing

A tag is the release: Go has no package registry upload. The module is
fetched from this repository through the vanity path
`oureawx.com/go/ensemble-weather`, and pkg.go.dev lists a version once the
module proxy has fetched it.

1. **Before the first release:**
   - The vanity import path must resolve.
     `https://oureawx.com/go/ensemble-weather?go-get=1` must serve
     `<meta name="go-import" content="oureawx.com/go/ensemble-weather git https://github.com/Ourea-Technologies/ensemble-weather-go">`
     (and a matching `go-source` tag). Check with
     `curl -s 'https://oureawx.com/go/ensemble-weather?go-get=1' | grep go-import`.
     Until it does, `go get` fails and the proxy warm-up in `release.yml` fails.
   - Create the `release` environment (Settings, Environments): a required
     reviewer, and a deployment rule for tags `v*` only. `release.yml`'s
     GitHub-release job runs in it. (A run that names a missing environment
     creates it unprotected, so create it first.)
2. Date the version's section in `CHANGELOG.md` (`## [X.Y.Z] - YYYY-MM-DD`;
   `release.yml` refuses `unreleased`) and check that `Version` in
   `version.go` is `X.Y.Z`; merge to `main`.
3. Tag `main`: `git tag -s vX.Y.Z -m vX.Y.Z && git push origin vX.Y.Z`.
   A tag is permanent: the module proxy and the checksum database keep
   what they first fetched, so a moved or deleted tag cannot be re-used;
   fix forward with the next patch version.
4. `release.yml` re-runs the gates, checks the tag against `version.go` and
   the changelog date, creates the GitHub release from the changelog
   section with a source archive and its build-provenance attestation
   (after the environment's approval), and warms the module proxy
   (`GOPROXY=proxy.golang.org go list -m oureawx.com/go/ensemble-weather@vX.Y.Z`).
5. Check pkg.go.dev: https://pkg.go.dev/oureawx.com/go/ensemble-weather@vX.Y.Z
   should show the version and the BSD-2-Clause licence within minutes. If
   it does not, request it there ("Request" on the not-found page) or run
   `GOPROXY=https://proxy.golang.org GO111MODULE=on go get oureawx.com/go/ensemble-weather@vX.Y.Z`
   in any module.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
