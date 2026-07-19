# How the hunter thinks

The hunter is built the way Alien: Isolation's xenomorph famously is: **two
brains**. The creature itself only knows what it senses; a director above it
always knows where you are, but is only allowed to gesture.

## Senses

Every tick the hunter checks its senses before anything else:

- **Vision** is a cone (~110°) with a range that scales with the light on
  *your* tile — darkness genuinely hides you, and holding still while
  sneaking shrinks your profile further. It always spots you at arm's length.
  Sight lines are grid ray-marches: walls, closed bulkheads and console
  cabinets block them.
- **Hearing** is radius-based. Every action carries a loudness in tiles:
  sneaking ~1.5, walking 4, running 9, crawling through a duct 6, a bulkhead
  6, the tracker ping 5, and a generator roaring online 12. Between footsteps
  you still leak a fraction of your gait's noise, so creeping right past it is
  never entirely free. Its effective hearing sharpens as the director's
  aggression rises.

## State machine

```
lurk → patrol ⇄ investigate → search → patrol
              ↘     hunt    ↙
```

- **Lurk** — dormant in a far vent mouth for the opening moments of a run.
  Loud noise (or the director's clock) wakes it.
- **Patrol** — walks room to room, preferring rooms the director has hinted
  at, pausing to scan.
- **Investigate** — something was heard: head to the noise, then sweep the
  area.
- **Hunt** — you were seen: while the trail is warm it re-paths straight at
  you twice a second. Lose it and it falls back to your last known position.
- **Search** — several rounds of checking cells around where it lost you,
  then back to patrol.

Movement is A* over the walkable grid. Vent cells cost 0.6 of a floor cell —
the ducts are its highways, and it moves half again faster inside them —
while closed bulkheads cost extra (it has to shoulder them open, a pause and
a bang you can hear). Crossing a vent grate creaks, which is exactly the
"vent creak at 12 4" trigger the visualiser logs and you may hear nearby.

## The director

The director runs above the creature and paces the run:

- **Menace** is the tension dial: it climbs while the hunter is near (or
  hunting) and bleeds away when it prowls elsewhere. The ambient audio bed
  follows it through five bands.
- **Nudges**: on a clock, the director points a patrolling or searching
  hunter at the *room you are in* — never your exact cell. The creature still
  has to find you with its own senses.
- **The leash**: a hunter that loiters near an unseen player too long is sent
  to a far room, so runs never stall with it camped on your hiding spot.
- **Escalation**: every generator you bring online raises its aggression —
  sharper hearing, more frequent nudges — and a generator roaring to life
  always draws it toward the noise. Unlocking your escape route arms your
  enemy.

## The tracker

The motion tracker is the one tool, and it obeys the fiction: it senses
motion, not bodies. Each ping paints at most one blip — bearing and distance
to the hunter, but only if it was moving at that instant — and the blip fades
until the next ping. A still hunter is invisible. The ping itself is a real
sound in the simulation, so checking the scope with the hunter close is a
gamble, and raising it caps you to a creep.
