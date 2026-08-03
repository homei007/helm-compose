# Contributing

## Commit messages

Use Conventional Commit messages so automated release notes and versioning
can be generated reliably:

- `feat:` for a new feature
- `fix:` for a bug fix
- `docs:` for documentation-only changes
- `chore:`, `ci:`, `refactor:`, or `test:` for maintenance changes

Use `!` after the type, or include a `BREAKING CHANGE:` footer, for a breaking
change. Features create minor releases, fixes create patch releases, and
breaking changes create major releases.

## Releases

Changes merged into `main` are collected into an automated release pull
request. Merging that release pull request creates the GitHub release and
builds packages for the supported platforms.
