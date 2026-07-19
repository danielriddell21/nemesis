# Contributing to nemesis

## Requirements

* [Go](https://go.dev) (stable — version from `go.mod`)
* [just](https://github.com/casey/just)
* [golangci-lint](https://golangci-lint.run/welcome/install/) (for `just lint`)
* [gremlins](https://github.com/go-gremlins/gremlins) (for mutation testing)

## Development workflow

```
just build           # build the binary
just run             # build and run
just run-visualiser  # build and run with the AI visualiser window
just test            # run unit tests
just lint            # golangci-lint
just fmt             # golangci-lint fmt (gofumpt + goimports)
just ci              # lint + test + build
just bench           # run benchmarks
just fuzz            # run fuzz tests
just vulncheck       # govulncheck
just tidy            # go mod tidy
just clean           # remove build artifacts
```

Run `just --list` to see every recipe. Run `just ci` (lint + test + build) before each commit. CI runs the same gate on every push to `trunk` and every pull request targeting `trunk`.

## Conventions

The CLI entrypoint and Ebiten GUI structure is shared across the tool family
(unum is the CLI reference; rubix/vivarium the GUI references). See
[CONVENTIONS.md](CONVENTIONS.md).

## Project layout

```
cmd/nemesis/         entry point
internal/cli/        cobra root + leader/child window coordination
internal/gui/        Ebiten windows: game + AI visualiser (the only package importing Ebiten)
internal/            implementation packages (render, sim, world, audio, …)
docs/                documentation
```

## Commit style

```
type(scope): short imperative description
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`. No period at the end of the subject line; keep it under 72 characters.

## Releases

Releases are triggered by pushing a semver tag — maintainers only. A GitHub Actions workflow runs GoReleaser to build the binaries and update the Homebrew tap; it requires the tap app credentials configured as repository secrets.
