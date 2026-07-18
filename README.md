# air-release

A tiny local release helper: computes the next [semantic version](https://semver.org/)
from [conventional commits](https://www.conventionalcommits.org/) since the latest
git tag, and generates a changelog section.

No configuration, no CI required, no dependencies beyond `git`.

## Install

```shell
go install github.com/air-verse/air-release@latest
```

## Usage

Run from the root of the repository you want to release:

```shell
air-release                # preview the next version and changelog
air-release -write         # also prepend the new section to CHANGELOG.md
air-release -tag           # also create the annotated git tag
air-release -tag -release  # also push the tag and create a GitHub release
```

Flags combine explicitly and nothing is implied — e.g. `air-release -tag -write`
to both tag and update CHANGELOG.md, or `air-release -write -tag -release` for
everything. Only `-write` touches CHANGELOG.md, and `-release` requires `-tag`.

With `-tag`, push the tag manually to trigger your release pipeline
(e.g. goreleaser):

```shell
git push origin vX.Y.Z
```

With `-release`, the tag is pushed for you and a GitHub release is created
via the [`gh` CLI](https://cli.github.com/), using the generated changelog
section as the release notes. Requires `gh` to be installed and authenticated
(`gh auth login`).

## How the version is decided

Commits since the latest `vX.Y.Z` tag are parsed as conventional commits:

| Commit | Bump |
| --- | --- |
| `feat!: ...` or `BREAKING CHANGE` in body | major |
| `feat: ...` | minor |
| anything else (`fix:`, `docs:`, non-conventional, ...) | patch |

The changelog groups commits into Breaking Changes / Features / Bug Fixes /
Performance / Others. Non-conventional commits are kept under Others rather
than dropped, so review the output before tagging.

## License

[GPL-3.0](LICENSE)
