# Leaders: named figures, where they stand, and what their loss costs

**Status:** Sharpened and built at `specs/plan.md` step 9 (2026-09-23); the open questions are answered in "As built" at the end, and the ledger row has the numbers. Depends on `continuity.md` (the succession filter reads it, and a leader's span reads lifespan) and on `specs/plan.md` step 7 for the streams. Read by `empires.md`, whose conquest wave is a leader; by `art.md`, whose arts are about leaders and whose contentment answers to their loss; and by `miracles.md`, whose delusion is built to respect the line drawn here — it corrupts a people's picture of the world and never its decision procedure.
**Last updated:** 2026-09-23

## Problem

Everything in the history is done by a whole people. There is no actor smaller than a civilisation, so the legends have no subject but "the X", every fall is the fall of a nation rather than of anyone, and the reader has nobody to blame or follow. The step-6 sampling also found a galaxy with no punctual shocks at all: occupancy settles and holds for a third to three quarters of every age, and the only ambient pressure in the model is ossification, so every empire dies the same way.

Real conquest waves — Macedon, the Mongols, the Rashidun — are the standard case of a periphery that unified, found an instrument, and went off, and the standard *end* of one is a succession the realm could not absorb. None of that is expressible.

## Scope

**In.** A named leader as a state a people can be in; where the leader physically is; what their presence does; their temperament; their death; the succession filter that follows it; the crazed immortal.

**Out.** The conditions that make a conquest wave, and the galaxy's response to one (`empires.md`). Lifespan and continuity themselves (`continuity.md`). What is made about a leader, and what their death does to a people's contentment (`art.md`).

## Design

### Leaders are rare, conditional and named

A leader is not rolled and then acts; **the conditions make one possible and, when one arises, it is named.** That ordering is what the history actually supports — Macedon's instrument and the belief that Persia was hollow both predate Alexander — and it keeps the mechanism tunable, since a leader becomes evidence about the state of the galaxy rather than a die roll on top of it.

A handful a run, not one per people per era. The name comes from the names pass, which already coins for peoples, wars, stars and plagues.

### At this tick, a leader is not always a person

The tick is a thousand years. For a short-lived kind the leader lives and dies inside one tick and what spans ticks is the consequence, so the legends tell a line or a dynasty. For a long-lived kind, for one kept through `hibernation`, and for an uploaded, hive or eldritch mind, the leader genuinely is one continuous individual across many ticks. **Both are the same mechanism**, and the view renders it by the people's profile — a conqueror and his heirs, a reconciliation of directives, a brood-line, a doctrine that took hold.

### They override the disposition, never the decision procedure

A leader contributes a term to the dials large enough to **invert** the people's own bent — a paranoid people that starts building coalitions, a peaceful one that starts conquering. That is more than the telling's capped 0.3, and it has to be, or the effect cannot exist.

But it arrives as numbers the council reads, never as a branch that skips the council. The mind still decides; it decides on very different inputs for a few thousand years. Two knobs: how large the term may be, and how strongly the leader's bent correlates with the people's own — mostly amplifying, inverting often enough to be noticed.

The mismatches are the valuable cases. A people acting against its known nature is legible to its neighbours through `Intel`, and when the leader dies the people **snaps back**, which explains an empire terrifying for one reign and inert the next, and gives the legends a closing line.

### They are somewhere

A leader is with a fleet, on a world, or on an object. Not everywhere:

- **At the front** — a large bonus to the force they are with (through `Morale`, which already feeds battle).
- **In the capital** — a small bonus across the whole realm, strategic direction communicated everywhere.

Broad and thin, or narrow and deep. So the temperaments are a shape rather than a ladder, and they line up with the two arcs: the capital-dweller is the empire temperament, the front-liner the conquest one.

**If the place is taken or destroyed, the leader dies with it.** Immortality here is the absence of natural death and an unbounded potential span; it is not invulnerability. That makes decapitation a real strategy, which is what made the Achaemenid empire fall as fast as it did, and it needs no new rule.

**Temperament governs flight as well as position**: at the front they die with the field, in the capital they die with the capital, and a coward runs — diminished rather than gone, a third state between reigning and dead. Temperament is drawn from the people's dials with noise, like the bent.

**Where a leader is should be knowledge, not fact.** Aiming at one runs through `Intel`, `Scouted` and the sighting machinery, which gives scouting a job it does not have.

### The temperament loop

This is what the mechanism is for:

- **At the front** — the bonus falls where it decides things, and they die soonest. Short reign, conquest, then a succession crisis: fragmentation.
- **In the capital** — safe, reigns long, buys little in the field. Long reign, continuity, stability, then sclerosis.

So the temperament is the choice between the two failure modes of the whole galaxy, and the distribution across a run shapes what kind of history it gets.

### Succession is a filter

When a leader is lost, the people faces it, in the codebase's own idiom for a crisis:

- **Overcome** — the succession holds.
- **Scarred** — it holds, but centralised, diminished, frightened. The most common real outcome, and currently unreachable.
- **Declined** — `breakDown`, already written: a civil war where there are parts and factions for one, a dark age otherwise.

**Difficulty is set by continuity and by how far the leader had pushed the people from its own bent.** A realm with no institution behind the person comes apart; one with deep records, long generations and an archive survives it diminished. The base is high — a lost leader should usually cost something — and the scaling decides how much.

What it inherits for nothing: the hazard term (`1.5 × (hazard − 1)`), so succession crises get harder as the galaxy declines; the per-trait difficulty table, so a hive's succession is a cut rather than a break as the Distance already is, a machine with `forking` has a backup copy, and an unconscious people has nobody to succeed; and the legends' existing rendering of a facing.

Succession is the **punctual sibling of ossification** — same three outcomes, same break function, opposite timing.

### The crazed immortal

A leader that is actually deathless (not one kept on ice) and has been awake, at war and accumulating its own crimes for long enough **narrows what it counts as wrong** — `Morality` is already "what a people counts as wrong" — until almost nothing is. The slide is gradual and readable in the chronicle before the galaxy acts on it.

It is **an unstable state, not an ending.** It wages war on everything until it is destroyed or until it runs out of enemies and turns inward; the purges raise the pressure on its own realm; and the realm eventually shatters. The purges also eat the realm's continuity — the immortal at the top, the amnesia below — so the mechanism that makes it unstable is the same one that decides how hard it falls.

A people carrying `ScarMortality` has made death sacred and will not have a deathless ruler, which is characterisation for free.

## Implementation notes

- A leader is a small record with an id, a name row, a bent, a temperament, a location that follows a fleet or sits at a star, and a span; leaders belong to the world, by id, like wars and plagues.
- The dials term sits beside `LoreDials`, which is the existing precedent for a second contributor.
- The succession filter is `def(&Filter{...})` like every other; its decline calls `breakDown`.
- A leader's arising, deeds and loss are facts, so they are in every telling and wear like anything else — the name surviving the deed, or the deed the name, and at myth collapsing into an archetype by the code that already turns a forgotten enemy into "the ones from the dark".

## As built (step 9)

The design above stands; what it left open was settled like this, and the code is `internal/history/leaders.go` with its numbers in `mind.Tuning.Leaders`.

- **What raises a leader.** Four occasions, each a state the people is already in, read in this order: an heir's first 20 kyr (the winner of a civil war, the founder of a line), 20 kyr after a dark age (a restorer), a renaissance's surge (a reformer), a war. Each has a rate a thousand years, and the chance is drawn from a stream of the people's leaders' own (`leader:<id>`), so a people's history is the same draw for draw until one rises. Only a people that has reached the stars, is its own master and has someone to follow raises one: the unconscious, the replicators and a living world cannot (a new `Leads` ability in the species registry), and nothing asleep does. The rates are set so the occasions come out near even and a run at 400 stars sees a few dozen leaders among several hundred peoples: 25.6 a world over the reference batch.
- **One at a time, and a vassal's.** One. A vassal keeps the leader it had, and raises none of its own while held; a people taken and kept as a slave loses its leader with no succession (`deposed`), since its master rules it.
- **The dials term.** Each of the eight dials gets `Size` (0.45, above the telling's 0.3) times a draw of `Corr` (0.65) times the people's own lean on it plus the rest as noise, clamped to one either way. The war decision reads a posture, not the aggression dial (the bar is `mind.Bar`, keyed on the stance), so a term on the dials alone could never have turned a peaceful people to war. The leader therefore also carries a **stance**, read off its bent on aggression: the people's own stance moved along the ladder pacifist, defensive, the middle, opportunist, conqueror by three rungs per unit of bent. `posture()` returns it while the leader reigns, and the council reads it as it reads any posture; `mind.Bar` is unchanged. **The disposition moves, the decision procedure does not.** How far the leader turned the people against itself (`Push`, nought to one) is the part of the bent against the people's own lean, plus a half when the stance crosses from peace to war or back.
- **Where it is, and what that does.** A leader is at the **front** or in the **capital**, by its dials with noise (aggression and risk less fear). At the front it rides with the greatest campaign out and adds 1.5 levels to that fleet's quality in battle; with no campaign out it stands at the first of its people's worlds fought over in the tick, adding the same to the guns and guard there. In the capital it adds 0.3 to Military and Social across the realm. At the front it falls with a broken fleet or a taken world always, in a lost battle with a chance of half the share of ships lost, and in a won one at 3%; in the capital it falls only with the capital. A quarter **run** instead, once: the bent halves, the bonus goes, and it reigns on at the seat until the next fall takes it.
- **The bonus** is levels, not `Morale`: morale does not reach battle in this model, and a fleet's quality does. The war will and the council's appetite are not touched beyond what the stance and the dials already move.
- **Its span.** A deathless leader (a people that does not turn over, and has not made death sacred) has none. Any other reigns about 3 kyr (log-normal, spread 0.5), or half its body's span if that is longer, twice as long through hibernation. A short-lived kind's leader is a **line** in the view, a founder and the dynasty after, since the founder dies inside the tick; the forms are a ruler, a line, a directive (machines), a brood-line (hives) and a doctrine (eldritch minds).
- **The succession** is faced the next time the people acts after the loss: difficulty 4.5 on Social from the table, plus 1 over it, plus 1.5 times the push, plus 0.5 per doubling of the continuity loss over the middle, less 2 with `forking`, plus everything `face` already adds (the hazard, the scars, the traits: caste and confederate ease it, the individualist and the solitary make it harder; a hive's brood-line eases it by one). Overcome, it holds. Scarred, the realm is drawn in under **centralism**, the iron scar the Distance gives, so a hard succession sets the realm faster. Declined is **`breakDown`**, the break ossification uses, under the cause `succession`: a civil war where there are parts and factions for one, a dark age otherwise; a people with no factions loses its far world instead, as the Distance does. Every succession is its own tale, carrying the leader lost.
- **When the people acts against its nature**, its neighbours read it the way they read any posture: the stance is what `hostile()` and the council see. When the leader is gone the people snaps back, and the chronicle says so.
- **Hunting a leader** is not built. Decapitation happens when a war reaches the leader's place, which it seldom does while battles are as rare as they are (six leaders fell with the field over the reference batch); nobody aims at one, so where it is stays a fact rather than knowledge. Aiming at one through `Intel` and the sightings is left for `empires.md`, whose conquest wave is a leader the galaxy has every reason to hunt.
- **The crazed immortal.** A deathless leader slides by 0.005 a thousand years at war and 0.02 a crime of its people under it. At a half its people's good narrows to the taking of worlds, at one to nothing (`morality` events, ways `narrowed` and `mad`); mad, it is a conqueror with its bent hard to aggression, hatred and fear, its purges add 0.5 to the realm's continuity loss, and once there is nobody left to fight the realm breaks under it at 10% a thousand years, which ends it (`overthrown`) with no succession after: the break is the succession. **The slide is not reversible**; death or the break ends it. A people carrying the mortality creed rotates its rulers on a span, deathless bodies or not.

## Open questions (answered at step 9, above)

- **What conditions raise a leader**, and how rare "rare" is.
- **Whether a people may have only one at a time**, and what a vassal's leader means.
- **How large the dials term may be**, and the correlation with the people's own bent.
- **Whether the bonus is `Morale` alone** or also touches the war will and the council's appetite.
- **Whether a leader can be deliberately hunted**, and what that does to the sighting machinery.
- **Whether the crazed slide is reversible**, and whether anything but death ends it.
