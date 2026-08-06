# nemesis

> *n.* the pursuer you cannot shake. This one remembers.

[![CI](https://github.com/danielriddell21/nemesis/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/nemesis/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/nemesis/graph/badge.svg)](https://codecov.io/gh/danielriddell21/nemesis)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_nemesis&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_nemesis)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A procedurally generated first-person stealth game written in Go with [Ebiten](https://ebitengine.org/) v2, hunted by an Alien-Isolation-style AI.

Every run drops you into a freshly generated station — rooms and corridors threaded with crawl-height vent ducts — with one other thing aboard. Bring the generator consoles online, then reach the airlock, without being caught by a hunter that hears your footsteps, sees you in the light, stalks you through the vents, and is quietly steered toward you by a director that always knows where you are but never tells it exactly. Clear a station and you descend to the next, deeper deck — against the same hunter, which **remembers what you did**. Levels are deterministic from a seed.

```sh
go run ./cmd/nemesis --seed 42
```

![nemesis gameplay](docs/demos/corridors.gif)

## Install

### Homebrew (macOS)
```sh
brew install --cask danielriddell21/tap/nemesis
```

### From source
On Linux/Windows, build from source:

```sh
go build ./cmd/nemesis
```

<details>
<summary>Linux: OpenGL/X11 libraries</summary>

The window needs OpenGL/X11. On Debian/Ubuntu:

```sh
sudo apt install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev
```
</details>

## Documentation

Full documentation lives in the [nemesis wiki](https://github.com/danielriddell21/nemesis/wiki) — features, controls, CLI flags, how the hunter thinks, and the architecture.
