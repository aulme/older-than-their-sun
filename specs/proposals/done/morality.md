# Morality: what a people counts as wrong

*Absorbed into `DESIGN_NOTES.md` § Simulation v2, "Morality" (plan step 19, 2026-09-21). The Design section below stays the specification where the notes are silent; the notes win where they differ.*

**Status:** Implemented (plan step 7; see the Morality bullet in `DESIGN_NOTES.md`)
**Last updated:** 2026-09-18

Assumes [one-tick](one-tick.md): every rate below is per thousand years, and "per tick" means a thousand-year tick, the same thing.

## Problem

Every people judges every act by the same table. A fact's moral sort (deed, crime, woe, bond, folly) is fixed per fact kind, so a hive holds enslavement as much a crime as an individualist does, a swarm is as squeamish about a farmed sentient as a herd people, and a people that exists to take worlds calls a conquest a crime when it is done to somebody else. The tellings, the monster reckoning and the dial shifts all read from that one table, so peoples differ in what they *remember* and in how they *slant* it, but never in what they think is wrong. The sentient Manna in the resources proposal needs an observer-dependent judgment, and so did enslavement, breeding and scouring all along.

## Scope

**In.** A per-civ morality with four kinds, rolled from organisation, kind and drive; an override table from morality and fact kind to sort; the reckoning, dials and scapegoats reading through it; fixation as a tilt on the direction order and on trade; morality in the portrait, the difference score and the tellings; drift by schism, scar and uplift; a `techstats` breakdown.

**Out.** Whether a people lives up to its own standard: that is honour, already rolled. Religion as a system. Player-facing choice of morality.

## Design

### The field

```
Morality { Kind: Amoral | Individual | Herd | Fixation; Object: string }
```

on `Civ`, set at birth from the species and changed only by the drift events below. `Object` is set for a fixation only: `conquest`, `spawning`, `knowing`, `old things`, `holding`.

| Kind | The good is | Portrait |
|---|---|---|
| amoral | nothing; there is no word for wrong | "who have no word for wrong" |
| individual | each one, in itself | "who hold each life a good in itself" |
| herd | the many; the one is nothing | "for whom the many are everything and the one is nothing" |
| fixation | the Object, and nothing else | "for whom the only good is more of themselves" / "the taking of worlds" / "knowing" / "what the old ones left" / "holding what they have" |

### The roll

Weights start at individual 30, herd 30, amoral 15, fixation 25, then:

| Given | Change |
|---|---|
| hive, herd or nonconscious organisation; swarm kind | herd ×3, amoral ×2, individual ×0.1 |
| individualist or solitary | individual ×3, herd ×0.2 |
| caste, collective | herd ×2 |
| eusocial | herd ×2 |
| machine-born | fixation ×2, amoral ×2 |
| parasite | amoral ×2, fixation ×1.5 |
| planetary mind | amoral ×3, fixation ×2, individual 0, herd 0 |
| conqueror | fixation (conquest) ×3 |
| expansionist, or a swarm | fixation (spawning) ×3 |
| curious, contemplative | fixation (knowing) ×2 |
| curious with the Sight, or a branch of a people that found much | fixation (old things) ×2 |
| pragmatic, greedy | fixation (holding) ×2 |
| pacifist | amoral 0, individual ×2 |

The fixation's object is rolled among the five with the same tilts, default even.

### The table

`sortFor(c, f)` replaces `f.sort()` wherever a people judges: `hold`, `regard`, `reckon`, `loreDials`, `scapegoat`, the testament and the telling text. The default is the fact's own sort; the table overrides by morality and fact kind. `-` means the fact is not a crime, a folly or a deed to this observer: it is held as a woe by the sufferer and as nothing by anyone else.

| Fact | individual | herd | amoral | conquest | spawning | knowing | old things | holding |
|---|---|---|---|---|---|---|---|---|
| enslaved | crime 4 | crime 1 | - | deed 2 | crime 2 | - | - | deed 1 |
| bred | crime 4 | - | - | - | deed 2 | deed 1 | - | - |
| scoured, home broken | crime 5 | crime 5 | - | deed 2 | crime 5 | crime 3 | crime 2 | crime 3 |
| burned | crime 3 | crime 3 | - | deed 1 | crime 4 | crime 2 | crime 3 | crime 3 |
| taken, war | crime 2 | crime 2 | - | deed 2 | crime 2 | crime 1 | crime 1 | crime 2 |
| betrayal | crime 3 | crime 4 | - | folly 1 | crime 2 | crime 2 | crime 2 | crime 3 |
| stripped | crime 3 | crime 2 | - | deed 1 | crime 2 | crime 1 | crime 2 | crime 4 |
| the Manna eaten (sentient) | crime 2 | - | - | - | folly 1 | crime 1 | - | - |
| kinfed (a caste for the table) | crime 2 | - | - | - | folly 1 | crime 1 | - | - |
| unleashed, horror made | folly 4 | folly 4 | folly 2 | folly 3 | folly 4 | crime 2 | crime 4 | folly 4 |
| sealed | deed 2 | deed 2 | - | deed 1 | deed 1 | folly 2 | folly 3 | deed 2 |
| mastered, find | deed 2 | deed 2 | deed 1 | deed 1 | deed 1 | deed 4 | deed 5 | deed 3 |
| uplift | deed 3 | deed 1 | - | deed 1 | deed 3 | deed 3 | deed 2 | folly 1 |
| freed | deed 4 | deed 1 | - | folly 2 | deed 2 | deed 1 | deed 1 | folly 1 |
| yield | deed 3 | deed 2 | deed 1 | folly 3 | deed 1 | deed 2 | deed 2 | deed 2 |
| settle, stars | deed 1 | deed 2 | deed 1 | deed 1 | deed 3 | deed 1 | deed 1 | deed 2 |
| a cut-off, an embargo | crime 1 | crime 1 | - | - | crime 2 | - | - | crime 3 |

Everything not listed keeps its default. A people's *own* acts are judged by the same row, so an individual people excuses its own enslavements in the retelling as now, while a conquest people never needs to. Woes are woes to their sufferer under every morality; a morality decides what is done *to others*, not what hurts.

### What follows without new code

- **Monsters.** The reckoning sums crimes. An amoral people therefore holds no monsters, and the fearful-neighbour rule and the no-trade rule never fire for it. A conquest people holds as monsters those who betray or yield.
- **Dials.** The lore shifts read sorts: remembered follies cut risk, crimes raise hate and fear. An amoral people's dials drift only from bonds and finds.
- **Scapegoats.** The sweep hangs crimes, follies and woes: an amoral people has nothing to hang but woes.
- **Testaments.** The eight dearest tales are the makers' judgment, and a finder of a different morality reads a proud deed as a crime. That is already the slant rule, now with a second axis.

### What needs code

- **Direction.** A fixation puts its category first before every other rule but a fleet in flight: conquest arms, spawning road, knowing mind, old things road (surveyors) and the Find's eagerness up a step, holding works.
- **Trade.** A holding people sends nothing and takes everything: its cap out is 0, its want is filled first; partners tire of it (embargo from their side after a stretch). A spawning people gives O freely (cap ×2 for O).
- **Difference.** Morality difference adds to the xenophobe's score: 1.0 between individual and herd, 0.5 to or from amoral, 0.5 between two different fixations, 0 between the same kind.
- **Drift.** A schism may split on morality: the branch rolls again with the parent's weights doubled toward the other kinds (chance 0.3). The Church scar turns a people to fixation, object `knowing` or `holding`. An uplift takes the uplifter's morality with chance 0.5, else rolls. A machine-born successor rolls with fixation ×3, object the thing its makers were doing when it outgrew them.
- **Text.** The portrait sentence; a telling line where the judgment differs from the default ("The X count it no crime"); `techstats` counts moralities by kind, monsters held by morality, and how often the same fact is a crime to one people and a deed to another.

## Implementation notes

One stage, after stage 1 of resources and trade lands, since the direction order has to exist for the fixation tilt. Order: the field and the roll, the table and `sortFor`, the reads, then direction, trade, difference and drift, then text and stats. Verify with the determinism check and the batch; the Tellings section of the report should show monsters falling for amoral and herd peoples and rising for fixations.

## Open questions

- Is a fifth kind wanted, `sacred` (the good is what the faith says; objects rolled from the religion nodes), or is that a fixation with a religious object?
- Should the herd row treat a *foreign* herd's ending (scoured) as a crime, as written, or only its own? Written: a crime, since a herd is the unit of value wherever it is.
- Does morality belong on `Species` (inherited by branches, printed with the traits) or on `Civ` (drifts)? Written: `Civ`, seeded by the species roll.
- The numbers in the table are a first pass.
