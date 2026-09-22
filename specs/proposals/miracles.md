# More miracles, and the things on the list that are not miracles

**Status:** Draft (2026-09-22). Reads `specs/plan.md` step 7 (the per-people stream is what makes the evil eye a decorator rather than a rewrite) and step 10 (Gehenna's conversion drive is an arc mechanism). Rendered by the codex step.
**Last updated:** 2026-09-22

## Problem

There are eight miracles. The frame is tight and worth stating before adding to it, because most of a list of new ones will not fit it:

1. **A miracle is held by a people**, by one of five routes — born to it, leapt to from a deep spine of the tree, found in what an earlier people left, wielded while a found thing lasts, or an eldritch people's own deepening. Nothing that cannot be gained one of those ways is a miracle.
2. **It is either causal or an object.** A causal miracle reaches into the state beneath and **wears the wall for everyone in the field**, per thousand years, whether or not its holder ever uses it; a thin wall wakes sleepers, starts transmitters and makes every miracle's filter bite harder. Nobody in the field is told. An object miracle makes a thing with forms, yields and levels.
3. **Five of the eight carry a filter of their own** — the ansible's drift, the brood, the unmaking, the chorus, the sight — so holding one is also a way to be destroyed, and `miracleDiff` makes every *other* filter harder for a people that holds any.
4. **It is the dominant fact about whoever holds it.** That is a statement about rarity: sixteen miracles are not twice as interesting as eight, they are half as dominant each.

Point 2 is the budget nobody sees. Five new causal miracles at the present wear rates thin the wall much faster, everywhere, for everyone. Either the rates come down or the galaxy gets considerably stranger — which may be wanted, but it is a choice and not a side effect.

## Scope

**In.** Eight new miracle rows and their filters. The three things on the list that are hazards, worlds or legacies rather than powers, placed where they belong. The wall budget.

**Out.** The Remnant (below: its own proposal, unscheduled). Anything that needs a second galaxy state.

## Design

### The ones that are miracles

**The evil eye** — curse one people at a time: every roll it makes is made twice and the worse taken, and while the curse stands the holder's own rolls are made twice and the better taken. *Causal.* This is the best fit on the list and the cheapest, because step 7 gives every people its own random stream: the eye is a decorator on two streams and touches nothing else. It is also the only power here that is **about** luck, which this simulation is made of, and the only one whose effect the legends can describe without describing a mechanism. *Filter:* the eye is a thing that looks both ways.

**Skip** — every world, fleet and structure of the people vanishes and returns a million years later, the intervening ticks not taken. *Causal.* Worth more to this project than it looks: step 6 found that the galaxy's problem is that it stops, and a people that steps out and returns arrives with perfect continuity into a galaxy that has forgotten it — its tellings exact, its grudges against the dead, its claims on worlds now held by strangers. It is a continuity mechanism wearing a miracle's clothes, and the tick loop can already skip a people. *Filter:* what a million years does to a thing that did not experience them.

**Echoes** — a temporary fleet doubled out of copies of the ships in it. *Causal.* Cheap, and a war lever step 10 can hand to a conquest wave: the marginal polity with the instrument nobody can match, and the instrument expires. *Filter:* what comes back is not quite what went.

**Empathy** — on communicating with a people, know its character, its grudges, its miracles and every crime it ever committed. Not causal; it is a sense. Interesting out of proportion to its cost because of what it meets: the whole intel, slant, lie and monster-reckoning apparatus exists to model peoples misjudging each other, and this is the people that cannot. *Filter:* knowing what everyone has done, and what they think of you, is not a gift — it is the thing that makes a people withdraw.

**The barrier** — walls in space that cannot be breached, only routed around unless the owner opens them; and what they keep out includes radiation, heat and blasts. *Causal.* The only power on the list that changes the **geometry**, which is why it is the most work: reach, hops, lines, interception and the blast rules all have to ask it. Worth the work — a galaxy where one people can close a corridor is a different galaxy, and it is a defensive power in a model whose only structural pressures are offensive.

**The leviathan lure** — call a leviathan to a known place; it swallows what is there and delves back; a chance it comes to the caller's seat instead. *Causal.* The backfire is the filter, built in, which is the shape every miracle here should have. Needs the leviathans (below).

**Mind control** — compel a people to make a chosen decision. This has to earn its row against `chorus`, which already takes a people at contact: memetic engineering is *a thought that takes root*, and this would be *one decision forced, from a distance, at will*. Different in kind, and a genuinely new lever because it reaches into the council rather than the disposition — which is the one thing `leaders.md` says a leader may never do, so it should be the miracle that can. *Filter:* a mind that has been in other minds.

**Memetic camouflage**, read as: the anti-memetic stops being a species modifier and becomes a miracle held by the route `born`. This is a refactor of a subsystem that already exists (`antimemetic.go`, `gap.go`, `perceives`, `Antimemetic Resilience` on the tree) and not a new power, and it is the right shape — the anti-memetic is already the dominant fact about whoever holds it, and `born` is already a route. It also opens the routes the modifier could not have: a people that *finds* the anti-memetic, or wields it and loses it. Care needed, because every use of `perceives` is in the blast radius.

### The one to cut

**Uplift** — triple research, triple wisdom. It fits the frame and tells no story: it makes a people better at what it was already doing, and the spec already records what happens when a bonus has no shape (the cultural nodes on the tree gave everyone 2.4 Social and quadrupled lifetimes: inflation, not character). Uplift is also a thing that is *done to you by someone else*, which the model already has in `Uplifts` and `Sire`. Recommendation: not a row; if it is wanted, it is what a patron does.

### The ones that are not miracles

Nobody leaps to a planet from a deep spine of the tree. These break rule 1, and each has a home that already exists.

**Gehenna** is a **world**, at most one per galaxy, claimed rather than gained: a world nothing should live on, whose holder is wholly immune to it and to the same conditions everywhere, whose survival and military are lifted while it is held, whose tech prerequisites the world lacks are moot, rich in iron and energy so ships are cheap there, and — the part that matters — whose holder **knows how to make more**, as a construction project needing no tech, and wants to. It is the best thing on the list after the evil eye, because it is the only one that is an *engine*: it changes what its holder wants and gives it a programme. That makes it step 10's business, not a data row.

**Eden** is the other world, and it is a wall where Gehenna is an engine: perfect, unattackable because nobody can bring themselves, immune to disease and to cosmic threat and to other miracles, enormous carrying capacity. Every clause is an immunity, and step 6's finding is that this galaxy's disease is stasis. Its one engine is the constraint — a people incapable of cruelty or deceit, growing vast, able to do nothing with it: a helpless paradise the galaxy orbits and cannot touch, and a prize that transforms whoever finally works out how to take it. Take Gehenna first; Eden needs that constraint to be real, and it collides with `Morality`, which already decides what a people counts as wrong.

**Leviathans** are a **cosmic hazard**, beside the sleepers, the transmitters and the elders: colossal things that surface out of the state beneath, swallow fleets, structures, moons or worlds, and delve back; sighted naturally almost never, and most often during a great battle. Cheap if they are an event at a battle, and they give the lure something to call.

**The mirror** is an **event**, not a power: a people finds an exact copy of itself on an identical world at a star that was not there, histories identical down to the individuals, each certain it is the real one. Nearly free — it clones a `Civ` and its telling, which `prune` already bounds — and it is the purest piece of dread on the list.

**The fleet of the elders** is a **legacy**: a found thing, wielded, no upkeep, no crew, a fixed and very high military level, a chance to resist other miracles. That is exactly what the Find already does with wielded artifacts and what the fields already do with salvage. A form, not a miracle.

**The Remnant** — a sealed volume where one or more peoples are still in a previous age, fully teched, holding miracles, unable to leave, the sky inside looking as it did then, nothing crossing outward and everything crossing inward — is the most evocative thing on the list and the most expensive by a wide margin. It needs a second galaxy state and a boundary rule on every mechanism that crosses space: reach, fleets, sightings, trade, tales, plagues, the Find. It is a proposal of its own. Written down here so it is not lost; not scheduled.

### The wall budget

Six of the eight new rows are causal. At the present rates that is roughly a tripling of what a miracle-holding galaxy does to the wall, and the wall is shared. Three ways to pay, and the choice should be deliberate: lower the wear per miracle so the total holds; keep the rates and accept a galaxy where the beneath leaks much more (sleepers waking, transmitters starting, every miracle's filter biting), which is a legitimate and interesting answer; or make the new causal ones rarer than the old. Measure first — what the wall reaches by the present, over the batch, before and after.

## Implementation notes

- Most of this is rows in `data/miracles.json` and a filter each in `miracle.go`; the filters are the work, not the powers.
- **The evil eye waits for step 7.** Before the per-people streams it is a rewrite of every call site; after them it is a wrapper on two streams.
- **The barrier** is the only one that needs a design pass of its own before it is costed: `Near`, the hops, the lines, the interception and the blast rules all have to ask it.
- **Memetic camouflage** is a refactor with a wide blast radius (`perceives` is called from everywhere) and no new behaviour; it should not be bundled with the new rows.
- Leviathans and the mirror are events; Gehenna and Eden are worlds placed at generation; the elder fleet is a legacy form. None of them is a miracle row and none should be filed as one.

## Open questions

1. **How many is too many?** Eight to sixteen halves what each one means. A defensible answer is that the new ones are rarer — found and wielded rather than born and leapt — so a galaxy still holds about as many miracles as it does now, spread over more kinds.
2. **Which way to pay for the wall** (above). It wants the measurement before the argument.
3. **Does mind control reach the council or only the disposition?** Reaching the council is the whole point and is also the first thing in the model that overrides a people's decision procedure. `leaders.md` deliberately refuses that power to a leader; giving it to a miracle is consistent, but it is the line being crossed on purpose.
4. **Is Eden's constraint real enough to carry it?** If incapable of cruelty or deceit does not actually stop it doing anything it wants to do, Eden is a pile of immunities and should wait.
5. **Does Gehenna's conversion drive belong to step 10's tuning?** It is a programme that makes its holder expand for a reason that is not conquest, which is the sort of thing that step's arc wants, and the sort of thing that step's gate would notice.

## Stages

1. **The cheap rows and their filters**: the evil eye, echoes, empathy, skip. *Gate:* each observed held, used and failed at; the wall measured before and after.
2. **The hazards and the events**: leviathans, the lure, the mirror, the elder fleet as a legacy form. *Gate:* a leviathan sighted at a battle; a lure that backfired; a mirror pair both believing themselves real.
3. **The worlds**: Gehenna, then Eden if its constraint survives question 4. *Gate:* a Gehenna holder observed converting a second world; the arc measures of step 10 not degraded.
4. **The barrier**, after its own design pass. *Gate:* a corridor closed and routed around in the record.

Unscheduled: memetic camouflage as a refactor; the Remnant as its own proposal.
