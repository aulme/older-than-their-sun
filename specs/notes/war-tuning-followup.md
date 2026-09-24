# War: research for the follow-up tuning (steps 10 and 11)

**Written:** 2026-09-24, when step 11 was parked.
**Read first:** `specs/proposals/war.md` (the design, the gate, "As built" for stages 0 to 2, "Tuning left for later"), `specs/notes/war-stage1-handover.md` (the tools and the loops of stage 1), `specs/notes/war-halving.md` (how to read war counts).

## Where things stand

- **On master**: stages 0, 1 and 2 of `war.md` (steps 10 and 11's first half), last commit `990b31b`.
- **On the branch `war-stage3-wip`** (`c36110c`, unpushed, not merged): stage 3, clients and the cold war, built and tested but parked with the gate short. Its "As built (stage 3)" draft is in that branch's `war.md` (the gate table is a placeholder). The branch does not re-pin `TestSameHistory` or the legends digest; landing it means re-pinning both, regenerating `reports/tech`, and the plan row.
- The user parked the work here on 2026-09-24 and asked for these results to be kept for a follow-up tuning, together with the other variable tuning work (the rows listed at the end of `war.md`).

## A new definition of a world war (the user's, not yet built)

The gate reads a world war as a system of wars joined by a shared people with eight peoples or more at war at once on three fought fronts. That counts peoples, so tiny peoples that join a cascade and fight nothing weigh as much as empires, and a quiet seed can never have one (below).

**Proposed instead:** a system of wars is a world war when, at its widest moment, **more than half the worlds in the galaxy are affected** — held by a people at war in the system, whether it attacks, is attacked, or both. To build it in `internal/warshape`: at each war's start inside a system (the moments `systems()` already reads), sum the worlds held then (`worldsAt`, folded from `world_held` and `world_lost`) by every people at war in the system, and divide by the worlds held by every people then; the widest share is the system's. The gate row then reads world wars by that share, and whether "most seeds" still fits is to be seen on forty seeds. The fought-front condition may still be wanted, so that a web of pacts declaring and never fighting does not count.

## What the batches taught about reading the gate

- **One batch of twenty seeds is not enough.** Stage 1's own rules on seeds 1–20 and on 21–40:

  | | seeds 1–20 | seeds 21–40 |
  |---|---|---|
  | old enemies | 127 pairs, 17 seeds | 66, 15 seeds |
  | loop pairs | 0 | 3 |
  | conquest waves | 53, 5 seeds | 1, 1 seed |
  | side declared on strikes back | 16% | 10% |
  | long wars in most seeds | 8 of 20 | 16 of 20 |

  Every rule change renumbers every history, and single seeds swing from 900 wars to 75. Read a row on seeds 1 to 40 (`WAR_SEEDS=40`, about 40 minutes on ten cores) before keeping or dropping a change for it; a difference smaller than the gap between these two columns is not a result.
- **World wars and the census.** `cmd/warshape -census` counts the peoples that send fleets. In most seeds between 12 and 34 peoples launch a campaign in the whole age, and no more than 10 to 27 send fleets of any kind in their busiest million years; the crowded seeds (38 to 100 senders in a million years) are where the world wars are (5 of 6 in one batch). No cascade rule makes eight peoples at war at once where there are twenty in the age: hence the new definition above, or `empires.md` filling the quiet seeds.
- **Joined wars dilute the conduct rows.** A war joined by pact is a war in the rows "fought", "short/long" and "strike back", and allies fight few of theirs. In stage 2 joined wars were 15% of wars and a third fought; on the stage 3 branch, where a war's end unwinds the wars joined to it, they are 24% of wars and 9% fought, which is most of why "fought" fell to 50% there. Whether those rows should read wars joined by pact apart is a question for the follow-up.

## Stage 2: the variants tried (twenty seeds each, seeds 1–20 unless said)

| variant | wars | fought | short / long | old enemies | world wars | border | strike back |
|---|---|---|---|---|---|---|---|
| stage 1 | 6561 | 64% | 35 / 28 | 127, 17 seeds | 27, 6 seeds | 25% | 16% |
| A: rivalry, escalation to the whole at the third war, the alliance's worth | 6511 | 58% | 38 / 28 | 133 | 15, 6 | 17% | 11% |
| B: A, and allies coming with relief at war in their own name | 5189 | 57% | 37 / 27 | 80 | 11, 4 | 18% | 10% |
| C: A, relief allies not at war; loops cut (hunts, buy-offs) | 5569 | 56% | 38 / 26 | 104 | 9, 5 | 18% | 12% |
| D: C without rivalry or the alliance's worth (tuned off) | 7633 | 65% | 38 / 24 | 143 | 38, 8 | 27% | 13% |
| E: C without the escalation to the whole | 7194 | 59% | 37 / 30 | 188, 18 | 24, 5 | 18% | 15% |
| F: the whole only for a declarer twice the other's size; allies stand with their principal | 5446 | 61% | 40 / 25 | 145 | 21, 3 | 18% | 13% |
| redress met at one world (the lost fleet fixed) | 5054 | 58% | 45 / 18 | 98 | 7, 3 | 28% | 10% |
| redress met at two worlds | 3304 | 63% | 29 / 33 | 71, 14 | 10, 1 | 11% | 16% |
| **landed**: redress asks back what the last war lost (1 to 3) | 5633 | 60% | 42 / 20 | 122, 18 | 32, 6 | 23% | 13% |

What they say: the escalation to the whole ended the rivalries it was meant to carry; the rivalry bar is what raises old enemies; redress met at one world makes the rivals' wars over as they begin (fought redress wars 58% short, 8% long), at two for all it fights them out and a third of the wars go with the peoples they kill. Long wars are the row stage 2 made worse on both samples (28% to 20%, 26% to 17%).

## Stage 3 (the branch): four batches of forty seeds

| batch | loops | proxy wars | cold wars, with the build-up | patron came | fought | long | ages | notes |
|---|---|---|---|---|---|---|---|---|
| 3a | 50 | 35, 8 seeds | 75 pairs, 13 seeds | 30% | 73% | 11% | 14–60 | clients passed between powers; punitive wars on repeat; cascades of empty joined wars |
| 3b | 154 | 108, 8 | 65, 13 | 6% | 84% | 8% | 17–70 | a client's home taken and spared, a free win, five hundred times |
| 3c (39 seeds) | 2 | 15, 4 | 68, 13 | 20% | 65% | 19% | 10–70 | clients' homes protected; patrons always hear their clients |
| 3d | 0 | 18, 7 | 111, 14 | 21% | 50% | 16% | 10–71 | wars come off worst remembered; patrons' will raised |

Stage 3's rows at 3d: proxy wars pass (five or more in three seeds or more); old enemies 298 pairs in 30 seeds; conquest waves 36 in 8 seeds; cold wars with the build-up in 14 of 40 seeds (most wanted); attacks on clients by a smaller people 32% (under a fifth wanted) and the patron coming in 21% (most wanted); standing tributes set at 0.13 / 0.15 / 0.17 (surrender 0.19, offer 0.15, sought 0.13), reviews moving them both ways, a third of the ticks owed paid short.

**Open on the branch**, for the follow-up:
- **Patrons seldom come.** Of 500 attacks on clients the patron joined 21 wars and sent ships to about 60, and abandoned its client 100 times; most of the rest never reached its council because the war was over before the call arrived. Levers: the call answered at the declaration rather than by message, `Client.Base` and `Back`, a patron's spare ships.
- **Cold wars in most seeds.** The build-up is read when each names the other its rival; 111 of 449 cold pairs have it. The rival is the most feared neighbour in reach, which is seldom mutual; a rival kept for longer, or read one-sided, would change it.
- **Seed 28's age ends at 10 Myr** on the branch (35 to 45 Myr under stages 1 and 2): its galaxy peaks at 3.2 Myr and the decline index ends it. Not traced; the tribute taken from vassals' incomes and the levels a vassal no longer gives its master are the first suspects.
- **Seed 22 runs 71 Myr with 2074 peoples** (1240 heirs of sunderings, 601 believers) and takes 41 minutes at forty seeds in parallel, over the run-time row; a profile shows contract brokering and plague suspicion hot, neither stage 3's own code.
- **The tribute paid short a third of the time** is higher than "mostly in lean ticks" wants; `Client.PayFirst` and where the tribute stands in the flows are the levers.

## The loops, all found with the tools and each a test

Stage 2 (on master): a master hunting its own anti-memetic slave for the settlers it lost at the slave's world; a hole hunted every 150 thousand years as a fading wariness allowed (`Civ.HuntsFailed`); a neighbour bought off twenty times (a war bought off is a yield; bought off twice, it offers itself as a vassal); settlers sent for a million years to an unseen people's home (a lost ship makes a dread star); a muster outliving its war and opening the next; a claim struck twenty-nine times under a wariness capped short of certainty (`War.WaryStop`); a lost fleet's people holding one world seven times, whose heirs shattered onto nothing.

Stage 3 (the branch): a patron punishing its client every two thousand years; a client passed between two powers 135 times; allies joined against an ally's war fighting on alone and again at every call; a client's home taken and spared, five hundred times; old enemies back every hundred thousand years as the fading wariness let them under `WaryStop` (`Civ.Worsted`, remembered).

The pattern worth remembering: **a threshold on a quantity that fades lets the thing it stops come back at the rate of the fading.** Three of these loops were that; a count that does not fade is the fix each time.

## The tools, for the follow-up

- `WAR=1 WAR_SEEDS=40 WAR_OUT=dir WAR_SAVE=gate.json [WAR_BASE=old.json] [WAR_LIMIT=50m] [WAR_TUNE=G.F=v,...] go test ./internal/history -run TestWarShape -v -timeout 4h`: the batch, the gate with a verdict per row (beside an earlier one), every seed's record kept, a seed over the limit or panicking reported and the rest read. `WAR_LIST=1,5,9` runs chosen seeds.
- `go run ./cmd/warshape [-base gate.json] [-pairs 20] [-systems 3] [-events civ,from,to] [-census] dir`: the same report and gate from kept records in seconds; the pairs past a count war by war; the widest systems of wars; one people's chronicle over a span; how many peoples send fleets. Every loop above was read this way.
- `go run ./cmd/scenario [-seed N] [-tune G.F=v] spec.json`: a small world told as its councils reason; the specs in `internal/history/testdata/scenarios` are tests on five seeds.
- `WAR_WATCH=dir` runs the watchers inside the batch (a pair past ten wars, masters changing, endless wars) with the councils' reasons.
