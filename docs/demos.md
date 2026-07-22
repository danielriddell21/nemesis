# Demos

A tour of the game and its hunter, in motion. Everything here is generated
programmatically by the committed capture tool (`tools/demogen`) straight from
the simulation and the headless renderers — the same pipeline the game uses. A
scripted pilot plays the objective loop (badly on purpose: it checks its
tracker when scared and dives into vents to escape) while the visualiser
records. Regenerate it all with:

```
just demos        # or: go run ./tools/demogen
```

Every capture runs from a fixed seed, so they are reproducible.

## Watching the hunter think

![the hunter AI visualiser, live](demos/hunter.gif)

Roughly two and a half minutes of play at ~3× speed, spanning several runs.
How to read it:

- **The hunter** is the large dot, tinted by its state — grey while it
  **lurks** dormant in a far vent mouth, blue on **patrol**, amber when it
  **investigates** a noise, red on a **hunt**, purple while it **searches**
  after losing the trail. The panel's STATE line names the same thing.
- **Its plan** is drawn as it forms: the small red squares are the remaining
  cells of its current A\* path (watch it thread the brown vent ducts — they
  cost less than corridors), and the orange ✕ is the target the path leads to.
- **The prey** is the small green dot with a facing tick: the scripted pilot
  sneaking, walking and running between the yellow objective consoles, which
  turn green as they come online. The red map tile is the sealed airlock; it
  turns green when every console is lit.
- **Sound is visible**: every noise — footsteps, a bulkhead banging open, a
  tracker ping, a vent creak — expands as a fading yellow ring with the
  loudness the simulation gave it. When a ring reaches the hunter, the feed
  logs `HUNTER HEARD NOISE AT x y` and the state flips to investigate.
- **The director's hand** shows as `DIRECTOR STEERS HUNT TO x y` in the feed:
  it always knows where the prey is, but only ever points at the
  neighbourhood.
- **The trigger feed** on the right is the live event stream — the same
  telemetry the game's audio and HUD consume: `VENT CREAK AT 12 4`,
  `HUNTER SPOTTED PREY AT 41 23`, `PREY KILLED AT 41 23`, `RUN START`.
- **The vision cone** fans out from the hunter along its facing, coloured by
  its state and reaching as far as it can currently see (shorter in the dark).
  Anything moving inside it with a clear line is spotted.
- **Decoys** show as a pale ring pulsing at the thrown noisemaker's position,
  with `DECOY CHIRPS` in the feed; the hunter treats it as noise and investigates.
- **Learning** is the two LEARNED rows and the red-tinted hot rooms. Pings,
  creaks, cold trails, decoys it walks up to, and lockers it sees you use all
  push the tiers up; heat marks the rooms where it keeps detecting prey. Watch
  the rows survive `RUN START`: a new deck is generated, but the same hunter
  walks into it — deeper and warier — remembering everything.

## In the corridors

![first-person play](demos/corridors.gif)

The same pilot from inside, at real speed: raycast corridors lit by each
room's own gloom, the gait readout and objective tally along the bottom, and
the tracker scope coming up whenever the hunter is near. The clip starts
rolling the moment the hunter first closes within earshot.

![first-person view with the tracker raised](demos/game.png)

A posed still of the encounter: the hunter's silhouette mid-corridor and the
raised motion tracker — its blip painted once per ping, only for a *moving*
contact, with the chirp audible to more than just you.

## Six stations

![six generated stations](demos/stations.png)

The generator's range across six seeds: BSP rooms joined by corridors, the
brown vent ducts threading the wall mass between them, bulkhead doors
(orange), objective consoles (yellow), lockers (cyan), spawn (green) and the
airlock (red). Every layout is deterministic from its seed and guaranteed
traversable.

## The AI visualiser

![the AI visualiser](demos/hunter.gif)

The second window in play (`nemesis --visualiser`): the true map
with rooms, corridors, ducts, consoles and airlock; actor positions; the
panel's state, target, menace gauge, objective tally and learned tiers; and
the trigger feed filling below.
