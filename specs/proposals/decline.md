# Decline: what the waning and the end of the age are made of

**Status:** Draft (2026-09-21). For after the plan: step 19 is the absorb; this is the first proposal of the next round.
**Last updated:** 2026-09-21

## Problem

The premise says the present is the aftermath of a declining galaxy. The sim's decline is a clock. `runAge` declares the **waning** when the cycle's fertility falls under `FineFertility` (0.5) and the **end** when it falls under `Cycle.Ends` (drawn between 0.05 and 0.2), plus a linger of up to two million years; fertility is `exp(-(now - last surge) / Fade)` with a fade of 16 to 30 Myr. `FineActive` and `EndActive` exist and are 0, so no count of civilisations enters it either. Fertility governs who is *born*: it says nothing about who is standing or what they hold, and since the ossification step nobody dies of age, so the empires of the youth sit through the whole waning holding what they took.

Measured on five seeds at 400 stars (a scratch test at plan step 18, sampled every tick; the sim settles systems, and a system has at most one habitable world, so "habitable" is the nearest thing to planets):

| seed | habitable systems | height: systems / habitable | when | present: systems / habitable | age, waning |
|---|---|---|---|---|---|
| 1 | 201 of 400 | 24% / 39% | 1.4 Myr before the present | 22% / 36% | 31 Myr, the last 19 waning |
| 2 | 190 | 44% / 55% | 42 Myr | 17% / 27% | 81 Myr, the last 61 |
| 3 | 193 | 45% / 56% | 24 Myr | 42% / 52% | 60 Myr, the last 44 |
| 4 | 195 | 31% / 43% | 42 Myr | 20% / 30% | 48 Myr, the last 33 |
| 5 | 209 | 52% / 71% | 17 Myr | 44% / 68% | 35 Myr, the last 20 |

At the height a quarter to a half of all systems and 40 to 70% of the habitable ones are held; at the present 17 to 44% and 27 to 68%. In three seeds of five the present holds 80 to 95% of what the height held, and in two the height falls inside the waning. The galaxy at the aftermath is old and full, not emptied. The legends' "THE WANING" heading and the techstats' "began in the waning" both read a line that the map does not show.

Two things are wrong with it, separately. The **definition** does not read anything a player can see (the map, the standing peoples, what they hold). And the **force** is missing: a definition read off the map would not end an age on its own, because nothing makes settlement fall once births stop. Seeds 3 and 5 would never reach half their peak and would run to `MaxFades`.

## Scope

**In.** A decline index built from a few variables the sim already has, read once a tick; the waning and the end declared from it, in place of the fertility thresholds; the pressures that make the index fall in the waning, from levers that already exist; the legends and techstats reading the index; the baseline above as the number to beat.

**Out.** New kinds of catastrophe. Changing the cycle (surges, fade, the state beneath). Anything about how peoples are born. The old-quarrel war loop (`specs/notes/war-halving.md`), which is its own case.

## Design

### The index

Decline is not one number in the fiction either, so the index is several, each on 0 to 1, each something the aftermath could show:

- **Held share**: systems held by living peoples over all systems, or over the habitable ones (the second reads better: it is the share of worlds someone could live on that someone does). Read against the age's own peak, not an absolute: `held / peakHeld`.
- **Rising share**: peoples active, not ossified and under stiffness 1 (`risingCount`, which exists) over all active peoples. This is what the fertility clock was standing in for.
- **Births**: peoples born in the last few million years over the age's peak rate. Fertility, but read off what happened instead of the clock.
- **Reach**: the sum of held systems' reaches, or the share of systems within anyone's reach; a galaxy where the old empires stand but cannot cross to each other is declining whatever it holds.
- **Youth of what stands**: the median stiffness, or the share of active peoples that are remnants, ossified, asleep, or held by another.

A first setting: `decline = 1 − mean(held/peak, rising share, births/peak)`, the reach and the youth terms held back until the batch says the three are not enough. The waning is declared when the index crosses a bar (say 0.4) and stays over it for a while (a million years, so one bad tick does not declare an age); the end when it crosses a higher bar (0.75) plus the linger, or when `MaxFades` says so, as now. `FineFertility`, `EndFertility` and `EndFertilityLow` go; `FineActive` and `EndActive`, which were never used, go with them.

Which variables to combine and how to weight them is the open question this proposal exists to settle, and the batch decides it: the index should be low through the youth, climb through the waning, and be high at the present in every seed, and the age lengths should stay where they are (30 to 80 Myr) or the cycle's numbers move to keep them there.

### The force

A definition read from the map needs the map to move. Nobody dies of age, and that stays; what falls in the waning should fall for reasons the legends can tell. The levers that exist:

- **The hazard** (`updateHazard`: the law's, plus eaten worlds, transmitters, tithes and the thinning wall) already makes every filter harder by 1.5 per point over one. It could carry a term for the decline index itself, so a late age faces its filters harder, which is the Long Dusk of v0 done through the filters rather than as its own roll.
- **What eats and what speaks**: the replicators strip worlds and the transmitters take listeners; both are rarer in the waning than in the youth because their makers are. A deep-pass-like late seeding (a dormant replicator woken, a transmitter starting to speak) as the index climbs would put the predators where the fiction wants them.
- **Ossification**: stiffness grows with stillness; the waning is still by definition. The dark age's depth (`DepthBase`, `DepthStiff`) could read the index so that a late-age fall is a deep one, and a shattering more likely than a renaissance.
- **The state beneath**: the wall's thinning already raises the hazard; leaks and the transmitter come through it. The waning could thin it faster.
- **Attrition of the old**: the one lever that does not exist. Something that ends a stiff, still, ancient empire that nothing else has touched: not age, but the sum of what it has stopped doing. The ossification proposal's true ossification is the nearest thing and ends nothing by itself.

Which of these, and how much, is the second half of the open question. The rule from the plan holds: tune named numbers first, change the design only when the batch cannot be tuned into range.

### Output

The legends' "THE WANING" heading stands but is declared from the index; a line when it is declared that says what it read ("Nothing new has risen in three million years, and half of what was held is dark"). Techstats prints the index at the waning and at the present per seed, the held share at the height and at the present (the table above, regenerated), and the age length; the gate for any change here is that table.

## Implementation notes

- `World.Decline` (the index and its parts) computed in a `decline` phase after `civs`; `runAge` reads it in place of `fertility()` for the waning and the end; `Cycle.Ends` kept for the state beneath if anything else reads it, else dropped.
- `peakHeld`, `peakBirths` on the world, kept as they are passed.
- Tunables in one `DeclineTuning`: the bars, the hold time, the weights, the hazard term.
- Tests: the index is 0 at the dawn of a fresh world and 1 with nothing held and nobody rising; the waning is not declared on one tick over the bar; a world with the same peoples standing but half the systems lost declares the waning where the fertility clock would not have.
- The five-seed occupancy reading above is the baseline; rerun it (the scratch test can be made a techstats line) before and after.

## Open questions

- **Which variables.** The three named first, or more; whether the held share is over all systems or the habitable ones; whether reach belongs in it.
- **Absolute or relative.** Against the age's own peak (every age declines) or against the galaxy (a small age never had a height to fall from). Relative reads the fiction better; absolute is what a player comparing two galaxies would want.
- **The force.** Whether the levers that exist are enough once the hazard reads the index, or whether the old need an attrition of their own.
- **Age length.** Whether the cycle's fade is retuned to keep ages at 30 to 80 Myr once the end is not the fade's.
