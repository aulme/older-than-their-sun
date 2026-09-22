# Decline: what the waning and the end of the age are made of

**Status:** Sharpened 2026-09-22 (`specs/plan.md` step 6), on eight seeds sampled through the whole age. The Design section is the specification; the Stages section is the unit of work.

**Stages 1 and 2 are plan step 7**, after the random stream per people, which is what makes the tuning here measurable. **Stage 3, the force, moved to `empires.md`** (plan step 10): the attrition this proposal specifies and the separatism that proposal specifies are the same check with two outcomes, and building the quiet one now and the loud one later would be building it twice, at a regeneration of the batch each time. The fertility end is therefore kept as a floor from step 7 until that stage lands, exactly as the staging below already says; the stage-3 gate below is inherited by `empires.md` unchanged.
**Last updated:** 2026-09-22

## Problem

The premise says the present is the aftermath of a declining galaxy. The sim's decline is a clock. `runAge` declares the **waning** when the cycle's fertility falls under `FineFertility` (0.5) and the **end** when it falls under `Cycle.Ends` (drawn between 0.05 and 0.2), plus a linger of up to two million years; fertility is `exp(-(now - last surge) / Fade)` with a fade of 16 to 30 Myr. `FineActive` and `EndActive` exist and are 0, so no count of civilisations enters it either. Fertility governs who is *born*: it says nothing about who is standing or what they hold, and since the ossification step nobody dies of age.

### What the map does

Eight seeds at 400 stars, sampled every hundred thousand years through the whole age (`internal/history/declineshape_test.go`, `DECLINE=<dir>`; the sampler draws nothing, and separate unsampled runs of seeds 3 and 5 at the same configuration end at the same year, 60.93 and 33.55 Myr, so these are the ages an ordinary run walks). "Held" is habitable systems held by living peoples, over all habitable systems. "Still" is how long the held share stayed within a fifth of its value at the present.

| seed | habitable | height | when | present | keep | age | waning declared | still | churn /Myr | median age standing | held by the old | largest holder | it has held them |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 194 of 400 | 0.46 | 7.3 Myr (23%) | 0.25 | 0.54 | 32.1 Myr | 12.1 | 23.6 Myr (74%) | 8.9 | 14.3 Myr | 92% | 12 worlds | 28.0 Myr |
| 2 | 182 | 0.48 | 4.6 (6%) | 0.13 | 0.28 | 81.0 | 20.3 | 28.0 (35%) | 1.9 | 53.1 | 100% | 13 | 76.8 |
| 3 | 188 | 0.48 | 53.1 (87%) | 0.44 | 0.91 | 60.9 | 16.0 | 12.3 (20%) | 23.5 | 52.6 | 94% | 59 | 59.8 |
| 4 | 190 | 0.42 | 20.0 (42%) | 0.35 | 0.83 | 47.2 | 14.4 | 26.9 (57%) | 5.4 | 37.1 | 91% | 61 | 42.7 |
| 5 | 204 | 0.65 | 21.4 (64%) | 0.63 | 0.96 | 33.5 | 14.2 | 19.1 (57%) | 5.4 | 26.3 | 98% | 109 | 29.9 |
| 6 | 186 | 0.28 | 9.8 (21%) | 0.20 | 0.70 | 46.1 | 11.9 | 15.3 (33%) | 4.4 | 28.5 | 89% | 15 | 44.6 |
| 7 | 186 | 0.49 | 28.6 (72%) | 0.35 | 0.71 | 39.5 | 15.2 | 10.6 (27%) | 31.2 | 21.3 | 89% | 57 | 39.2 |
| 8 | 194 | 0.38 | 25.2 (42%) | 0.20 | 0.52 | 59.6 | 20.7 | 13.2 (22%) | 7.0 | 48.6 | 89% | 16 | 51.7 |

Three things the sampling says, none of which the old five-seed reading showed:

**1. The age reaches a stationary occupancy and sits in it.** The present holds 28% to 96% of what the height held, a median of 71%, and the last 20% to 74% of every age — a median of 35%, and three quarters of seed 1 — has a held share within a fifth of the value it will have at the present. In three seeds of eight (3, 5, 7) **the height falls in the last third of the age**: the galaxy is at its fullest at the moment the legends call it the waning. Seed 2 spends its last forty million years, half its age, at an occupancy that does not move.

**2. What stands is old, and the old do not move.** At the present, 89% to 100% of the held habitable worlds are held by peoples over five million years old; the median standing people is 14 to 53 Myr old; the largest holder has held its worlds for 28 to 77 Myr, and holds 21% to 54% of everything held. What churn there is — 2 to 31 stars a million years changing hands, being settled or being lost in the last third of the age — happens among the small and the young and leaves the total flat. So the map is not frozen; it is in a balance of settlement and loss around a level that nothing pushes down, with the bulk of it in hands that nothing reaches.

**3. Where a force is most needed there is no room left in the one that exists.** The hazard (`updateHazard`) is capped at 2.5, and in four seeds of eight (3, 4, 5, 8) it sits at exactly 2.5 through the whole last third of the age. Among those four are the three largest standing empires in the batch, of 109, 61 and 59 worlds. Adding a decline term inside that `min` would be inert in half the runs, and inert in the runs that need it most.

The legends' line reads what none of this shows. `KWaning` says "The age is waning. Few still rise, and those that stand are old." At the moment it is declared, two thirds to four fifths of the active peoples are still rising (19 to 46 of 29 to 59), and their median stiffness runs 0.10 to 0.51.

Two things are wrong with it, separately, and the measurement changes which is the hard one. The **definition** does not read anything a player can see. And the **force** is missing, not as a gap to fill after the definition but as the whole of the work: an index read off this map cannot end an age, because this map does not empty. See "Why the index cannot end an age on its own".

## Scope

**In.** A decline index built from variables the sim already has, read once a tick; the waning and the end declared from it, in place of the fertility thresholds; the pressures that make the index climb, from levers that exist and one that does not; the legends and techstats reading the index; the table above as the number to beat.

**Out.** New kinds of catastrophe. Changing the cycle (surges, fade, the state beneath). Anything about how peoples are born. The old-quarrel war loop (`specs/notes/war-halving.md`), which is its own case.

**Carried here by the plan** (step 6 folds them in, since they are all "what a people is to remember and when"): how often a tale wears, whether an heir inherits its parent's whole telling, and whether a first meeting trades tales about third parties. See "What a people is to remember"; they are answered there and are not part of the decline design.

## Design

### The index

Decline is not one number in the fiction either, so the index is several, each on 0 to 1, each something the aftermath could show. Three terms, each against the age's own running peak:

- **Held**: habitable systems held by living peoples, over the highest that share has been. Habitable rather than all systems, because it is the share of worlds someone could live on that someone does, and because the denominator then does not move with what an envelope happens to allow.
- **Rising**: peoples active, not ossified and under stiffness 1 (`risingCount`, which exists), over the highest that count has been. This is what the fertility clock was standing in for.
- **Births**: peoples born in the last two million years, per million years, over the highest that rate has been. Fertility, but read off what happened instead of the clock.

`index = 1 − (held + rising + births)/3`, each term clamped to [0,1].

**Two candidates measured and dropped.** *Reach* — the share of systems within some living people's reach — **rises** through the age and keeps rising through the waning: 0.02 to 0.81 over seed 2, 0.06 to 0.70 over seed 5, and flat or climbing at the present in all eight. Reach is a technology, not a possession; a galaxy whose last empires can cross it easily is not thereby young. It is a good diagnostic — held over reach is the real occupancy, and it separates the batch where the held share does not: 0.84 in seed 5, whose last empire holds everything it can cross, against 0.10 in seed 2, which can cross the galaxy and holds a tenth of it — and it belongs in techstats, not in the index. *The youth of what stands* — median stiffness over active peoples — is 0 at the present in seven seeds of eight and bounces between 0 and 0.8 tick to tick in the eighth, because stiffness grows with worlds held and most peoples hold one. The age of what stands is the real signal (median 14 to 53 Myr), but it is already in the rising term and reads better in the dossier than in the index.

**Relative, with the absolute beside it.** Every term is against the age's own running peak, never an absolute: an age that never had a height has not declined from one, and a poor galaxy would otherwise read as declined at its dawn. The peaks are running maxima, taken as the age goes and never reset — a height the age does not return to is exactly the point. The absolute numbers (held share of habitable systems, peoples standing, births per Myr) go in the dossier beside the index, so a player comparing two galaxies has what the relative reading hides.

**The bars.** The waning is declared when the index is at or over `WaningBar` continuously for `HoldMyr`, so one bad tick does not declare an age; the end when it is over `EndBar` for `HoldMyr`, plus the linger, as now. First setting: `WaningBar` 0.35, `EndBar` 0.55, `HoldMyr` 1. They come from the runs as they stand — the index at today's waning declaration is 0.21 to 0.58 (median 0.36) and at today's present 0.47 to 0.80 (median 0.54) — so these bars reproduce roughly today's age before the force is added, and are then tuned with it. `MaxFades` stays as the backstop and `Capped` stays the flag. `FineFertility`, `FineActive` and `EndActive` go at stage 2; `EndFertility`, `EndFertilityLow` and `Cycle.Ends` stay until stage 3 as the end's floor (see the stages), and go with it. `Cycle.Floor` stays whatever happens, since `ageEnd` and the deep pass read it.

### Why the index cannot end an age on its own

The index above, computed over the eight sampled ages exactly as the sim would compute it (running peaks, two-million-year birth window, held a million years):

| rule | crosses 0.4 | crosses 0.6 | crosses 0.75 |
|---|---|---|---|
| the three terms, mean | all eight, at 5.3 to 30.3 Myr | seeds 2 and 8 | **seed 2 alone** |
| held alone | seeds 1, 2, 8 | seed 2 | seed 2 |
| rising alone | six of eight | seed 2 | seed 2 |
| births alone | all eight, at 0.5 to 4.2 Myr | all eight, at 2.0 to 7.5 | all eight, at 3.0 to 10.0 |

The proposal's first setting — the three terms, the end at 0.75 — would end **one age in eight**. The other seven would run to `MaxFades` and be flagged. Drop the end bar to 0.55 and five of eight end — at 70% to 91% of today's length in four of them, and seed 2 at 11 Myr where the clock ran it to 81, since seed 2 is the one age that genuinely empties and it empties early. The three that never end (3, 4, 5) are exactly the three whose present holds 83% to 96% of the height. Lean on the births term instead and the index is the fertility clock again, worse: it passes 0.75 between three and ten million years into every age, which is the youth.

So the force is not the second half of this proposal. It is the proposal. A definition read off the map is a few hundred lines and a table; it is worth having on its own, and it declares nothing until the map moves.

### The force

Four levers, in the order they are added, each with a first setting to be tuned against the table.

**1. The old let go of the outer worlds.** The attrition the old proposal called "the one lever that does not exist", and the measurement says it is the one that is needed: 89% to 100% of what is held is held by the old, and nothing in the sim takes a world from a standing people that is not at war. A people that is **active, holds more than one world, whose ways are set (`Stiff` at or over `LetGoStiff`, first setting 1) and which has been still longer than `LetGoStillKyr` (first setting twice the ossify tuning's `StillKyr`)** lets go of its farthest holding at a rate per thousand years of `LetGoRate × index × (Stiff − 1) × (worlds − 1)`. The index is a factor, so this does not happen in the youth at all. The seat and the cradle are never let go, so a people always keeps a world.

What it writes: the world becomes unowned, a remain is left on it (the ruins of a province, with its testament), and `FLetGo` is a woe of the people that let it go — the fact the legends tell, and the tale its neighbours inherit.

What it must **not** do: it must not renew. `loseSystem` ends with `w.renew(c, 0.05) // a loss is something new` (`civ.go`), which takes stiffness off and resets `Still` — and stiffness and stillness are the two conditions this lever reads, so a realm that let a world go would stop qualifying and the lever would switch itself off after one use. Letting go therefore takes a path that skips the renewal, and the exception carries a comment saying why: a province let go is not something new happening to a people, it is the sum of what that people has stopped doing.

It is not contraction in the sense the spec forbids. The ossification rule is that a people is never forced into the contracted end state and never dies of age; a realm shedding its provinces leaves the people standing, at its seat, alive. The spec's ossification paragraph gains a sentence saying where the line is.

**2. A world let go is spent for a while.** `Spent []Year` on the world: until `Spent[star]`, `canLive` says no and the colony targeting skips it. First setting `SpentKyr` 5000, five million years. This clause exists because of the churn: settlement and loss already run at 2 to 31 stars a million years around a flat total, so a world freed by the attrition would be refilled by the balance and the total would not fall. Set `SpentKyr` to zero to measure exactly that, which is the first thing stage 3 should do.

**3. The hazard reads the index, and its ceiling rises with it.** `w.Hazard = min(HazCap + HazCeil×index, law + … + HazTerm×index)`, first setting `HazTerm` 0.5 and `HazCeil` 0.5 over the present cap of 2.5. The ceiling has to move or the term is inert: the hazard already sits at 2.5 through the last third of four seeds in eight. Every filter is 1.5 harder per point of hazard over one, so at an index of 0.6 this is about half a point of difficulty on everything, late.

**4. A late dark age is deeper.** `depth` (`civ.go`, `DepthBase + DepthStiff×… + DepthPrior×… + noise`) gains `DepthDecline × index`, first setting 0.2, so a late fall takes more of the tree and the break is likelier to shatter a people into heirs than to be climbed out of.

**Not in the force, and why.** The old proposal's "what eats and what speaks" (a late seeding of replicators and transmitters) and the wall thinning faster in the waning both work by raising the hazard, and the hazard is at its ceiling where it is most wanted; they buy little until lever 3 moves it, and each adds events the legends have to carry. They are deferred with a trigger: if the table after stage 3 still shows a height in the last third, seed the predators late.

**Tunables.** One `DeclineTuning` group in `mind.Tuning`, so the whole calibration is `-tune Decline.LetGoRate=…` and no recompile: `WaningBar`, `EndBar`, `HoldMyr`, `BirthWindow`, `LetGoStiff`, `LetGoStillKyr`, `LetGoRate`, `SpentKyr`, `HazTerm`, `HazCeil`, `DepthDecline`.

### Output

- `World.Decline`: the index, its three terms, the three running peaks, and when the bar was first crossed. Exported in the dossier with the absolute numbers beside it.
- The `KWaning` line says what it read, from the event's own parameters: "Nothing new has risen in three million years, and half of what was held is dark". The present line, which claims few still rise and those that stand are old, goes.
- `FLetGo`: a fact, a woe of the people, with the world and the remain it leaves.
- `techstats`: the occupancy table above regenerated (height, when, present, keep, still, churn, median age standing, share held by the old, largest holder), the index at the waning and at the present per seed, held over reach, and the age length. That table is the gate for every change in stage 3.
- The war measures the plan already owes — first wars per distinct pair, wars per thousand people-ticks, on twenty seeds — print here too, before and after, since the force changes how many peoples stand to fight.

## Stages

Three stages on one design, inside plan step 7 and after the random stream per people, which is what makes stages 2 and 3 near-paired comparisons rather than resamples of a heavy-tailed distribution.

**Stage 1: the index, read only.** `World.Decline` computed in a `decline` phase after `legacies` and before `hazard`; exported; techstats prints it and the occupancy table. The fertility clock still declares the waning and the end.
- *Delivers*: the index, the dossier fields, the techstats table, the tests that the index is 0 at a fresh dawn and 1 with nothing held, and that the peaks never fall.
- *Testable*: no behaviour change at all — the legends are byte-identical and `TestSameHistory` does not move. It is the one stage of this proposal that can be checked that hard.
- *Why first*: it is the instrument stages 2 and 3 are read with, and it is worth having even if the force is slow to land.

**Stage 2: the index declares the age.** The waning is read from the index and `FineFertility` goes. The end is read from the index **or** from the fertility floor as today, whichever comes first, and the floor stays until stage 3 has the force working. This matters for a practical reason: with the floor gone and nothing to end the age, three seeds in eight run to `MaxFades`, which is eight fades — an age of 130 to 240 Myr, three to seven times today's, at a cost linear in its ticks. The `KWaning` line says what it read.
- *Shifts the histories*: the waning moves everywhere and the end moves in the seeds where the index reaches the bar first, so everything after differs.
- *Testable*: the waning is not declared on one tick over the bar; a world with the same peoples standing and half its systems lost declares the waning where the fertility clock would not; over twenty seeds, how many ages the index ends and how many still fall through to the floor — that count is the measure stage 3 drives to zero.

**Stage 3: the force.** Lever 1 (the attrition) and lever 2 (the spent world) first, since they are what reaches the old; then levers 3 and 4, each measured on its own with the others off.
- *Delivers*: `FLetGo` and the remain it leaves, `Spent`, the hazard's term and ceiling, the dark age's depth term, the `DeclineTuning` group.
- *Testable*: a people never lets go of its seat or its cradle and a one-world people never lets go; letting go leaves `Stiff` and `Still` where they were; a spent world is not settled while it is spent; and the table — **the gate is that the present holds under two thirds of the height in every seed and under half in the median; that no seed has its height in the last third of its age; that the still span is under a third of the age in the median; that the age length stays inside 20 to 80 Myr; and that no run is capped.**
- *Order within the stage*: measure `SpentKyr` at zero before setting it, so the claim that the churn would refill the freed worlds is tested rather than assumed.
- *Closing the stage*: the fertility floor and `Cycle.Ends` come out last, once no seed in twenty falls through to them.

## Implementation notes

- The `decline` phase is after `legacies` and before `hazard`, so the hazard of the tick reads the index of the tick. It draws nothing.
- The births term keeps a small ring of per-tick birth counts rather than walking `w.Civs`; the held and rising terms are two walks the tick already makes and should be folded into one.
- The peaks live on the world (`PeakHeld`, `PeakRising`, `PeakBirths`) and are passed as they are.
- The attrition is the one lever here that draws, and it draws from the people's own stream, which step 7 has split before this lands. That is what lets stage 3 be tuned at all: the galaxy, the species pool and the deep pass stay byte-identical across a change to `LetGoRate`, so what moves in the table is the mechanism and not the draw order.
- `Sample func(*World)` on the config (added at step 6 for the sampling above) stays: it is how the table is regenerated and how stage 3 is tuned.
- The five-seed occupancy reading of the old draft is superseded by the eight-seed table above, which is the baseline.

## Open questions (sharpening)

- [x] **Which variables.** Held (habitable, over the running peak), rising, births in a two-million-year window over the peak rate; the mean of the three. Reach is dropped: it rises through the waning in all eight seeds, because reach is a technology. The youth of what stands is dropped: the median stiffness is 0 at the present in seven seeds of eight and bounces tick to tick in the eighth, since stiffness grows with worlds and most peoples hold one. Both go to techstats, where they read well.
- [x] **Absolute or relative.** Relative, against the age's own running peaks — an age that never had a height has not fallen from one. The absolute numbers go in the dossier beside the index, for the player comparing galaxies.
- [x] **The force.** The levers that exist are **not** enough, and the hazard reading the index is not even the biggest of them: it is already at its 2.5 ceiling through the last third of four seeds in eight, so the ceiling has to move before the term does anything. The old need an attrition of their own, because 89% to 100% of what is held at the present is held by peoples over five million years old and nothing but a war takes a world from a standing people. The attrition is lever 1, and the spent world beside it, because settlement and loss already balance at 2 to 31 stars a million years around a total that does not fall.
- [x] **Age length.** The fade is not retuned. The age length band (30 to 80 Myr today) is a check, not a target: what the force removes is the still span, which is a median 35% of the age and three quarters of seed 1, so shorter ages are the intended outcome. The gate keeps it inside 20 to 80 Myr; if the index drives ages under 20 Myr the bars move before the cycle does, since the cycle is also the deep pass's clock.
- [x] **The index's shape.** The mean of three terms, rather than the worst of them or a weighted sum. A weighted sum (0.5 held, 0.3 rising, 0.2 births) sits 0.05 to 0.14 lower everywhere and orders the eight seeds the same but for the bottom three, which lie within 0.03 of each other: it moves the bars and not the reading, so the simpler rule wins. The worst-term rule turns out to be the births term in all eight seeds at every bar — it is the fertility clock again, wearing the index's clothes.

## What a people is to remember

The three questions the plan folded into this step from the optimisation pass. They are about the telling, not about decline; they are answered here because step 7 builds both.

- [x] **How often a tale wears.** **Every fourth tick**, with each tale's rate compounded over the gap and the peoples staggered so no tick bears them all (`Config.WearEvery`, `worldgen -wear`, already in and inert at 1). Asked of the wearing on its own, with the feedback cut (`internal/history/wearcadence_test.go`), a telling put through every fourth tick comes out **within 1%** of one put through every tick, every eighth within 3%, every sixteenth 7% out. The error is all in the forgetting, because a compounded roll grants at most one step where forgetting wants three.

  **What it is worth is a few percent of a run, not the tenth the optimisation pass claimed.** The work falls as 1/n exactly — asked of the wearing alone, in one process, cadence after cadence (`TestWearCost`), every fourth tick saves 74.9% of its draws and every eighth 87.4%, the time following (152 ms, 136, 72, 38, 23 at 1, 2, 4, 8, 16). But the wearing is most of what the lore phase *draws* and only a part of what it *costs*: about 16 ns per tale-tick against the whole phase's 47 to 70, and the phase is 12.7% of a run. So every fourth tick should take roughly a fifth off the lore phase, two or three percent of a run. Two seeds run at four cadences cannot see an effect that size: the same normalised measure (the lore phase's nanoseconds per tale-tick) ranges from 27.9 to 51.2 across those runs with no trend in the cadence, while the draws in the very same runs fall 0.41, 0.13, 0.11, 0.07 per tale-tick exactly as 1/n. The draws are the part that is measured; the time is inferred from the isolated cost and is the weaker claim.

  So it is taken because it is nearly free and already built, not because it is the saving. **The lore phase's cost is its other five walks** — `reckon`, `experienced`, `revise`, `suspicion`, `loreDials` — and those are what step 7's incremental summaries are for. Drawing the number of steps from a binomial over the gap would carry the cadence further and is written down under Deferred, not built.
- [x] **Whether an heir inherits its parent's whole telling.** **Yes, unchanged.** It is what makes a line a line: an heir's grudges, its claim on the old realm and its monsters are its parent's, and two thirds of every telling in the galaxy is inherited. It is the single largest term in what a telling costs, so the temptation is real, but the cost it drives is the walking of a telling every tick, not the holding of it, and step 7's incremental summaries remove most of that walking: after them what remains per tale per tick is the wearing (now a quarter as often), the revision and the suspicion. **The trigger**: if after step 7 the lore phase is still the largest cost of a run, the change to make is not a cap but a step of wear on everything inherited (a successor remembers the empire mythically), which is already what a wall read by a stranger does.
- [x] **Whether a first meeting trades tales about third parties.** **Yes, unchanged.** It is what makes a people a monster to someone it has never met, which is 20.5% of every telling and some of the best of the legends. It is also not what makes tellings large: `tellOf` is capped at the five heaviest tales each way, only those of weight 2 or more, only those not already worn to myth, once per pair ever. Five tales per pair is a rounding error beside the hundreds a telling holds.

## Deferred

- **Late seeding of the predators**: a dormant replicator woken and a transmitter starting to speak as the index climbs. The trigger is above.
- **The ossification filter's outcome reading the index**: at a high index the near miss is the break rather than the set. The next thing to try if lever 1 is not enough.
- **The wearing's steps drawn from a binomial over the gap**, which would let the cadence go past four without the forgetting falling behind.
