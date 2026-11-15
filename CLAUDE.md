# nemesis — Claude Code instructions

nemesis is a procedurally generated first-person stealth game hunted by an Alien-Isolation-style AI: each station level is generated from a seed, and the hunter stalks it through the corridors and vents. Correctness and simplicity over cleverness — prefer a well-tested, focused function over a clever abstraction. If the standard library does it, use it.

## Before every commit

* Run `just ci` autonomously (lint + test + build + eval). All must pass.
* `just eval` is the balance gate: the pilot plays 100 seeds, and the run fails if the solvable or escape rate misses the pass marks in `internal/eval`. After touching `internal/sim`, `internal/pilot` or `internal/world`, report the escape rate before and after. Never widen a mark to make a change pass.
* When adding a function or package, write unit tests alongside the code in the same commit.
* The Ebiten window lives in `internal/gui` and is the only package that imports Ebiten. It is ungated — this is a pure game with no headless mode — but stays behind the `gui.Run`/`gui.Available` seam.
* Never commit until the user explicitly confirms. Propose changes as diffs, run `just ci` autonomously, then stop and wait before `git commit`.

## Code quality

* Run `just lint` before proposing a diff. Fix all lint errors before committing.
* Prefer early returns over nesting.
* Do not add error handling or fallbacks for scenarios that cannot happen.
* Comments in `internal/` and `cmd/`: only *inside* functions, and only for non-obvious logic — never a doc comment on a declaration, a file or package header, or a `doc.go`. (`//go:build`, `//go:generate`, and `//nolint` are directives, not comments, and stay.)
* Public library code (module root or `pkg/…`, where present) keeps full godoc — a doc comment on every exported symbol and the package, enforced by the revive `exported` rule.

## Invariants

* The simulation is deterministic from the seed: nothing in `internal/sim`, `internal/world` or `internal/pilot` may read the clock or an unseeded RNG.
* The torch is one fact with three readers. `sim.Game.TorchStrength` decides whether the beam is lit (battery, flicker and all); the hunter's sight, the renderer's cone and the HUD all read it, so what the player sees lit is exactly what gives them away.
* The renderer lights from `world.Level.Light` plus the torch; it never invents light. A cell the level leaves dark stays dark on screen and short-sighted for the hunter.

## Conventions

* The CLI entrypoint and Ebiten GUI structure is shared across the tool family and documented in [CONVENTIONS.md](CONVENTIONS.md). Keep new code consistent with it.
