# Contributing to terraform-provider-stackguardian

Thank you for taking the time to contribute! 🎉
Contributions are welcome, and they are greatly appreciated!
Every little bit helps, and credit will always be given.

The following is a set of guidelines for contributing to terraform-provider-stackguardian on GitHub. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a Pull Request.

## Contribution types

### Report Bugs

We use GitHub issues to track bugs at [https://github.com/stackguardian/terraform-provider-stackguardian/issues](https://github.com/stackguardian/terraform-provider-stackguardian/issues). Please use Bug report issue template.

### Fix Bugs and implement Features

All contributions to solve GitHub issues tagged with "bug", "enhancement" and "help wanted" are most welcome and greatly appreciated.

### Documentation

The StackGuardian Terraform provider could always use more documentation, whether as part of the
docs and example directories, or even on the web in blog posts or news articles.

### Submit Feedback

Please use GitHub Discussions to submit feedback and engage with community [https://github.com/StackGuardian/feedback/discussions/8](https://github.com/StackGuardian/feedback/discussions/8).

## Git/Github Guidelines

### Guidelines for contributing changes

If you want to contribute code changes with commits, please follow these simple guidelines:
- Open a Pull Request as soon as you start working on a feature, bug, test or hotfix and label it with `work-in-progress`, while it is not ready to be merged.
- Write a useful description in the commit messages of a Pull Request.
- Give the Pull Request a clear, descriptive title. Release notes are generated
  from PR titles, so the title is what users read in the release it ships in.

### Guidelines for Maintainers collaboration

If you are a contributor with repository write access, please follow these sensible guidelines:
- Do NOT use `git push --force` on the `main` branch.
- Do NOT commit to other contributor's branches without their consent.
- Use Pull Requests if you are unsure and to suggest changes to other contributors.

## Releasing

Releases are cut from an annotated tag. Do **not** create the GitHub release by
hand — the workflow creates it, and creating it yourself publishes an empty
release before any artifact exists.

```bash
git checkout main && git pull
git tag v1.13.0
git push origin v1.13.0
```

Pre-release tags use a dot before the number - `v1.13.0-beta.1`, `v1.13.0-rc.2` -
not a hyphen. Semantic versioning compares `beta.1` as a *numeric* identifier and
orders it correctly, whereas the older `-beta-1` form is a single alphanumeric
identifier compared as text, which would sort `-beta-10` before `-beta-2`.
GoReleaser's `prerelease: auto` marks these as pre-releases automatically, so they
never become the Registry's "latest" version.

Pushing the tag triggers `.github/workflows/release.yaml`, which:

1. **Preflight** — verifies the run is on a tag, and refuses to continue if a
   *published* release already exists for it. Registry versions are immutable,
   so a shipped version must never be rebuilt; cut a new patch version instead.
2. **Test** — runs the acceptance tests against the API.
3. **Build** — GoReleaser builds 13 platform archives, checksums them, signs
   `SHA256SUMS` with the release GPG key, and attaches everything to a **draft**
   release.
4. **Verify and publish** — downloads every asset back off the draft, checks it
   against `SHA256SUMS`, asserts the expected archive count, and only then flips
   the release to published.

The Terraform Registry ingests on the publish event, so it never sees a release
until every artifact is attached and verified.

### If a release fails

Re-run the failed jobs from the Actions UI. Re-runs are idempotent while the
release is still a draft — GoReleaser replaces existing assets rather than
failing on them. Nothing needs to be deleted by hand.

Once a release is **published**, it is final. Do not delete the tag or re-run the
workflow against it: anyone who has already resolved that version has its
checksums recorded in their `.terraform.lock.hcl`, and swapping the artifacts
breaks their `terraform init`. Ship a new patch version instead.

### Adding a build platform

The platform matrix lives in `builds.goos` / `builds.goarch` in `.goreleaser.yml`.
If you change it, update `EXPECTED_ZIPS` in the `provider-release_publish` job of
`.github/workflows/release.yaml` to match, or the verification step will fail the
release.
