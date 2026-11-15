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
just eval            # balance gate: pilot over 100 seeds, must hit its pass marks
just lint            # golangci-lint
just fmt             # golangci-lint fmt (gofumpt + goimports)
just ci              # lint + test + build + eval
just bench           # run benchmarks
just fuzz            # run fuzz tests
just vulncheck       # govulncheck
just tidy            # go mod tidy
just clean           # remove build artifacts
```

Run `just --list` to see every recipe. Run `just ci` (lint + test + build + eval) before each commit. CI runs the same gate on every push to `trunk` and every pull request targeting `trunk`.

### The balance gate

`just eval` plays the scripted pilot (`internal/pilot`) over seeds 1–100, first with the hunter parked and then against it, and exits non-zero if a pass mark is missed:

| Check | Pass mark | What a failure means |
| --- | --- | --- |
| Decks that fail to generate | 0 | the generator is broken |
| Decks the pilot clears with the hunter off | ≥ 95% | a deck is unsolvable, or the pilot is stuck |
| Decks the pilot escapes against the hunter | 8–60% | under: the hunter is unfair; over: it is toothless |

The marks are constants in `internal/eval`, read by both `just eval` and the CI `eval` job, so they cannot drift apart. A change to the hunter, the pilot, or the generator will move the escape rate: re-run `just eval` and say in the PR what it moved from and to. Do not widen a mark to make a change pass.

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
tools/               developer tooling (demogen, eval)
docs/                documentation
```

## Commit style

```
type(scope): short imperative description
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`. No period at the end of the subject line; keep it under 72 characters.

## Releases

Releases are triggered by pushing a semver tag — maintainers only. A GitHub Actions workflow runs GoReleaser to build the binaries and update the Homebrew tap; it requires the tap app credentials configured as repository secrets.
