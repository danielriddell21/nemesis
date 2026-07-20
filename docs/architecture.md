# Architecture

The project is built around three deliberately decoupled core layers, with the
Ebiten front-end on top and a few supporting packages alongside. Dependencies
only ever point one direction — **render → sim → world** — so the world
generator and the simulation can be exercised headlessly, with no graphics in
sight.

1. **`internal/world` — generation.** Pure Go, zero rendering knowledge, no
   Ebiten import. A level is a 2D grid of tiles (walls, floors, doors, vents,
   consoles, a spawn and an exit) produced deterministically from a seed.
   Generation uses recursive **BSP** grid splitting: the map is cut into
   sub-regions, a room is carved into each leaf, and sibling regions are joined
   with corridors. The solid wall mass between rooms is then grouped into
   connected components and crawl-height **vent ducts** are dug through the
   components that touch several rooms — a second circulation system that
   bypasses the door-controlled corridors. Bulkhead doors land in
   doorway-shaped corridor cells, objective consoles are mounted into room
   walls far from the spawn and each other, the escape airlock takes the
   farthest reachable floor cell, and every room rolls its own gloom (some
   with failing, flickering fixtures). A flood-fill reachability check
   guarantees the exit and every console can be reached; failed candidates are
   regenerated from derived sub-seeds, so a seed still maps to exactly one
   level.

2. **`internal/sim` — simulation.** The whole game, headless: player movement
   in three gaits with axis-sliding collision, the noise model, doors and
   objectives, the motion tracker, thrown noisemaker decoys, locker hiding,
   per-deck difficulty, and the hunter AI (see [ai.md](ai.md)). State advances
   one fixed step at a time via an explicit `Tick`. Everything the simulation
   notices is emitted through a single `Observer` seam as structured
   observations — footsteps, creaks, state changes, director nudges, learning
   milestones — which is how telemetry, audio and the HUD hear about the world
   without the sim knowing they exist.

3. **`internal/render` — rendering.** A pure-CPU column raycaster (DDA over
   the tile grid) that reads simulation state and produces an RGBA frame
   buffer. Wall faces borrow the light of the open cell in front of them, so
   corridors stay murky and flickering rooms stutter; procedural textures
   cover plating, bulkheads, duct ribbing and console cabinets; the hunter and
   the airlock beacon are z-buffered billboards. The tracker scope, status
   line and overlay text are drawn on top, and crawling through a duct
   letterboxes the view.

A few supporting packages sit alongside these, all pure and observing inward:

- **`internal/audio`** synthesises every sound effect and the ambient bed as
  deterministic PCM — no files. The bed comes in five menace bands that
  flatten and quicken as the hunter closes in; the Ebiten playback lives in
  `internal/gui`, which attenuates effects by distance.
- **`internal/telemetry`** adapts observations into serialisable events with
  human-readable trigger lines, fanning them out to subscribers and keeping a
  bounded recent feed. It is the wire format of the visualiser link.
- **`internal/hud`** is the small overlay the gui posts messages to and the
  renderer reads back when drawing a frame.

The Ebiten front-end lives in `internal/gui` — the only package that imports
Ebiten — behind the family `Run(Config)`/`Available()` seam. It hosts two
window roles: the **game** and the **visualiser**. The game runs a small
state machine — title, playing, paused, settings — over a multi-deck campaign
(input, fixed-step sim ticks, frame upload, positional audio), persisting
player settings and run history (deepest deck, fastest clear) to disk between
sessions. The **visualiser** is a top-down view of the hunter's mind: the
real map, its state and current A* path, the director's nudges, noise ripples
and the trigger feed. The two are separate processes: `internal/cli` owns a
small hub that spawns the visualiser as a child of the same binary and
bridges them with line-delimited JSON over stdin/stdout (the rubix/hegemony
link pattern). The level itself never crosses the pipe — the game sends its
generation parameters and the visualiser regrows an identical map from the
seed.

`cmd/nemesis` is the thin entrypoint; `internal/cli` owns the cobra command
tree.
