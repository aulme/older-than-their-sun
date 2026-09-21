# The war halving at step 16 — an open case

**Status:** open; for step 19. Recorded 2026-09-20 after step 16 landed (`1e913f6`).

## What was seen

The ten-world batch at 800 stars (`go run ./cmd/techstats -seeds 10 -out reports/tech`), step 15 (`d6a9bfc`) against step 16 (`1e913f6`):

| | step 15 | step 16 |
|---|---|---|
| peoples a world | 464 | 383 |
| standing a world | 32 | 28 |
| wars a world | 1020 | 437 |
| civil wars a world | 46 | 33 |
| facings (stiffness) | 677 at 2.27 | 563 at 2.27 |
| ages | unchanged | unchanged |

The war causes tell more than the total (`reports/tech/report.md`, the Cause table):

| Cause | step 15 | step 16 |
|---|---|---|
| the old quarrel | 6033 | 1751 |
| opportunity | 892 | 63 |
| revenge | 711 | 161 |
| the sundering | 757 | 536 |
| the embargo | 359 | 169 |
| conquest | 240 | 97 |
| a border | 189 | 275 |
| extermination | 282 | 613 |

So it is not a flat scaling: **opportunity** fell fourteen-fold and **revenge** four-fold, the old quarrel (a second war between the same two, mostly grudge-fed) three-fold, while border wars held and extermination doubled (xenophobes hate every eldritch). On seed 11 alone at 200 stars: "the old quarrel" 6033 → 1751 is the batch, but seed 11's "a border" went 232 → 7, "opportunity" 892 → under 97, "now hold N systems" lines 36 → 5 — fewer colonies, so fewer borders to quarrel over on that seed.

## What was ruled out

- **The species draws.** Running step 16's history with `species.Options{Legacy: true}` (the old kind table, old numbers) still gives the drop. So it is not the eldritch or the new cradle shares by themselves.
- **The species package.** Running step 15's history over step 16's species package still gives the drop — i.e. the change is in `internal/history`, not `internal/species`.
- **Seed 11 as the yardstick.** Old seed 11 at 200 stars had 326 peoples, a cascade outlier; old seeds 12 and 13 give 135 and 159. Part of the seed-11 colony drop may be seed noise. The batch numbers above are across ten seeds and are not noise.
- **Runtime.** Seed 11 at 200 stars is 87 s both before and after; seed 7 went 267 → 330 s with a flat profile. Not a pathology in the loop.

## What was not finished

A debug test (`TestDbgExpand`, deleted before the commit) that spawned a plain opportunist people at star 0 with a short known list and watched `systems`/`reach` for 3000 ticks showed "systems 0, reach 0" from tick 500 in **both** trees — the test's own setup was wrong (a people on Sol with a hand-made node list never goes starfaring), so it proved nothing. A fixed version should spawn as `realm`/`runSmall` in `kinds_test.go` do: on a livable non-Sol star, with a full era-3 known list, and compare colonies and wars old vs new over the same seed.

## Leads, in the order to try them

1. **`resent` as the only grudge write** (`ossify.go:206`): every grudge write in expedition, contract, lore (`takeToHeart`), pact, sighting, slight, sunder, weapon and war goes through it now; it is a no-op for `!HoldsGrudges` and for `by <= 0`. Revenge and the old quarrel are grudge-fed. Check that no former write passed a negative or zero `by` that mattered, and that `HoldsGrudges` is on for every plain biological chain (the unconscious is the only one meant to lack it). Diff the grudge sums per people old vs new on seed 12.
2. **The opportunity bar.** `appraise.go`: strength `Ships += bodyGuns`, and `NoShips = standing == 0 && bodyGuns == 0`. `bodyGuns = int(HomeDefence*(1+Mil)+0.5)` — if a plain biological profile composes a nonzero `HomeDefence` (check `Compose`'s `mulOrAdd` and the substrates' defaults), every people reads as armed and the opportunity path never opens. This alone could explain 892 → 63.
3. **`expandMul × Profile().Expand`** (`civ.go`): `Expand` was never applied before step 16; now it multiplies. If a plain chain composes `Expand` below 1 (a modifier's 0.1 leaking through `Compose`, or a zero treated as 0 rather than 1 somewhere), colonies fall and border wars with them — the seed-11 "now hold N systems" 36 → 5.
4. **The wall.** Eldritch tears (the wound, the unmaking's 0.3) raise the hazard; `sim.go` prints `wall %.2f` in the debug line. Compare wall and hazard traces on seed 12 old vs new.
5. **The tithe** (`sources.go`, `w.tithed(c, in)` before Gross) takes 10% of income within the reach of a tithe holder; at 4467 tithed peoples over ten worlds it is broad. Turn `Kinds.Tithe` to 0 with `-tune Kinds.Tithe=0` and re-batch.
6. **`disturbed`/sleep**: 162 wars from "the disturbing of its sleep" are new, not missing; ignore unless the sleepers hold worlds that others would have colonised.

Cheapest first check: `go run ./cmd/worldgen -seed 12 -stars 200 -ai | grep -c opportunity` on both trees, then lead 2.

## Addendum, step 18 (2026-09-21): the war counts are one loop's

Step 18's batch had 11989 wars over ten worlds, 9611 of them in seed 5, and step 17's 4862 had 1132 in seed 1 and 951 in seed 3. In every batch since step 15 the count is dominated by a handful of pairs fighting the same war hundreds of times: at step 15 one pair 2121 times, at step 17 pairs of 552 and 497, at step 18 five pairs of 300 to 500 in seed 5 alone (`Nekhuutran` and `Gaundnaivur`, two pacifist heirs of one sundering, 500 wars "over the old quarrel" between 17.2 and 24.8 Myr ago, each 2 kyr long, ending in "peace" with nothing taken, the next declared as soon as the truce ran out). Every such pair is heirs of a sundering or a shattering with claims on each other's worlds, and the cause is always "the old quarrel". So the war-per-world number is not a measure of anything but how many of these loops a batch happened to roll, and the halving at step 16 may be nothing more than which seeds rolled a loop. For step 19: (a) count wars with the loop pairs removed (pairs with more than twenty wars) before comparing steps; (b) look at why the loop does not end: the grudge that `sunder.go`'s claims feed, the truce (5 to 15 kyr) shorter than the grudge's decay, and `yield` giving "peace" with nothing to take instead of a truce long enough to let the grudge fade; a pacifist heir should not be declaring at all, so the declaration comes through the claim, not the posture.
