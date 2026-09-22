# Continuity: how long a people lasts, and how much of its past reaches its present

**Status:** Draft (2026-09-22). The foundation the leaders and empires proposals both read, and `art.md` after them (an art is a memory that does not wear, which makes art a continuity term). Depends on `specs/plan.md` step 7 (the random stream per people) for anything it can be tuned against.
**Last updated:** 2026-09-22

## Problem

The simulation has no lifespan. `longlived` and `shortlived` are binary traits that multiply the rate a people's ways set (1.4 and 0.6 in `stiffGrowth`) and tilt one filter. `life_extension` is a node a people either knows or does not. Nothing can express four hundred years, nothing can express a plague halving it, and nothing can express the difference between a people that turns over thirty times a tick and one that turns over once.

More to the point, **continuity already exists and does only one job.** `memory(c)` in `lore.go` computes the multiplier on a tale's wear from the profile's memory (machines a twentieth, planetary minds a sixth), the `swarming` and `collective` discounts, the `memoryTable` of nodes — writing, printing, networks, substrate minds — and the communion boon. That is a measure of how much of a people's past reaches its present, and it is applied to forgetting and nothing else.

Meanwhile the step-6 sampling found a galaxy where peoples stand unchanged for tens of millions of years (the median standing people is 14 to 53 Myr old and the largest holder has held its worlds for 28 to 77) with no variable in the model expressing why some should and some should not.

## Scope

**In.** A lifespan on the species and the tech and conditions that move it; a derived continuity on the people; `memory()` rewired to read it; continuity as an input to ossification and to the depth of a dark age; the archive as a structure; the deathlessness cluster — the `uploading` filter that does not exist, and what a people becomes when it passes.

**Out.** Leaders (`leaders.md`). Anything about who holds what (`empires.md`). Population, which `art.md` settles: there is no population *field*, but the organic yield of a people's holdings is its carrying capacity and over a thousand-year tick it is filled, so population is already computed and neither proposal adds one.

## Design

### Lifespan

`Species.Lifespan` in years, drawn at birth between 30 and 1000 for biologicals — a tree-kind that lives nine centuries is as ordinary here as a mayfly kind. The profile decides whether the draw applies at all: a machine people, a hive, a planetary mind and an eldritch thing have no natural turnover and their lifespan is effectively the age. A parasite's is its hosts'.

**Generations per tick** is the number every mechanism reads, not the lifespan itself: at a thousand-year tick a thirty-year kind turns over thirty-three times and a thousand-year kind once, and that ratio is what institutional memory is a function of. It is scale-free and it drops straight into a per-tick decay of the same shape as the wearing's compounding.

Tech moves it: `medicine` and `sanitation` at era 1, `life_extension` at era 3 (which already triggers the Long Silence), `germline`, and the machine road through `uploading` and `substrate_minds`. Plague and squalor cut it hard and temporarily. Most of those nodes currently grant half a point of survival and nothing else.

### Continuity

`Civ.Continuity` on 0 to 1, derived each tick from the generations per tick, the memory nodes already in `memoryTable`, the archives standing, and the profile. Never quite 1, even for a machine: a mind that has run ten million years has drifted or chosen to forget. **A dark age cuts it hard**, which is what a dark age is.

**Both ends fail, differently**, and this is the point of the whole proposal:

- **High continuity** — institutions persist, the past has authority, nothing is let go. Faster setting (`stiffGrowth`), and the brittle shatter when it finally breaks. A people that remembers everything cannot change.
- **Low continuity** — knowledge lost between generations, no institution to inherit, succession is chaos. Deeper dark ages, faster forgetting, and the inability to hold a large realm.

So there is no best lifespan, and different kinds fail in different ways, which is where the variety of causes the spec keeps asking for comes from.

**Continuity is not stiffness.** Stiffness is the present's rigidity — can this people change? Continuity is the past's reach — does this people know what it was? The corners are the interesting cases: high continuity with low stiffness is a people that remembers everything and still adapts, the closest thing to a golden age; low continuity with high stiffness is a people doing things for reasons nobody can name any more, which is currently unreachable and is one of the better stories available. The proposal fails if the two collapse into one number during tuning.

### The archive

A structure at a star that raises continuity, and earns its place four times: it has an owner and a location; it slows forgetting; **it can be taken or destroyed**, which makes it a war target whose loss lands ticks later; and when its people falls it becomes a remain whose testament a finder reads — the existing wall-reading path, where a people that finds its own wall gets its memory back.

### The deathlessness cluster

**`uploading` has no filter.** Twenty nodes trigger one; this is not among them, so biological deathlessness costs a facing at difficulty 4.5 (the Long Silence) and machine deathlessness is free. A filter on it, in the Blindsight register:

- **Declined — Heaven.** The people uploads into a paradise and stops acting. It is **gone**: it stops holding, its worlds free, and the decline index sees the contraction. What it leaves is not wreckage but a running substrate on a world, full of minds that are still there and are not coming out, carrying the telling of a people frozen at the moment it left.
- **Scarred.** A partial withdrawal, or the mirror of `mortality`: a creed against uploading, the flesh made sacred.
- **Overcome.** They upload and keep going: **software now**, a substrate change of the kind `transcendence` and the `becomings` in `causes.json` already model. Continuity jumps, stiffening changes, lifespan goes effectively infinite, and biology's plagues are traded for computation's — `basilisk`, `cognitive_immunity` and `antimemetic_resilience` are already in the tree.

This gives machine-kind a second road that is elective rather than a rebellion, and it makes Heaven a new contraction cause, which `DESIGN_NOTES.md` already asks for ("the Long Silence produces most remnants ... worth diversifying the contraction causes").

## Implementation notes

- `memory()` becomes a reader of `Continuity` rather than a computation of its own; the `memoryTable` moves behind it. The wearing is already on a four-tick cadence after step 7, so this costs nothing per tale.
- Continuity is derived, not stored durably; the fold test's durable/derived table gains a row.
- `stiffGrowth` gains a continuity term, and `longlived`/`shortlived` become derived from the lifespan rather than consulted directly.
- The `silence` filter's trait table (`longlived` +1, `memory` +2, `shortlived` −1) is the existing statement of the sign this proposal generalises; it should keep working unchanged.

## Open questions

- **The shape of continuity.** What exactly it is a function of, and whether the memory nodes belong inside it or beside it.
- **Where high continuity bites.** Ossification only, or also the odds that a break is a shatter rather than a renaissance.
- **Whether the archive is one structure or a class**, and whether a people can read another's archive without the people being dead.
- **What Heaven leaves behind mechanically**: a remain, a legacy, or a new kind, and whether anything can wake it.
- **Whether uploading's filter is faced on learning the node** (like every other) or on a decision to use it.
- **The lifespan band**, and whether 30 to 1000 is the right range once medical tech is on top of it.
