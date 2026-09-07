## What does this change?

<!-- A short description of the change and why it is needed. Link the issue it closes. -->

Closes #

## Type of change

- [ ] Bug fix
- [ ] New resource or data source
- [ ] New attribute or behaviour on something existing
- [ ] Documentation only
- [ ] Build, CI or tooling

## Checklist

- [ ] `make build` succeeds.
- [ ] `CHANGELOG.md` is updated (see [CONTRIBUTING.md](../CONTRIBUTING.md)).

If you touched a schema, `docs-templates/`, or `docs-examples/`:

- [ ] `make docs-generate` has been run and the resulting `docs/` changes are committed.
      `docs/` is generated — never edit it by hand.
- [ ] `make docs-validate-examples` passes.

If this is a breaking change or changes an existing resource's behaviour:

- [ ] The effect on existing state is described below, including whether a resource is
      replaced rather than updated in place.

## Notes for reviewers

<!-- Anything that would be hard to see from the diff: an API quirk you worked around,
     a decision you were unsure about, or something you deliberately left out of scope. -->
