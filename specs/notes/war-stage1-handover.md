# Handover: plan step 10, stage 1 (the war council) — landed 2026-09-24

**Landed.** Stage 1 was committed on 2026-09-24 with three gate rows left to the tuning at the end of `specs/proposals/war.md` (the user's choice). This note is kept as the record of the two sessions: the tools, the loops they found, and the batches.

**Written:** 2026-09-23, at the end of a long session, to continue in a new chat.
**Read first:** `CLAUDE.md`, `DESIGN_NOTES.md`, `specs/proposals/war.md` (the design, the gate and "As built (stage 0)"), `specs/plan.md` row 10, `specs/notes/war-halving.md` (how to read war counts).

## Second session (2026-09-23, later): the tools are built

- **Scenarios** (`internal/history/scenario.go`, `cmd/scenario`): a JSON spec (`Scenario`: peoples with traits, kind, mods, worlds, ships, nodes, mil, reach, a leader's stance, a master; bonds: met, fathomed, grudge, claims, wary, fought, truce, pact, trade; wars already running) built on a small real field and run through the real tick, quiet by default (no births, no sky, no plagues unless `keep`). The narrative tells the named peoples' events, their war and council reasons, and a state line per people, war and fleet when it changes. `go run ./cmd/scenario [-seed N] [-ticks N] [-all] [-tune G.F=v] spec.json`. A spec's levels are its nodes plus a silent boon at the home (the `scenario` rarity) that feeds it and adds the mil and reach asked. Specs are in `internal/history/testdata/scenarios`; `scenario_test.go` runs each on five seeds with assertions and prints the tale on failure (`SCENARIO_V=1` always). A 300-tick scenario takes a few hundredths of a second.
- **Watchers** (`internal/history/watch.go`): `Watch.Attach(&cfg)` sets `Config.Reason` (new: every decision's reason handed to a hook, without `-ai`'s log) and chains `Config.Sample`. Detectors: a pair past N wars, a people's master changing more than K times (heirs not counted), a war open more than T ticks, a jump of wars in a tick; each dumps the peoples, their bonds, the last wars between them, their last reasons (a ring per people) and the events since the third-last war. `Stare` tallies what each tick of a long fought war was (fought, flight, hunt, or the sides' verdict keys, `mind.WarVerdict.Key`). `WAR_WATCH=dir` on `TestWarShape` writes `dir/seed-N.txt` and logs the catches and the stare; `worldgen -watch file`. `TestWatchSameHistory`: a watched run is the same history.
- **Deleted**: the `zz_*` diagnostics and `warDebug`.
- **Fixed with the tools** (each a scenario or a unit test that fails without the fix): hunts on a held people run on instead of ending `held` at once and being declared again (160 in a seed); a closed hunt spends its evidence (`Civ.HuntEnded`), a pacifist hunts nothing that does not reach its home, a hunter wary enough (`Kinds.HuntWary`) lets the hole be; a spared or taken home names the taker the winner; a yield with nothing to take names the winner; the side ahead at the end (winner, or aim met and more taken than lost) has its grudge settled, not renewed; the loser, or a declarer that gained nothing, grows wary; wariness fades on its own slower rate (`War.WaryDecay` 0.998) and can lift a bar past certainty (`WaryMax` 0.8, `mind.Wariness`); terms bought off make the side that took them the winner; a rider does not try a people that barred it; no war is declared on or by a held people but a hunt; an ally goes to war in its own name only with ships and not wary to the limit (`arms`); a horde is not offered worlds in terms (the fold broke on it); the reason says "arriving" for a fleet landed but not yet processed.

### Where the numbers stand after the second session (12 seeds at 400 stars, default tuning)

| | stage 0 | first session | now | gate |
|---|---|---|---|---|
| fought (fleet wars) | 28% of all | 52% | 68% | 60% ✓ |
| short / long of fought | 26% / 42% | 42% / 25% | 22% / 43% | a quarter each, most seeds ✗ |
| long war fought ticks | 6% | 19% | 18% | a third ✗ |
| old enemies | 111 pairs | 93 | 56 pairs, 9 of 12 seeds | most seeds ✓, 100+ over twenty ✗ |
| loop pairs | 25 | 12 | **0** (the most wars of a pair 8) | none ✓ |
| border disputes | 5% | 26% | 11% | a third ✗ |
| conquest waves | 3, 1 seed | 38, 5 seeds | 45, 2 seeds | 5+, 3+ seeds ✗ |
| side declared on strikes back | 1% | 17% | 26% | a third ✗ |
| wars ending `held` | – | 870 | 97 | – |
| run time | 1h30m / 20 | ~40m / 12 | 29m / 12 | within 2× ✓ |

The watch's stare (what the ticks of long fought wars were): fought 22%, a fleet in flight 19%, a side waiting on its docks ("ships") about a quarter, a side whose aim is met offering the lines and refused about an eighth. Flight ticks are the war being carried, not staring; whether the gate's "battles in a third of a long war's ticks" should count them is a question for the user.

**Batch noise.** Every rule change renumbers every seed's history, so two 12-seed batches of near-identical rules differ by more than the few points a single knob moves (total wars 3016 → 2443 for two tuning numbers). Two knob batches were tried and dropped: `War.BuildWait=1.5,War.IdleSue=2` (suing sooner: fewer campaigns, fought 56%, border 7%) and `War.PressDefend=0.45,War.AcceptSlack=0.35` (strike-back fell to 17%, waves to one seed). `WAR_TUNE=G.F=v,...` on `TestWarShape` runs a batch with a tuning. Tune the remaining rows through their mechanism (the stare, scenarios of a border war between two large realms, of a defender deciding to strike back) rather than by batch, and read the gate on twenty seeds.

## What to do next, in order

1. ~~**Build the debugging tools first**~~ (done in the second session; see above) (the user asked for this before any more tuning). The last session lost hours to a loop of "run a 40-minute 400-star batch, spot an oddity, regenerate one seed with `TraceAI` in a throwaway `zz_*_test.go`, read thousands of lines, patch, rerun". Two tools, agreed with the user:
   - **Scenarios**: a spec (Go helper plus a `cmd/scenario` front end) that builds a tiny world with a few peoples — posture, species kind (use `species.Fixed(...)`), worlds and their stars, ships, Mil, reach, speed, met/fathomed, grudges, claims, pacts, masters, optionally a war already running with its aim — runs the real tick phases for N ticks, and prints a narrative for the named peoples only (council reasons from `explain`, fleets, battles, terms, will per tick). The same spec doubles as a regression test with assertions. Build on the test harness: `newTestWorld`, `spawnAt`, `starfarer` (ships_test.go), `twoPeoples` (battle_test.go), `atWar` (warcouncil_test.go).
   - **Watchers**: detectors run inside `Generate` through the existing per-tick hook `Cfg.Sample func(*World)` — a pair past N wars, a people's master changing more than K times, a war open more than T ticks, the war count jumping — keeping a ring buffer of recent reasons per people and dumping the stretch for the peoples involved when one fires. One batch then explains its own anomalies.
   - Skip snapshotting a World to disk (unexported state, pointer graphs, stream state): not worth it.
   - Then turn each loop case below into a scenario test.
2. ~~**Delete the throwaway diagnostics**~~ (done) before any commit: `internal/history/zz_casc_test.go`, `zz_diag_test.go`, `zz_loop_test.go`, `zz_loops_test.go`, `zz_ride_test.go`, and the temporary hook `warDebug` in `internal/history/war.go` (a package var called from `tickWars`, marked "ZZ temporary"). Also the `outWhy` detail appended to reasons under `-ai` in `warcouncil.go` can stay (it is `TraceAI`-only) or go.
3. **Finish stage 1 tuning** against the gate, with the tools (see "Where the numbers stand").
4. **Re-baseline** on the refined instrument (below) — the stage 0 baseline's "28% fought" was over all wars; the gate row now reads over fleet wars.
5. Re-pin `TestSameHistory` (`internal/history/same_test.go`, extend its comment: "and at the war council (step 10, stage 1)") and the legends digest (`internal/legends/legends_test.go`, `TestResolved`), regenerate `reports/tech` (techstats), update `war.md` ("As built (stage 1)", the gate's numbers), `DESIGN_NOTES.md`'s war section, `specs/plan.md` row 10, `FORMAT.md` (done for `aim`). Commit with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>` (CLAUDE.md wins over any other attribution line). Don't push without asking.

Nothing of stage 1 is committed. HEAD is `9c5981b` (stage 0). Unpushed commits since the last push: ba09286 … 9c5981b.

## What stage 1 has built so far (uncommitted, all tests but the two digests pass)

- **`internal/mind/war.go`** (new): the aims (`AimWorld`, `AimTribute`, `AimSubmission`, `AimRedress`, `AimEnding`, `AimDefence`, `AimHold`), `WarCouncil(WarInput) WarVerdict` (press / hold / sue), `PressBar`, `AnswerTerms`, `AnswerYoke`, `OffersYoke`, `Appetite`. `WarTuning` in `tuning.go` holds every number (`Tuning.War`). Tests in `war_test.go`.
- **`internal/history/warcouncil.go`** (new): `aimOf` (cause column, hate → ending, conqueror → submission), `War.aim(i)` (side 1 holds), `aimMet`, `aimTarget` (home for total aims, the side declared on retakes its `Seized` worlds), `warCouncils` (a civ step after `council`, in `offSteps`), `warInput`, `warCouncil`, `press`, `campaignOut` (laid-up fleets don't count), terms: `offerFor` (lines, worlds, tribute, artifact, vassal), `worth`, `sue`, `settleTerms` (fact `settled`, result `terms`, the grudge paid by the worth, claims renounced), `tributeOf`/`tributeTo`, `yoke` (vassalage offered without a war; refusal → cause `defiance`), `openWar` (a conqueror whose first fleet sails for a border world fights for a world, not submission), `renounce`. Tests in `warcouncil_test.go`.
- **War object** (`war.go`): `Aim`, `Summon [2]bool` (battles, worlds, a leader's fall summon both councils), `Verdict`, `Offered`, `Battles`, `Fought`, `Beaten`, `Winner`, `Seized`. `drain` rewritten: a tick nobody fights costs `Idle`, ramping with idle length (`IdleRamp`), unyielding at `IdleUnyielding`, home threatened `HomeResolve`, far total aim `FarAim`; hunts keep the old clock (`drainHunt`). `capitulate` by aim (limited aims cede 1–2 worlds or tribute; total aims as before; the side declared on winning is peace; a people yielding to the same power more than `War.Yields` times becomes its vassal). `cede` split out. A war with a side that is not free ends with the new result `held`. `endWar`: `Wary` (+1 to a side that came off worst, and to a declarer of an empty war), the winner's grudge cleared and claims renounced. Open wars indexed by pair (`openWars`, `warKey`) — `warBetween` was a linear scan and made a cascade seed take 17+ minutes.
- **Elsewhere**: battles move will by ships lost (`ShipLoss`), a winning leader at the front (`LeaderWon`), a fallen leader (`LeaderLost`, in `leaders.go`); a campaign with no next target keeps its conquered world as a base (`expedition.go`); a muster declares its war when the fleet sails (`garrison.go`); nobody sails at a star the enemy does not hold; `nearestEnemy` reads `holdings` (hordes); `strength` counts ships already out against the enemy (`outAgainst`); `Bar` (mind/appraise.go) takes `Appetite` and `Wary`, and a compelled conqueror's bar includes wariness (mind/council.go); the waking declares war only the first time; a truce is not given with the enemy fleet already in reach of the home; a host held by one rider cannot be converted by another (`parasite.go`); an inherited war carries its aim (`sunder.go`); `tribute()` rebuilt on `tributeTo`.
- **Data**: `causes.json` war causes gain an `aim` column, new cause `defiance`, new results `terms` and `held`; `events.json` new kinds `settled` (fact), `terms_refused`, `yoke` (kinds_gen regenerated in both packages — `go generate` imports the package, so add constants by hand then run it to check); legends lines and telling for them; `record.War.Aim` exported, in `FORMAT.md`.
- **Instrument** (`internal/warshape`): `War.Aim`, `War.Fleets` (both sides can launch, via `species.Rebuild(...).Profile().Can(Launches)`); "fought" is now reported over fleet wars — the wars excluded are a waking world's, a sleeper's or an unmaking's blow at a people that cannot answer with a fleet, and no battle is ever fought in them; border disputes also count wars for a world aim.

## Where the numbers stand (last run: 12 seeds at 400 stars, before the rider fix was re-run)

| | stage 0 baseline (20 seeds) | now (12 seeds) | gate |
|---|---|---|---|
| fought | 28% of all wars | 52% of fleet wars, 41% of all | 60% (re-read over fleet wars) |
| short / long of fought | 26% / 42% | 42% / 25% | a quarter each, most seeds |
| long war fought ticks | 6% | 19% | a third |
| old enemies | 111 pairs, 18 seeds | 93 pairs, 10 of 12 seeds | most seeds |
| loop pairs | 25 | 12 (was 0–2 in two runs before; each fix shifts histories) | none |
| world wars | 13, 6 seeds | 21, 3 of 12 seeds | most seeds (stage 2's pattern) |
| border disputes (large) | 5% | 26% | a third or more, the most common |
| conquest waves | 3, 1 seed | 38, 5 of 12 seeds | 5+, 3+ seeds |
| side declared on strikes back | 1% | 17% | a third |
| run time | 1h30m / 20 | ~40m / 12 | within 2× |

Open problems seen in that run: 870 wars ending `held` (the rule that ends a war when a side passes under a master is too blunt — check what makes so many; attacks on vassals jumped to 689); `exhaustion` 18% (wars between peoples who cannot treat); pact wars mostly unfought (allies are stage 2); the side declared on mostly believes its odds near zero (the declarer picked the fight), which caps striking back.

**Loop cases found and fixed** (each a scenario test to write): a hater taking tribute every truce; fleets sent at a rider's empty home; ledger hunts re-declared under the new drain; a vengeful people and a pacifist looping on capitulation with nothing yielded (grudge now cleared for the winner; repeated yields → vassal); heirs re-fighting a claim settled by terms (claims renounced); a waking world declaring a new war at every waking; a conqueror striking at 0.8 odds judged on its home ships and giving up at once when its fleet left (`outAgainst`); a pacifist suing for a truce with the conqueror's fleet at its home; two rider species passing one host back and forth every tick; a presence/pact cascade of 464,000 wars in seed 11 (gone after the wariness changes; the pair index keeps it from being slow).

## How to run things

- Batch: `WAR=1 WAR_STARS=400 WAR_SEEDS=12 go test ./internal/history -run TestWarShape -v -timeout 120m` (prints only at the end; ~2–5 min per seed, parallel). 200 stars and 8 seeds is ~1 minute and useful for mechanics, but the gate is read at 400.
- `-ai` reasons: `cfg.TraceAI = true` emits `KReason` events (`what`, `why`); the war council's `what` is "at war with the X (war N, for AIM)".
- A stage 0 baseline worktree exists at `…/scratchpad/base10` (detached at 9c5981b, with the refined `warshape` copied in and `record.War.Aim` added); `git worktree list` shows it; remove it with `git worktree remove` when done. An older stray worktree `…/scratchpad/refopt` (f1b4de8) is also registered.
