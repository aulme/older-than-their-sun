# One tick

*Absorbed into `DESIGN_NOTES.md` § Simulation v2, "One tick and the mind" (plan step 19, 2026-09-21). The Design section below stays the specification where the notes are silent; the notes win where they differ.*

**Status:** Implemented (step 1 of specs/done/plan.md, 2026-09-17)
**Last updated:** 2026-09-17

Read by every other draft: [ships-and-garrisons](ships-and-garrisons.md), [fleet-interception](fleet-interception.md), [resources-and-trade](resources-and-trade.md), [contracts-and-mercenaries](contracts-and-mercenaries.md), [wisdom](wisdom.md), [morality](morality.md). It is assumed by all of them from now on.

## Problem

The current age runs at two grains: twenty thousand years a tick in its youth, one thousand in the waning, switching when few enough peoples are active and the galaxy is barren enough. The engine is one and the rates are given per thousand years and scaled to the tick, so the two passes are meant to be the same history at two resolutions. They are not quite. A campaign of a few light years launches, crosses, fights its battles and comes home inside one youth tick; a scout's whole tour is one tick; a war between neighbours is declared, fought and ended between two council meetings; a muster, a withdrawal, a siege have no ticks to happen in. Sixty-five random checks in the history package are per tick and not per thousand years, so they fire twenty times more often in the waning than in the youth, and everything tuned on the fine pass runs differently on the coarse one. Every proposal in this folder has had to say what a rule does "in the coarse pass", and the answer is usually "it does not matter, a tick refills anything".

The reason for the coarse pass was cost. It no longer holds.

## Scope

**In.**
- One tick of one thousand years for the whole current age, dawn to present: the waning's tick, everywhere. The youth and the waning are the same engine at the same grain.
- The waning stays as an event: the line is logged and `Waning` is set when the condition is met, and nothing else changes.
- A convention for rates: every rate in the code and in the proposals is per thousand years and goes through `chance` or `count`. Nothing is "per tick" except a thing that happens once a tick by construction (a council, a battle).
- What the other drafts must re-read.

**Out.**
- The deep pass (the ten-million-year substrate before the dawn) and the myth of earlier ages: untouched.
- Finer than a thousand years, and a variable tick: see the open questions.
- Re-tuning any rate. The change should make the batch look like the waning looks now, stretched over the youth; anything that then looks wrong is a separate fix.

## Design

`MidStep` and `FineStep` become one `Step` of 1000 years, the waning's tick now. `runAge` runs at it from the dawn; the waning test stays and logs its line and sets `Waning`, and no longer changes the step. `dt` is 1 for the whole age, so a rate per thousand years and a rate per tick are the same thing, which is why this grain and not a finer one: everything in the code and the drafts is written per thousand years already.

Measured on the current tree by building with the step overridden, three seeds, one line of stats each:

| Tick | seed 1 | seed 2 | seed 3 |
|---|---|---|---|
| 20 kyr youth, 1 kyr waning (now) | 8 s | 6 s | 6 s |
| 1 kyr throughout | 12 s | 14 s | |
| 500 yr throughout | 22 s | 41 s | 30 s |

A thousand years throughout costs about twice what the two passes cost now, against a budget of three minutes. Five hundred was measured too and fits, and is held back for the open question: the cost grows faster than the tick count between the two, which says the per-tick work of wars and fleets is what is being bought. Age length, fade, the count of peoples and their lifespans, remnants and standing peoples all stayed within seed noise across the three settings.

What it does to the drafts:

- **Ships and garrisons.** One battle per tick at a world, so a siege is several ticks and a withdrawal and re-appraisal have somewhere to happen. Docks build at one ship per tick. The silo and grid odds are re-run at the new cadence in that draft. Its "in the coarse pass a tick refills anything" lines go.
- **Sightings and interception.** A crossing of a few light years at slow drives is now several ticks with a sighting on each, so a picket has a tick to report and a fleet a tick to turn. The flat muster was already dropped by ships-and-garrisons.
- **Resources and trade.** Flows are recomputed each tick as levels are and need no change; "per tick" in its text means per recomputation. Its two per-tick chances (the tap's glare at 1%, the Manna loose at 0.0005 and rising at 0.001) are read as per thousand years.
- **Wisdom.** The retry chances written "per tick" (0.02, 0.05) are per thousand years.
- **Contracts and mercenaries.** Flow terms are per tick as resources' flows are; the "each tick a running guard or strike stands against" check is per tick by construction, once a millennium.
- **Morality.** Nothing per tick; its rates are already per thousand years.

## Implementation notes

- `Config`: `MidStep`, `FineStep` become `Step Year` (1000). `FineActive`, `FineFertility` stay as the waning's condition. `DefaultConfig` and the `techstats` config change with it.
- `runAge`: one `w.dt = float64(cfg.Step) / 1000` before the loop; the waning block keeps its log and `w.Waning = y` and drops the step change.
- The 65 unscaled `w.R.Float64() <` checks: leave them; they now run at one cadence. Any that are meant as a per-kyr rate and were tuned on the fine pass are already right; any tuned on the coarse pass will show in the batch and get `chance`.
- Batch: re-read ten seeds before and after against the stats line and `techstats`; the acceptance test is that nothing moves outside seed noise except things that count ticks (battles, sightings, council actions), which should match the waning and grow twenty-fold against the youth.
- Tests: the age runs at one step; `Waning` is set and logged once; a rate given per kyr fires the same expected number of times over a fixed span as it did in the waning.

## Open questions

- **Finer still.** Five hundred years is measured and fits the budget; it would give a crossing of a few light years at a slow drive a few ticks instead of one or two. It is held back because everything is written per thousand years and `dt = 1` keeps the code and the drafts saying the same thing. If the fleet rules turn out to want it, the change is one number and a re-read of the batch.
- **A variable tick.** The old scheme was one; a tick that shortens when a war is on and lengthens in a dead galaxy would be another. It would bring back "what does this rule do at the other grain", which is what this removes. Not proposed.
- **The stats line.** Some numbers in `techstats` are counts per tick (battles, sightings). They should become per thousand years so runs at different steps compare.
