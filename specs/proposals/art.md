# Art and contentment: what a people makes, and what it is content with

**Status:** Draft (2026-09-22). Depends on `specs/plan.md` step 10 for the circumstances art is about, on `continuity.md` (an art is a memory that does not wear, which is a continuity mechanism) and on `leaders.md` (a leader is one of the things art is about). Read by the codex step, which renders it.
**Last updated:** 2026-09-22

## Problem

**Contentment already half exists, and it can only fall.** `Civ.Morale` is moved at some twenty-five sites — a world lost, a fleet lost, a plague, a betrayal, a filter failed, a sundering — and raised at exactly two: a miracle gained (`miracle.go`) and a renaissance (`ossify.go`). It feeds two readers, Social (`levels.go`, `soc += c.Morale`) and Aggression (`dials.go`, `0.1*clamp(Morale,-2,2)`), decays toward zero at 2% a thousand years in peace, and is pinned at zero for a people whose species cannot `Wavers`, which is the unconscious. So the dial this proposal wants is already in the code, correctly absent where it should be absent, wired to two things, and structurally incapable of rising.

**Nothing a people makes is about anything.** Culture is on the tech tree — Burial, Song, Religion, Philosophy, Law, Doubt — but as prerequisites and small level, not as objects. A people's works are mines and shipyards. It has arts in the sense of techniques and none in the sense of a thing somebody made about something that happened.

**The legends have the gap `leaders.md` names from the other side.** Events happen and are remembered, and nobody ever says anything about them.

## Scope

**In.** Art as a made thing: what it is about, its potency, where it is, how it travels, how it is lost. Contentment promoted from a misery counter to a dial with two ends and more than two readers.

**Out.** A population number (there is none; see below). A domestic option set for the council (open question 1). Art as a resource in `flow`.

## Design

### Contentment is Morale, promoted — not a second dial

Two floats both meaning "how the people feels" would have to be kept consistent at twenty-five call sites, and the existing one already carries the rule that unconscious peoples have no such thing. Three changes, then:

**A rising side.** Art (below), a war won, a world settled, a boon, a filter overcome, a pact kept through a generation, tribute arriving, and the lifespan `continuity.md` adds — a people that stopped dying young is a happier people. The present imbalance is not a design; it is what happens when a dial is added for one purpose and then only ever used to punish.

**More readers.** Beyond Social and Aggression: the civil-war roll (a contented people does not split), `empires.md`'s separatism check (a province whose people is content stays in), the dark age's depth, and the slave's revolt, which already watches `Declines` and `Seen` and should watch this too.

**A bound.** Clamped both ways, and the decay toward zero stays. The decay is what makes a golden age a golden age rather than a plateau.

### The council is moved by contentment; it does not maximise it

The council scores other peoples: whom to strike, whom to scout, whom to ask for a pact. There is no domestic option to put a happiness goal into, and adding one is a new decision procedure, not a new dial — a build larger than the rest of this proposal together. The cheap and truer version is that discontent tilts the options already there: a discontented people's bar for a war of gain falls, which is the war fought because the streets are angry, and its bar for a pact rises. This is `leaders.md`'s rule in the same words — **it moves the disposition, never the decision procedure** — and it means unhappiness produces conquest, which is the historically common thing and the one that feeds the arc.

### There is no population, and art must not ask for one

The simulation has no population number anywhere: not on a people, not on a world. `Density` in the code is stellar density, a law of the galaxy. So "the chance scales with population" becomes **held habitable worlds and Social**, which is what everything else that wants size already reads. Naming it because two separate ideas have now leaned on this number — the other was plague rate in overcrowded imperial capitals, which became connectivity in `empires.md`.

### An art is a made tale

A `Tale` is already: about a fact, held by a people, carrying a provenance, a slant, a wear, a weight and a rank; spread by contact; inherited by heirs; forgotten in a dark age; pruned at five hundred; rendered by the tellings. An art differs in three ways and no more — it was **made** rather than witnessed (a new `Provenance`), it carries a **potency** where a tale carries wear, and it does not degrade. Spread, inheritance, loss, the pruning and the whole rendering path come for free, and the codex gets art without a second parallel body of remembered material about the same events sitting beside the first.

The gain beyond cheapness: **an art is a memory that does not wear.** A people with much art about its own past forgets its past more slowly. That is a continuity term, not only a contentment one, and it is why this belongs in this plan rather than beside it.

### What an art is about

**One to five references**, each a fact the maker's telling already holds. The poem and the tale it was drawn from are the same event seen twice: one worn by retelling, one fixed at the moment of making — which is also the honest account of why we know what we know about the past.

**A word of circumstance tags** taken from the maker's condition when it made the thing: at war, besieged, enslaved, holding slaves, plagued, in a dark age, newly great, scarred by a filter, led, bereft of a leader, starving, far from home. A bitmask, not a list.

**Potency** rolled at making, tilted by Social, by Wisdom, by era, and by hardship. A set people makes **less** art, not worse art: the rate falls and the potency does not, which is the truer claim and the cheaper one.

### What art gives, and the fit

Contentment from an art is `potency × fit`, where `fit = popcount(art.tags & mine) / popcount(art.tags)`. You have a war too, so it lands harder; you have also been enslaved, so it lands harder still.

This is the best idea here and the one most easily built expensively — done naively it is a walk over every art in the galaxy per people per tick. It is not: each people computes its own circumstance word once a tick in the civs phase, and the match is two instructions against an art the people already holds.

**Capped, and capped harder in the sum.** A bonus every people gets is inflation, not character. The spec already records this lesson in exactly this area: the first batch with the cultural nodes on the tree gave everyone 2.4 Social and quadrupled lifetimes.

### Cultural and physical

**Cultural art** is everywhere the people is, cannot be lost while the people lives, and passes on contact as any tale does. This gives the engine worth having: the empire's art outlives the empire, carried by the people who broke it.

**Physical art** is a `Work` at a star with a `Legacy` — located, lost with the star, glassed with the world, found by others through the Find as any remain is. Loot is then already built: a physical art changes hands when its world does, and is dealt with the other works in a civil war's division of the realm.

**Cultural art is not traded.** Trade runs over three fungible flow kinds; a thing that costs the giver nothing and cannot be scarce is not a trade good. It spreads free with contact, which is cheaper and the better story. Only physical art is loot.

### Not a fourth term on the break

Less art when set, plus art gives contentment, plus contentment holds a realm together, is a compounding spiral into the break. Directionally right — old empires are supposed to break — but the break already has three terms and step 10 is where they are tuned against each other. Art's contribution is small and is tuned there, with the rest, not on its own.

## Implementation notes

- **Measure before building**: over step 10's batch, what share of peoples ever make an art, arts per people at the present, and the distribution of fit. A fit that is nearly always 0 or nearly always 1 means the tag vocabulary is wrong, and that is cheaper to find before the rest.
- `Civ.circumstance()` computed once per people per tick and held, like the summaries step 7 stops re-summing.
- `Provenance` gains `made`. `prune` must not drop a people's own arts ahead of its inherited tales; the rank rule needs a term for it.
- The tellings need one new shape, which is step 12's to write.
- Nothing goes into `flow`.

## Open questions

1. **Should the council actually maximise contentment** — a domestic option set beside the appraisal of enemies — rather than only being tilted by it? The cost is a new option kind in `mind` with its own appraisal, which is a bigger build than everything else here. Recommendation: tilt only, and revisit once the arc is in.
2. **Does potency fade at all?** "Does not wear" is clean, but it means the oldest arts are always the best attested. A slow fade past some great age may be wanted, or may be exactly the wrong thing.
3. **Art about other peoples**: a conquered people's art about its conqueror. Nothing stops it — the references are facts and a people holds facts about others — but it changes what fit means.
4. **Can an alien read it?** A fit penalty across species, or across the herd-to-solitary spectrum. `Wavers` already says an unconscious people feels nothing from art; whether a merely very different people feels less is a separate and more interesting question.
5. **Does the aftermath text say what a people was content with?** Probably yes, and it is the cheapest legibility in the whole plan.

## Stages

1. **Contentment.** Morale's rising side, the new readers, the bound, the tilt on the bars. No art. *Gate:* the dial rises as well as falls across the batch; the civil-war and separatism rates move with it in the direction claimed and by an amount step 10's numbers still tolerate.
2. **Art.** The made tale, potency, references and tags, the fit gift, cultural and physical, the Find and the loot. *Gate:* the share of peoples that ever make one; the fit distribution not degenerate; an art observed outliving its maker in the hands of the people that ended it.
3. **What art keeps.** The continuity term and the small ossification term. *Gate:* tuned with step 10's numbers, not against them.
