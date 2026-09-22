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

**In.** Eight new miracle rows and their filters. The wonders (unique, claimed, no route), the hazard, the event and the legacy form. The Remnant as a people under three rules. The wall budget.

**Out.** Uplift, cut. A second galaxy state, which nothing here needs.

## Design

### The ones that are miracles

**The evil eye** — curse one people at a time: every roll it makes is made twice and the worse taken, and while the curse stands the holder's own rolls are made twice and the better taken. *Causal.* This is the best fit on the list and the cheapest, because step 7 gives every people its own random stream: the eye is a decorator on two streams and touches nothing else. It is also the only power here that is **about** luck, which this simulation is made of, and the only one whose effect the legends can describe without describing a mechanism. *Filter:* the eye is a thing that looks both ways.

**Skip** — every world, fleet and structure of the people vanishes and returns a million years later, the intervening ticks not taken. *Causal.* Worth more to this project than it looks: step 6 found that the galaxy's problem is that it stops, and a people that steps out and returns arrives with perfect continuity into a galaxy that has forgotten it — its tellings exact, its grudges against the dead, its claims on worlds now held by strangers. It is a continuity mechanism wearing a miracle's clothes, and the tick loop can already skip a people. *Filter:* what a million years does to a thing that did not experience them.

**Echoes** — a temporary fleet doubled out of copies of the ships in it. *Causal.* Cheap, and a war lever step 10 can hand to a conquest wave: the marginal polity with the instrument nobody can match, and the instrument expires. *Filter:* what comes back is not quite what went.

**Empathy** — on communicating with a people, know its character, its grudges, its miracles and every crime it ever committed. Not causal; it is a sense. Interesting out of proportion to its cost because of what it meets: the whole intel, slant, lie and monster-reckoning apparatus exists to model peoples misjudging each other, and this is the people that cannot. *Filter:* knowing what everyone has done, and what they think of you, is not a gift — it is the thing that makes a people withdraw.

**The barrier** — walls in space that cannot be breached, only routed around unless the owner opens them; and what they keep out includes radiation, heat and blasts. *Causal.* The only power on the list that changes the **geometry**, which is why it is the most work: reach, hops, lines, interception and the blast rules all have to ask it. Worth the work — a galaxy where one people can close a corridor is a different galaxy, and it is a defensive power in a model whose only structural pressures are offensive.

**The leviathan lure** — call a leviathan to a known place; it swallows what is there and delves back; a chance it comes to the caller's seat instead. *Causal.* The backfire is the filter, built in, which is the shape every miracle here should have. Needs the leviathans (below).

**Delusion** — not a decision forced, but the inputs to it corrupted: the victim decides freely, by its own procedure, on a picture of the world that is a lie. This is better than the compulsion it replaces on every count, and it is native to the model. `Civ.Intel` is what a people believes about each other people — the level seen with the noise of the seeing, the guard in the sky, the guns, the ships standing everywhere, the relief, where and when the look was taken, and whether a plague was raging in them — and it is exactly what the appraisal reads before the council strikes, scouts, watches or asks. `look()` already builds one with noise in it. A delusion is a look with a lie in it.

So the deluded people attacks the strong believing them weak, cowers from the weak believing them strong, and quarantines a people that was never sick — that last being the sharpest, since a false `Sick` reaches the suspicion and quarantine machinery and turns a people against a neighbour without a shot. It never touches the council, which keeps `leaders.md`'s line intact: **the disposition and the picture may be moved; the procedure may not.** And it is better fiction, because nobody is a puppet — they are wrong, and they are wrong in their own voice.

A deeper second form is available if wanted and should be decided separately: planting a false *tale*, which the telling already supports through `Blamed` and `scapegoat`. That is a lie about the past rather than about the present, and it is the one that survives the miracle's holder. *Filter:* a people that makes pictures for others stops being able to tell which of its own are its own.

**Memetic camouflage**, read as: the anti-memetic stops being a species modifier and becomes a miracle held by the route `born`. This is a refactor of a subsystem that already exists (`antimemetic.go`, `gap.go`, `perceives`, `Antimemetic Resilience` on the tree) and not a new power, and it is the right shape — the anti-memetic is already the dominant fact about whoever holds it, and `born` is already a route. It also opens the routes the modifier could not have: a people that *finds* the anti-memetic, or wields it and loses it. Care needed, because every use of `perceives` is in the blast radius.

### The one to cut

**Uplift** — triple research, triple wisdom. It fits the frame and tells no story: it makes a people better at what it was already doing, and the spec already records what happens when a bonus has no shape (the cultural nodes on the tree gave everyone 2.4 Social and quadrupled lifetimes: inflation, not character). Uplift is also a thing that is *done to you by someone else*, which the model already has in `Uplifts` and `Sire`. Recommendation: not a row; if it is wanted, it is what a patron does.

### Wonders: the ones that are not miracles

Nobody leaps to a planet from a deep spine of the tree, so these break rule 1. Three of them are one thing and want a name: **wonders** — unique, at most one per galaxy, placed at generation or left by an elder, **claimed rather than gained**, with no route into them at all. The word is already half in use for this: `species/generate.go` calls the deep pass that seeds past ages "where the wonders and the leftovers come from". The other two are not wonders either; they are a hazard and an event, and they already have homes.

**Gehenna** is a wonder and the best thing on the list after the evil eye, because it is the only one that is an *engine*. A world nothing should live on, whose holder is wholly immune to it and to the same conditions everywhere, whose survival and military are lifted while it is held, whose missing prerequisites are moot, rich in iron and energy so ships are cheap there — and whose holder **knows how to make more**, as a construction project needing no tech, and wants to. It changes what its holder wants and hands it a programme it can execute. That makes it step 10's business as much as this proposal's.

**Eden** is the other wonder, and its stasis is the point rather than the objection. It is the calm eye: perfect, unattackable because nobody can bring themselves, immune to disease, to cosmic threat and to other miracles, enormous in carrying capacity, and its people incapable of cruelty or deceit. Nothing else in this galaxy persists — that is the finding step 6 was written about, and every other mechanism in this plan exists to make things end. Eden is the one fixed point, and a fixed point is worth having precisely because everything around it moves: it is the thing every age's legends can agree on while agreeing on nothing else, the coordinate the rest of the record is read against. Past ages are myth in this model, one tick per rise and fall, so Eden can be in those myths too at almost no cost, which is what makes it *the* constant rather than this age's constant.

**Leviathans** are a **cosmic hazard**, beside the sleepers, the transmitters and the elders: colossal things that surface out of the state beneath, swallow fleets, structures, moons or worlds, and delve back; sighted naturally almost never, and most often during a great battle. Cheap as an event at a battle, and they give the lure something to call.

**The mirror** is an **event**: a people finds an exact copy of itself on an identical world at a star that was not there, histories identical down to the individuals, each certain it is the real one. Nearly free — it clones a `Civ` and its telling, which `prune` already bounds — and the purest piece of dread on the list.

**The fleet of the elders** is a **legacy form**: a found thing, wielded, no upkeep, no crew, a fixed and very high military level, a chance to resist other miracles. Exactly what the Find already does with wielded artifacts and the fields with salvage.

### The Remnant, at a tenth of the cost I first put on it

I costed it as a second galaxy state with a boundary rule on every mechanism that crosses space. It is not that. It is three rules on a people:

- **It cannot leave its systems.** Nothing launches outward: no colony, no campaign, no scout. The model has the shape already — `Aloft` and `Rested` are states that change what a people may do with its ships, and `presenceCouncil` is the council of a people that does not launch.
- **It can receive matter and cannot send it.** One condition in `sendGoods`. It may be traded *into*, and pays in the only thing that crosses: what it knows.
- **It is born to the whole tree and one or more miracles.** `born` is already a route.

The different sky is flavour, written once. What comes out of the three rules is better than what I imagined I was buying: a people of enormous knowledge, permanently harmless, that can be visited, learned from, besieged — and that can never retaliate anywhere but at home, where it is formidable. An oracle you may rob. That is a cost like the nomad's or the parasite's, not like a second galaxy, and it belongs in the stages.

### The wall budget

**Decided: the wall ends up roughly where it is now.** The new rows pay for themselves rather than the galaxy paying for them.

The arithmetic is not the obvious one, and it is worth writing down so the tuning is not done by guess. Miracles are **era-4 nodes on the tech tree**, each behind its own two prerequisites, so this is not a fixed pool being redistributed: more rows means more peoples whose spine happens to reach one, and holdings go up. Two other channels move with it — the `born` route, and elder artifacts, of which `miracleShare` says two in five stand for a miracle rather than an art, drawn from the miracle list.

So it is one number solved by measurement, not by argument: take the batch, count **miracle-holdings summed over the age** and the wall at the present before and after the new rows, and set the new causal ones' wear so the second matches the first. If the holdings rise more than the wear can absorb without making the new miracles trivially thin, the fallback is the other lever — make the new causal ones harder to reach rather than cheaper to hold.

## Implementation notes

- Most of this is rows in `data/miracles.json` and a filter each in `miracle.go`; the filters are the work, not the powers.
- **The evil eye waits for step 7.** Before the per-people streams it is a rewrite of every call site; after them it is a wrapper on two streams.
- **The barrier** is the only one that needs a design pass of its own before it is costed: `Near`, the hops, the lines, the interception and the blast rules all have to ask it.
- **Memetic camouflage** is a refactor with a wide blast radius (`perceives` is called from everywhere) and no new behaviour; it should not be bundled with the new rows.
- Leviathans are a hazard and the mirror an event; Gehenna and Eden are wonders placed at generation; the elder fleet is a legacy form; the Remnant is a people. None of them is a miracle row and none should be filed as one.
- **Delusion writes `Intel`**, which is a struct `look()` already builds; the work is the choosing of the lie, not the writing of it.

## Open questions

1. **How many is too many?** Eight to sixteen halves what each one means. A defensible answer is that the new ones are rarer — found and wielded rather than born and leapt — so a galaxy still holds about as many miracles as it does now, spread over more kinds.
2. **Which way to pay for the wall** (above). It wants the measurement before the argument.
3. **Does delusion also plant false tales, or only false looks?** The false look is this stage's work and is bounded. The false tale outlives its planter and reaches the grudges, the monster reckoning and the dials, which is a much larger thing to let loose.
4. **Does Eden appear in the myths of every past age, or only stand in this one?** Every age is the stronger claim and nearly free, since past ages are one tick each; it is also a commitment that Eden is older than anything and was never made.
5. **Does Gehenna's conversion drive belong to step 10's tuning?** It is a programme that makes its holder expand for a reason that is not conquest, which is the sort of thing that step's arc wants, and the sort of thing that step's gate would notice.

## Stages

1. **The cheap rows and their filters**: the evil eye, echoes, empathy, skip, delusion. *Gate:* each observed held, used and failed at; a deluded people observed striking or sparing on a picture that was false; the wall measured before and after and the wear solved so it lands where it started.
2. **The hazards and the events**: leviathans, the lure, the mirror, the elder fleet as a legacy form. *Gate:* a leviathan sighted at a battle; a lure that backfired onto its caller's seat; a mirror pair both believing themselves real.
3. **The wonders and the Remnant**: Gehenna, Eden, and the Remnant's three rules. *Gate:* a Gehenna holder observed converting a second world and its wanting measurable in the dials; Eden present, untaken and named in the myths of an earlier age; a Remnant people traded into, taught from, and never once launching.
4. **The barrier**, after its own design pass. *Gate:* a corridor closed and routed around in the record.

Unscheduled: memetic camouflage as a refactor, which has a wide blast radius and no new behaviour.
