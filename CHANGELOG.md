# Changelog

Notable changes to this library, newest first. Versions are git tags; this file is written
for whoever bumps the dependency — what changed, and what it means for code that already
uses it.

## v1.1.5

Dependency maintenance with one thing to act on: **this library now needs Go 1.27**. No source
changed here and nothing it does behaves differently — but the platform kit moves two patch
releases in one step, and one upstream behaviour change rides along with it. See *Notes*.

### Changed

- **The module declares `go 1.27.0`** (was `1.26.6`), so your own module has to be on Go 1.27
  before it can build against this one. A dependency's `go` line does **not** make the go command
  fetch a newer toolchain for you — measured both ways: a consumer whose own `go` directive is
  lower stops with a `requires go >= 1.27.0 (running go 1.26.6)` error, and it stops there with
  `GOTOOLCHAIN` on its `auto` default just as it does under `local`. Raise your own `go` directive
  to `1.27.0` first; from there the go command downloads and uses the 1.27 toolchain by itself, so
  nobody has to install Go by hand. CI that reads `go-version-file: go.mod` follows the bump with
  no workflow edit — a workflow naming a Go version in the YAML needs that line changed.

### Notes

- **`github.com/gmb-lib/go-platform-kit` → v1.11.3** (was v1.11.1), and with it the framework:
  `azugo.io/azugo`, `azugo.io/core` and `azugo.io/opentelemetry` → **v0.38.1**,
  `github.com/valyala/fasthttp` → **v1.74.0**. Nothing this library takes from any of them moved —
  it has no metrics code and touches fasthttp only in a test.

  **One thing in that framework release is visible to your monitoring, not to your code, and it
  arrives with nothing to opt into.** From azugo v0.38.1 the metrics endpoint no longer negotiates
  OpenMetrics: a scraper sending `Accept: application/openmetrics-text` is answered
  `Content-Type: text/plain; version=0.0.4; charset=utf-8` with no `# EOF` terminator, where it used
  to get the OpenMetrics format. **The metric names, labels and values are unchanged.** It reaches a
  service through the platform kit, which binds azugo's metrics configuration — so if your scrape
  configuration demands the OpenMetrics content type, or treats a missing `# EOF` as a truncated
  scrape, **check it before you deploy**.

- **One module leaves the dependency graph and another joins it**, both transitively through
  fasthttp: `github.com/andybalholm/brotli` is gone and `github.com/molecule-man/go-brrr` v1.1.0
  provides the brotli implementation in its place. Also moved indirectly:
  `go-playground/validator/v10` → v10.30.4, `klauspost/compress` → v1.20.0, `golang.org/x/crypto` →
  v0.57.0, `x/net` → v0.59.0, `x/sys` → v0.48.0, `x/text` → v0.42.0, and the
  `google.golang.org/genproto/googleapis/{api,rpc}` snapshots → 20260911204522. Nothing in this
  library calls any of them directly.

- The repository gained a code of conduct, and the advisory DCO workflow was removed now that the
  sign-off is enforced by the organisation's app together with a branch ruleset. What a contribution
  has to carry is unchanged.

- The gate is green on the new set: `go mod verify`, `go mod tidy -diff`, build, vet, `gofmt`, and
  `go test -race` with **0 races**; `govulncheck` finds nothing.

## v1.1.4

### Changed

- **`azugo.io/azugo` and `azugo.io/core` → v0.38.0, `github.com/gmb-lib/go-platform-kit` →
  v1.10.0.** No source change here: the platform-kit release is additive — a size cap on a
  JetStream stream, which this library does not configure — and nothing else reaches this code.

  One thing in the framework release is worth knowing if you use azugo directly: `user.Basic`'s
  `MarshalJSON` **moved to a pointer receiver**, so marshalling a `Basic` *value* silently produces
  default field JSON instead of the custom form — no compile error.

### Notes

- The repository gained the open-source kit it was missing — `SECURITY.md`, `CONTRIBUTING.md`,
  a secret-scan configuration and the README sections pointing at them — plus this file.

---

The entries below were **reconstructed from git history** rather than written at the time, so they
say what each tag contains, not why it was decided.

## v1.1.3 · v1.1.2 · v1.1.1

- Dependency updates only.

## v1.1.0

- No library change: continuous-integration and linter configuration, a dependency-review workflow,
  and dependency updates. The API is identical to v1.0.2.

## v1.0.2 and earlier

- Not reconstructed. See the git history and the tag list.
