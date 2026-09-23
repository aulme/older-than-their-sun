# Continuity: how long a people lasts, and how much of its past reaches its present

**Status:** Sharpened and built at `specs/plan.md` step 8 (2026-09-23); the open questions are answered at the end, and the ledger row has the numbers. The foundation the leaders and empires proposals both read, and `art.md` after them (an art is a memory that does not wear, which makes art a continuity term).
**Last updated:** 2026-09-23

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

## As built (step 8)

The design above stands; what it left open was settled like this, and the code is `internal/history/continuity.go` with its numbers in `mind.Tuning.Continuity`.

- **Lifespan** is `Species.Lifespan` in years, **hashed** from the seed and the blood's id when the blood is registered, never drawn, so giving the bloods a span moves no stream. It is log-uniform in the band the traits name: the short-lived 30 to 60, the very long-lived 300 to 1000, everyone else 60 to 300. **The traits stay** as the names of the band's ends (the epithets, the aptitudes, the evolver's opposites and wisdom read them as before) and nothing reads them for a rate any more: `stiffTraits` lost the long-lived, the short-lived and the unbroken memory, since continuity reads the span itself. An evolver's drift into the other band re-hashes the span inside the new one. A machine, an eldritch thing, a living world and a blood in software have **no turnover** (`Species.Mortal`), span 0. A parasite's span is drawn as flesh's, not read off its hosts; a rider moving between hosts of different spans was more machinery than it would show.
- **The span a body lives now** is the blood's times medicine 1.3, sanitation 1.2, life extension 5 and germline 1.5, the last two refused under a mortality creed, and halved while a plague of the body is in the people. Squalor is not modelled: the flows have no state for it.
- **The shape of continuity.** `exp(-x)`, x the loss a thousand years: the generations (1000 / span) times what one generation loses (0.0668, times the profile's `Memory`, the unbroken memory 0.3, swarming 0.8, collective 0.7, each memory node of the old `memoryTable`, 0.8 per archive up to three, communion 0.5), plus a **drift** of 0.025 that nothing escapes, plus a dark age's cut (1, falling to nothing over 200 kyr). **The memory nodes are inside it**, as a multiplier on a generation's loss, which is what they always were. A deathless mind loses only the drift; so the profiles' old `Memory` of a machine (0.05) and a living world (0.15), and their `Stiffen` (1.5 and 1.3), are gone — continuity says it now.
- **`memory()` reads it**: the wear multiplier is the loss over the loss at 0.5, the continuity the wearing was tuned at. A preliterate hundred-year kind wears at the old rate; a machine at a twentieth, as before.
- **Where high continuity bites**: ossification only, as `(MidLoss / loss)^0.3` on the rate the ways set at, clamped 0.5 to 1.6. Not the odds a break is a shatter: that is already stiffness's (a stiffer people's dark age is deeper), and a second term would make the two ends one number.
- **Low continuity** bites twice: a dark age's depth plus 0.08 per doubling of the loss over `MidLoss`, and the Distance's difficulty plus 0.5 per doubling. Every term reads **the log of the loss**, not continuity: continuity bunches under one for every people with letters and medicine, and the loss is where they differ, a machine's a third of a mayfly's. `MidLoss` is 0.04, where the setting's term averages 1.02 over a batch's people-ticks, so the terms move peoples apart and not the batch.
- **The archive is one structure** (`archive` in `works.json`, unlocked by writing, three at most, fed a fifth of organic under the mind), which the builder values by a `Keeps` term when nothing is wanting. It is a war target in the way every work is: taken, it is a ruin at the star, and whoever holds the star reads it. **Anyone can read another's archive** once it is a ruin, by reaching its star, the way every wall is read; nobody reads it while its people holds it. Its remain holds three walls' worth of the telling (24 tales), where a work's holds one.
- **Uploading is faced on learning the node**, like every other filter (the `upload` row, difficulty 6 on Social, plus half a point per point of stiffness up to three: a people whose ways have set has nothing to stay for). A machine, an unconscious people and a replicator never face it. **Overcome** is a blood of its own with `Software` set and no span, sick of the body's plagues at a quarter and of the mind's as any mind is (at twice, the machines' own rate, a software blood's heirs and cults bred a cascade: see the ledger). **Scarred** is the creed of the flesh. **Declined is Heaven**: a contraction (cause `heaven`) to the world the substrate runs on, where the remnant raises an `upload substrate`, a work the pick never offers; a people in heaven has continuity one, so its telling does not wear. **What Heaven leaves** when the remnant goes at last is the substrate's remain, carrying five walls' worth of the telling frozen at the moment it went in. **Nothing wakes it** here; that is for `miracles.md` if it wants it.
- **The lifespan band** is 30 to 1000 before medicine, and 30 to about eleven thousand after it; generations run from 33 a thousand years to under a tenth.

## Open questions (answered at step 8, above)

- **The shape of continuity.** What exactly it is a function of, and whether the memory nodes belong inside it or beside it.
- **Where high continuity bites.** Ossification only, or also the odds that a break is a shatter rather than a renaissance.
- **Whether the archive is one structure or a class**, and whether a people can read another's archive without the people being dead.
- **What Heaven leaves behind mechanically**: a remain, a legacy, or a new kind, and whether anything can wake it.
- **Whether uploading's filter is faced on learning the node** (like every other) or on a decision to use it.
- **The lifespan band**, and whether 30 to 1000 is the right range once medical tech is on top of it.
