# Security policy

## Reporting a vulnerability

Please do **not** open a public issue, discussion or pull request for a
security problem.

Report it privately through GitHub's
[private vulnerability reporting](https://github.com/Ourea-Technologies/ensemble-weather-go/security/advisories/new)
("Report a vulnerability" on the Security tab). Include what you found, how to
reproduce it, and the SDK version (`ensembleweather.Version`).

We acknowledge reports within three working days, keep you updated while we
fix the problem, and credit you in the advisory unless you prefer otherwise.

## Supported versions

Security fixes go to the latest release. Before 1.0, that is the latest 0.x
minor version.

## What the SDK guarantees

- The API key is sent only in the `X-API-Key` header, only to the configured
  base URL. A redirect to any other host (or scheme) is followed without it.
- The key never appears in errors, log output or `fmt` formatting of the
  client.
- TLS verification is never disabled by the SDK, and it has no option to
  disable it.
- Response bodies are bounded: a body over 64 MiB is an error, not an
  unbounded allocation.
- Responses are decoded with `encoding/json` into fixed types only.

Vulnerabilities in the Ourea API itself can be reported the same way.
