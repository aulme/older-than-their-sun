# Simultaneous turns: the civ pass made order-free, and run on every core

**Status:** Draft 2026-09-23, written when plan step 7 measured what the parallel civ steps would actually cost and did not take them. **Not scheduled**, and deliberately not part of `specs/plan.md`'s numbered steps: it is a rearchitecture of the civ engine, not a step, and it should be taken after the plan's mechanisms have landed and stopped moving. The Design section is the specification; the Stages section is the unit of work.
**Last updated:** 2026-09-23

## Problem

`tickCivs` is 54% of `runAge` in the profile (seed 5, 200 stars, `-cpuprofile`), and the machine it runs on has ten cores it does not use. The v2 optimisation notes named the wall: "every per-people step draws from `w.R` in civ order, so civ steps cannot run in parallel without a stream per people and a fixed merge order for anything that writes shared state."

Step 7 built the streams. They were necessary and they are not sufficient, and the measurement says why.

### The pass is order-dependent

Seed 11 at 200 stars, the peoples stepped in reverse order in `tickCivs` and nothing else changed: **206 peoples and 36446 events become 136 peoples and 35265 events**, digest `2312728f…` against `2a185476…`. The streams made a people's *draws* independent of what any other people did — `TestCivStreamIsTheIdAlone` pins that — but the world a people acts on is still whatever the peoples before it in the list left behind. "The peoples may be stepped in any order" is true of the random source and false of everything else.

### The coupling is the record itself

It is not the scratch buffers. The step-7 ledger named `tally` and `regardHolder`'s stamps as what stood in the way of a parallel pass; those are the small end. The large end is `told` and `fact` (`internal/history/lore.go`), which every civ step calls and which both end in `spread`:

- `spread` sends a `Message` to every met partner, ally and hearer of both parties — a write into another people's queue;
- and for a weighty fact at a star, it calls `hold(e, f, Witnessed, …)` for every active people with a holding in watch range — a tale written directly into another people's `Lore`.

So a people's turn writes into the memory of every neighbour that can see it, before the next people takes its turn. `setOwner`, `addExpedition`, the war list, `uplift`'s spawn and `w.Civs` growing mid-loop are the same shape. A merge order fixes the order the writes *land* in; it does nothing about a people needing to *read* what an earlier people wrote this tick, which is what the pass does everywhere.

### Why this is not a tuning problem

A conflict-free schedule would need a sound footprint per people per tick — every star and every people its turn might touch — and `find`, `explore`, `spread` and the sightings make that footprint most of the galaxy for a large realm. Optimistic execution with rollback would need a snapshot of the world per tick. Both are larger than the change proposed here and neither is easier to read afterwards.

## Scope

**In.** The civ pass restructured so that a tick's peoples act on the tick's opening state and their writes land afterwards in civ order; the pass then run on a worker pool; the gate that a parallel run and a serial run of the same build are byte-identical.

**Out.** The other phases. Parallelism across ages or seeds (`techstats` already runs seeds on all cores, and that stays). Any change to what a civ step decides — the decision functions are untouched; only when their effects become visible moves.

**Adjacent, and better taken first** (they are byte-identical by construction and want none of this): the indexing the optimisation notes name — `suspicion`'s telling walk, `sortedInts` in the per-tick passes, `regard` through `warm`, the pairwise passes that grow with the square of the standing peoples.

## Design

### Simultaneous turns

Today a tick is a relay: people 0 acts on the world, people 1 acts on the world people 0 left, and so on. The proposal is that a tick is simultaneous: **every people acts on the world as it stood when the tick opened, and what they do lands in civ order once they have all decided.** That is a change to the fiction's own physics as much as to the code, and it is defensible on its own terms — a galaxy where a message takes millennia to cross is a poor fit for a turn order in which a people a thousand parsecs away has already heard.

It is also what makes the parallel run provable rather than hopeful. If no turn can read another turn's writes, the turns commute, and a worker pool is a pure speed-up.

### Reads: the opening state

A civ step reads through `w`. The proposal does not snapshot the world; it makes the mutable state a civ step can reach **double-buffered at the fields that matter**, which the measurement above says are:

- `Civ.Lore` and the message queue (what `spread` writes);
- `World.Owner` and the holdings (what settlement, conquest and loss write);
- `World.Civs` (what `uplift` and the sunderings write);
- the fleet, war, contract, pact and source lists;
- `World.Events` and `World.Chronicle`.

Each gets a reader's view fixed at the tick's open and a writer's buffer per worker. A read of the current tick's own writes — a people reading back what it recorded a step ago in its own turn — stays exact, because a worker's buffer is visible to the people that owns it.

### Writes: a merge in civ order

A turn appends its writes to its own buffer. After the pass, the buffers are applied in civ id order, which is exactly the order the serial pass would have produced them in, since a serial pass runs one people's whole turn before the next one's. Event ids are assigned at merge, not at record, so a buffered event carries a slot and the fixups (`FallEvent`, `EndEvent`, `Tale.Fact`, `factsAt`) are resolved there.

Two writes can now contradict: two peoples settling the same star in one tick. The merge resolves by civ order — the lower id lands, the higher one's write is dropped and the turn is told it failed, which is a case the steps already handle, since a settlement can fail for other reasons.

### What it costs to read afterwards

This is the real price and the reason it is not in the plan. Every civ step becomes a function over a view rather than over the world, and "what is visible when" becomes something a reader of any step has to hold in their head. The mitigation is that the split is at the field level and is few: the list above is the whole of it, and a step that touches nothing on it is unchanged.

## Stages

**1. The order-dependence, pinned.** A test that runs the civ pass with the peoples permuted and asserts the history moves — the inverse gate, so that the later stages have something that goes green when they work. Cheap, and it documents the finding above in the code.

*Gate:* the test fails on a build with simultaneous turns and passes on this one.

**2. The reads double-buffered.** The fields above given an opening view; the pass still serial. Histories move once, and the batch is regenerated.

*Gate:* `TestSameHistory` and the legends digest re-pinned and the reference batch re-read; the age lengths, people counts and war rates within the spread the plan already documents.

**3. The writes buffered and merged.** One buffer per turn, applied in civ order, event ids assigned at merge.

*Gate:* byte-identical against stage 2 — the merge alone must change nothing.

**4. The worker pool.** `Config.Workers`, `-jobs`, defaulting to one.

*Gate:* **the civs phase byte-identical between a serial run and a parallel run of the same build**, on ten seeds; `go test -race` clean; the wall-clock win reported against the profile above.

## Open questions

- **Does simultaneity change the histories for the better or only differently?** Stage 2 is a regeneration with no intended behaviour change, which is the expensive kind. The batch measurements — age length, held share, war rates, the decline index — say whether the galaxy the plan tuned survives it.
- **How much of the win survives the uneven peoples?** A tick's work is not spread evenly over the standing peoples; one empire of a hundred worlds beside thirty peoples of one. The pool should be measured before the design is committed to, with the step timings under `-phases` as the input.
- **Whether `World.Events` can stay one list.** The merge assigns ids; an alternative is one list per worker kept for the run and an index that resolves an id to a list, which costs every reader.
