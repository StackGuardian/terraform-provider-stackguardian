# Harden `terraform-provider-stackguardian` against the HashiCorp provider baseline

## Context

`terraform-provider-stackguardian` is published on the Terraform Registry (community tier, 57 versions,
~236k downloads, latest `v1.12.2-rc3`). Its *provider code* is already ahead of
`hashicorp/terraform-provider-consul`: it uses terraform-plugin-framework v1.19 + terraform-plugin-testing
v1.16, while consul is still on the legacy terraform-plugin-sdk **v1**. The gap is not in the Go code —
it is in the **engineering scaffolding around it**, and that scaffolding is what makes a provider
contributable by outsiders and safely releasable.

Concretely, today:

- An outside contributor's PR **cannot pass CI**. The only test job runs acceptance tests against the
  *production* StackGuardian API using repo secrets, which fork PRs never receive.
- There is **no lint, vet, or gofmt gate at all**. The `Unreleased` CHANGELOG section documents
  hand-fixing errcheck / staticcheck / vet / gofmt findings — evidence that these are caught manually.
- `make test` is **broken** (`go test -i`, removed in Go 1.20) and CI never invokes it, so nobody noticed.
- Published binaries report version `0.0.1`, and the provider serves the HashiCups-tutorial dev address.
- The provider **logs the API key in plaintext** at debug level.
- `.goreleaser.yml` is a GoReleaser **v1** config, ships 6 of the 13 platform targets HashiCorp ships,
  and marks every release a draft requiring a manual publish click.

The outcome we want: a fork PR runs build + lint + unit tests + docs-drift check with zero secrets and
goes green; maintainers get acceptance tests on trusted refs; releases are reproducible and automatic;
and the repo has the issue/PR/security/dependabot furniture contributors expect.

**Scope chosen by the user:** CI & testing, Release & registry, OSS community hygiene.
Docs/examples restructuring is explicitly **out of scope** (see *Deferred* at the end).

---

## Baseline comparison (what we are measuring against)

| Area | consul (`hashicorp/terraform-provider-consul`) | scaffolding-framework (modern ref) | stackguardian (today) |
|---|---|---|---|
| Plugin SDK | plugin-sdk **v1** (legacy) | plugin-framework | plugin-framework v1.19 ✅ **ahead** |
| Fork-safe CI | ✅ tests run a local `consul` binary, no secrets | ✅ no secrets needed | ❌ prod API + secrets only |
| Lint gate | `gofmt` + `go vet` + `errcheck` scripts in CI | `.golangci.yml` + golangci-lint-action | ❌ none |
| Unit tests | 64 test files / 68 source files | per-resource tests | 22 test files / 103 source files, 84/93 are `TestAcc*` |
| `CheckDestroy` / sweepers | ✅ used | ✅ | ❌ zero of both |
| TF version matrix | 4 Consul versions | 2 Terraform versions | ❌ single TF 1.14.0 |
| Docs drift check | ✅ regenerate + `git status --porcelain` fail | ✅ `git diff --exit-code` | ⚠️ `tfplugindocs validate` only (weaker) |
| GoReleaser | v1 syntax, but CRT-managed 13-target build | **v2** syntax, 13 targets | v1 syntax, 6 targets, `draft: true` |
| Version wiring | CRT `-X main.version` | `var version = "dev"` ✅ | ❌ `const Version = "0.0.1"`, ldflag is a no-op |
| Issue / PR templates | ✅ both | ✅ PR template | ❌ neither (CONTRIBUTING.md promises one) |
| Dependabot | ❌ (CRT/TSCCR instead) | ✅ gomod + /tools + actions, grouped | ❌ none |
| SECURITY.md | ⚠️ `.github/SUPPORT.md` | ❌ | ❌ |
| CHANGELOG | ✅ current, per-release, PR-linked | ✅ | ⚠️ last release entry is `0.1.0` (2024-03-14) vs 57 published versions |

Note: consul is a poor model for *provider code* (SDKv1, `helper/schema`) — do not copy its resource
patterns. It is a good model for CI shape, changelog discipline, and release rigor.

---

## Plan

Five independent PRs, ordered so each lands green. PR 1 is a prerequisite for PR 2.

### PR 1 — Correctness & security fixes (small, land first)

These are real defects found during the comparison; they are cheap and unblock PR 2.

| File | Change |
|---|---|
| `internal/provider/provider.go` | **Remove the `tflog.Debug(ctx, fmt.Sprintf("API Key: %s", api_key))` call.** It writes the secret to `TF_LOG=DEBUG` output. If a log line is wanted, log only whether the key was sourced from config or env. |
| `internal/provider/provider.go` | Fix the three copy-pasted `"Missing Organization Name"` diagnostic summaries so `api_key` and `api_uri` report their own attribute. |
| `internal/provider/provider.go` | Fix the post-unknown-check guard: it tests the local `diags` instead of `resp.Diagnostics`, so unknown-value errors do not short-circuit `Configure`. |
| `main.go` | Replace `const Version = "0.0.1"` with `var version = "dev"` and pass `version` to `sgprovider.New(...)`. GoReleaser already injects `-X main.version={{.Version}}`, which today silently targets a symbol that does not exist. |
| `main.go` | Change `Address` from `"terraform/provider/stackguardian"` to `"registry.terraform.io/StackGuardian/stackguardian"`, and delete the copy-pasted HashiCups tutorial comment. Local dev keeps working via a `dev_overrides` block in `~/.terraformrc` — document that in `CONTRIBUTING.md` (see PR 5). |
| `Makefile` | `test:` — drop the `go test -i` line (removed in Go 1.20; this target currently fails outright). Use `go test -race -cover $(TESTARGS) -timeout=120s -parallel=4 ./...`. |
| `Makefile` | `release:` — `--rm-dist --skip-publish --skip-sign` → `--clean --snapshot --skip=publish,sign` (GoReleaser v2 flags). |
| `Makefile` | `clean-examples:` — points at a nonexistent `examples/` dir; retarget at `docs-guides-assets/` or delete the target. |
| `Makefile` | `docs-generate` / new `tools` target — call `go run -modfile=tools/go.mod github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs` instead of a bare `tfplugindocs` from `$PATH`, so CI needs no install step. |
| `Makefile` | `install:` — derive `OS_ARCH` from `$(shell go env GOOS)_$(shell go env GOARCH)` instead of hardcoding `linux_amd64` (currently broken on the macOS dev machine). |
| `go.mod` | `go 1.26.7` → `go 1.26` plus `toolchain go1.26.7`. A patch-pinned `go` directive forces every consumer and CI runner onto one exact patch. Add `.go-version` containing `1.26.7`. |
| `.github/workflows/.old/release-old.yml` | Delete — dead file (Go 1.16, goreleaser v2.9.1, `on: {}`) still tracked in git. |

---

### PR 2 — Fork-safe, layered CI

Rewrite `.github/workflows/test.yaml` into four jobs. Add at workflow level:

```yaml
permissions:
  contents: read
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

and `timeout-minutes` on every job (none has one today).

**Job `build`** — no secrets, runs on every PR including forks:
`actions/checkout` → `actions/setup-go` with `go-version-file: go.mod`, `cache: true` → `go mod download`
→ `go build -v ./...` → `go vet ./...` →
`golangci/golangci-lint-action` (SHA-pinned). This is the scaffolding-framework `build` job shape.

**Job `unit`** — no secrets: `make test` (fixed in PR 1) with `-race -cover`.

**Job `generate`** — no secrets, replaces the weak `docs-validate` check. Follows the consul and
scaffolding pattern of *regenerate then diff*:

```yaml
- uses: hashicorp/setup-terraform@... # terraform_wrapper: false
- run: make docs-generate
- run: git diff --compact-summary --exit-code || (echo "docs/ is stale — run 'make docs-generate' and commit"; exit 1)
```

`tfplugindocs validate` alone cannot detect that `docs/` is out of date with the schema; this can.

**Job `acceptance`** — the only job that touches secrets. Gate it so fork PRs skip rather than fail:

```yaml
if: >
  github.repository == 'StackGuardian/terraform-provider-stackguardian' &&
  (github.event_name != 'pull_request' ||
   github.event.pull_request.head.repo.full_name == github.repository)
strategy:
  fail-fast: false
  matrix:
    terraform: ['1.9.*', '1.12.*', '1.14.*']
```

Optionally add a maintainer-gated path for fork PRs later (`pull_request_target` + an `ok-to-test`
label); not required for this PR, and it must never run untrusted code with secrets.

Also: `.github/workflows/test-api.yaml` declares no `secrets:` in its `workflow_call` block yet
references `secrets.SG_STG_API_URI` etc. — it works only because callers pass `secrets: inherit`.
Declare them explicitly. And `test-api-{stg,prd}.yaml` use cron `*/10 * 1-10 4 *`, which fires only in
April; either fix the schedule or delete the workflows.

---

### PR 3 — Lint config + real unit tests

**`.golangci.yml`** (new, root). Model on the scaffolding-framework v2 config:

```yaml
version: "2"
linters:
  default: none
  enable: [copyloopvar, depguard, durationcheck, errcheck, forcetypeassert, godot,
           ineffassign, makezero, misspell, nilerr, predeclared, staticcheck,
           unconvert, unparam, unused, usetesting]
  settings:
    depguard:
      rules:
        main:
          list-mode: lax
          deny:
            - pkg: "github.com/hashicorp/terraform-plugin-sdk/v2"
              desc: "Use terraform-plugin-framework / terraform-plugin-testing"
formatters:
  enable: [gofmt]
```

The `depguard` rule is worth keeping permanently: `terraform-plugin-sdk/v2` is currently an *indirect*
dep via plugin-testing, and this stops it becoming a direct one.

Expect a first-run backlog. Fix it in this PR; do not add blanket `//nolint`.

**Unit tests** — the highest-value gap. 103 source files, 22 test files, and 84 of 93 test funcs are
`TestAcc*` requiring production credentials. Add tests that run with **no credentials**:

- `internal/provider/provider_test.go` — construct `New("test", http.Header{})()`, assert `Schema()`,
  `Resources(ctx)` (15) and `DataSources(ctx)` (17) return without diagnostics. This is the single
  cheapest guard against schema/`AttributeTypes()` mismatches, the #1 gotcha in `CLAUDE.md`.
- `internal/expanders/*_test.go` and `internal/flatteners/*_test.go` — pure functions (`bool`, `int`,
  `json`, `list`, `maps`, `string`, `float64`), currently at **zero** coverage, and the documented
  source of runtime panics. Table-driven tests over null / unknown / empty / populated inputs.
- Extend the pattern already proven in
  `internal/resource/workflow_from_template/resource_internal_test.go` (`TestSplitTemplateOrg`,
  `TestConvertEnvironmentVariablesFromAPI_NilConfig`) to the other resources' `ToAPIModel` /
  `convertXxxFromAPI` functions.

**Acceptance-test hardening** (consul does both, this repo does neither):

- Add `CheckDestroy` to `resource.Test` cases — zero exist today, so "destroy actually destroys" is
  entirely unverified. Write one shared helper per resource using `acctest.SGClient()`.
- Add sweepers: `TestMain` + `resource.TestMain` in a new `internal/sweep` package, with
  `AddTestSweepers` registering prefix-matched cleanup. Failed acceptance runs currently leak
  StackGuardian resources forever. Start with the resources acceptance tests create most often
  (`workflow_group`, `connector`, `runner_group`, `workflow_template`).
- Replace `internal/acctest/random_acc_test_name.go` `GenerateRandomResourceName()` — an unseeded
  `math/rand` reimplementation — with `acctest.RandomWithPrefix()` from
  `terraform-plugin-testing/helper/acctest`.
- Scrub the real-looking AWS account ID / external ID hardcoded in test configs
  (`internal/resource/connector/resource_test.go`, `arn:aws:iam::209502960327:role/StackGuardian`);
  move to env vars with a skip when unset.

---

### PR 4 — Release & registry

**`.goreleaser.yml`** — migrate v1 → v2:

- Add `version: 2` header.
- `archives[].format: zip` → `formats: [zip]` (v1 key, deprecated).
- `changelog: { skip: true, use: github }` → `changelog: { disable: true }` (`skip` is the v1 key).
- Expand the build matrix from 6 targets to the HashiCorp standard 13:
  `goos: [darwin, freebsd, linux, windows]` × `goarch: ['386', amd64, arm, arm64]`, with the
  conventional `ignore:` list (`darwin/386`, `darwin/arm`, `freebsd/arm64`, `windows/arm`,
  `windows/arm64`). The stale `origin/feat/windows-386` branch suggests users have already asked.
- `draft: true` → `draft: false`. The `v*` tag is already the gate; a draft means every release needs a
  manual publish click before the registry can ingest it, which is why `v1.12.2-rc3` is the registry's
  current "latest".
- Keep the existing GPG `signs:` block, `mod_timestamp`, `-trimpath`, and the double manifest
  attachment (`checksum.extra_files` + `release.extra_files`) — those are already correct.

**`.github/workflows/community.yml`** is a vendored, drifting copy of HashiCorp's reusable release
workflow, pinning `goreleaser/goreleaser-action@v5.1.0` (tag-pinned, unlike every other action in the
file, and two majors behind). Either:

- **(recommended)** replace the vendored copy with the upstream reusable workflow
  `hashicorp/ghaction-terraform-provider-release/.github/workflows/community.yml@<sha>`, or
- bump `goreleaser-action` to v7.x SHA-pinned and add `version: "~> v2"` so the v2 binary matches the
  migrated config.

Also pass `gpg-private-key-passphrase` from `release.yaml` — the input exists and is never supplied.

---

### PR 5 — OSS community hygiene

| File | Content |
|---|---|
| `.github/ISSUE_TEMPLATE/bug_report.yml` | Structured form: provider version, Terraform version, affected resource, config, debug output, expected vs actual, repro steps. `CONTRIBUTING.md` already tells people to "use Bug report issue template" — it does not exist. |
| `.github/ISSUE_TEMPLATE/feature_request.yml` | Use case, proposed schema, references. |
| `.github/ISSUE_TEMPLATE/config.yml` | `blank_issues_enabled: false`, contact link to the existing `StackGuardian/feedback` discussion. |
| `.github/pull_request_template.md` | Description, linked issue, "CHANGELOG updated" checkbox (CONTRIBUTING already mandates this), acceptance-test evidence, docs regenerated. |
| `SECURITY.md` | Supported versions + private disclosure contact. Absent today, and this provider handles API credentials. |
| `.github/dependabot.yml` | Three ecosystems, following the scaffolding config: `gomod` on `/` with a `terraform-plugin` group matching `github.com/hashicorp/terraform-plugin-*`; `gomod` on `/tools`; `github-actions` on `/` grouped as one weekly PR. The separate `tools/` module makes the second entry mandatory — nothing updates `tfplugindocs` today. |
| `.editorconfig` | Go tabs, YAML/HCL 2-space, LF, trailing-newline. |
| `CHANGELOG.md` | Backfill. The file's last released entry is `[0.1.0] - 2024-03-14` while 57 versions are published; the `Unreleased` section holds work already tagged. Generate historical entries from the GitHub Releases API, then keep the Keep-a-Changelog discipline the PR template enforces. |
| `CONTRIBUTING.md` | Add a "Development" section: `make build` / `make test` / `make lint` / `make docs-generate`, the `dev_overrides` snippet for local testing (needed after the `Address` change in PR 1), and what acceptance tests require. Mirrors the consul README's dev section, which this repo lacks. |
| `docs-guides-assets/README.md` | Fix the stale `cd examples/role_example` instruction — that directory does not exist. |

Optional, low cost: `.github/workflows/lock.yml` (`dessant/lock-threads`, the scaffolding pattern) to
lock issues/PRs after 30 days of inactivity.

---

## Verification

Per PR, and end-to-end before tagging:

```bash
# PR 1 — the previously-broken targets now work
make build
make test                      # was failing on `go test -i`
make release                   # goreleaser v2 flags, snapshot only
go run . -h                    # confirm no panic; version wiring compiles

# PR 3 — unit tests pass with NO credentials in the environment
env -u STACKGUARDIAN_API_KEY -u STACKGUARDIAN_API_URI -u STACKGUARDIAN_ORG_NAME make test
golangci-lint run              # must be clean, no blanket nolint

# PR 2 — docs drift check reproduces locally
make docs-generate && git diff --exit-code

# PR 4 — release config validates and builds all 13 targets
goreleaser check
goreleaser build --snapshot --clean   # confirm freebsd/386/arm artifacts appear in dist/
```

CI-level verification, which is the actual goal:

1. Open a PR **from a fork**. `build`, `unit`, and `generate` must go green; `acceptance` must be
   skipped, not failed. This is the check that today's setup cannot pass.
2. Open a PR from a branch on the origin repo. All four jobs run; acceptance runs across the three
   Terraform versions in the matrix.
3. Push a throwaway `v0.0.0-test` tag on a fork to exercise `release.yaml` end-to-end, confirm 13
   artifacts + `_SHA256SUMS` + `.sig` + `_manifest.json`, and that the release is **not** a draft.
   Delete the tag and release afterwards.
4. After a real release, confirm the registry ingests it without a manual publish step, and that
   `terraform providers` reports the correct version rather than `0.0.1`.

---

## Deferred (out of the chosen scope)

Not addressed here, worth a later pass:

- Rename `docs-templates/` → `templates/` and `docs-examples/` → `examples/` to match the registry and
  `tfplugindocs` defaults, which would also let `make docs-generate` drop `--website-source-dir`.
  Touches all 35 template files' `{{tffile ...}}` paths — a mechanical but wide diff, and the
  `docs-examples/datasources/` vs `docs/data-sources/` naming mismatch should be fixed at the same time.
- `docs/index.md` still pins `version = "1.0.0-rc5"` in its example while `README.md` says `~> 1.12`.
- 15/15 resources have acceptance tests but only 2/17 data sources do.
- `github.com/spf13/viper` is a direct dependency used solely to read three environment variables in
  `internal/config/config.go`; `os.Getenv` with a small wrapper would remove a large dependency tree
  from a provider binary.
