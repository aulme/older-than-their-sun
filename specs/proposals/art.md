# Art and contentment: what a people makes, and what it is content with

**Status:** Draft (2026-09-22). Depends on `specs/plan.md` step 10 for the circumstances art is about, on `continuity.md` (an art is a memory that does not wear, which is a continuity mechanism) and on `leaders.md` (a leader is one of the things art is about). Read by `miracles.md`, which follows it, and by the codex step, which renders both.
**Last updated:** 2026-09-22

## Problem

**Contentment already half exists, and it can only fall.** `Civ.Morale` is moved at some twenty-five sites — a world lost, a fleet lost, a plague, a betrayal, a filter failed, a sundering — and raised at exactly two: a miracle gained (`miracle.go`) and a renaissance (`ossify.go`). It feeds two readers, Social (`levels.go`, `soc += c.Morale`) and Aggression (`dials.go`, `0.1*clamp(Morale,-2,2)`), decays toward zero at 2% a thousand years in peace, and is pinned at zero for a people whose species cannot `Wavers`, which is the unconscious. So the dial this proposal wants is already in the code, correctly absent where it should be absent, wired to two things, and structurally incapable of rising.

**Nothing a people makes is about anything.** Culture is on the tech tree — Burial, Song, Religion, Philosophy, Law, Doubt — but as prerequisites and small level, not as objects. A people's works are mines and shipyards. It has arts in the sense of techniques and none in the sense of a thing somebody made about something that happened.

**The council cannot act on discontent at all.** It appraises other peoples — whom to strike, scout, ask — and `mind.Direct` chooses what to feed first from war, fear, hunger, greed and fixation. Nothing in either reads how the people feels, so discontent could only ever be a hidden clock that runs out in a civil war after the fact. There is no decision a council could make today to remedy it.

**The legends have the gap `leaders.md` names from the other side.** Events happen and are remembered, and nobody ever says anything about them.

## Scope

**In.** Contentment promoted from a misery counter to a dial with two ends and more than two readers, and a `Hearth` category the council can spend on it in advance. Population as carrying capacity. Art as a made thing: what it is about, its potency, where it is, how it travels, how it is lost.

**Out.** Art as a tradeable commodity in `flow`. The uses inside `Hearth` beyond a first few (they are a data table, and it can grow).

## Design

### Contentment is Morale, promoted — not a second dial

Two floats both meaning "how the people feels" would have to be kept consistent at twenty-five call sites, and the existing one already carries the rule that unconscious peoples have no such thing. Three changes, then:

**A rising side.** Art (below), a war won, a world settled, a boon, a filter overcome, a pact kept through a generation, tribute arriving, and the lifespan `continuity.md` adds — a people that stopped dying young is a happier people. The present imbalance is not a design; it is what happens when a dial is added for one purpose and then only ever used to punish.

**More readers.** Beyond Social and Aggression: the civil-war roll (a contented people does not split), `empires.md`'s separatism check (a province whose people is content stays in), the dark age's depth, and the slave's revolt, which already watches `Declines` and `Seen` and should watch this too.

**A bound.** Clamped both ways, and the decay toward zero stays. The decay is what makes a golden age a golden age rather than a plateau.

### The council provisions the circuses, and it is a category in the flow

If discontent's only effect is that a civil war becomes likelier, the council never acts on it, and low contentment is a hidden death clock rather than a pressure that produces behaviour. The decision has to be one a people makes *in advance*, at a price, and can fail to afford.

The machinery is already there and is the most complete thing in the model. `flow.Category` is a priority order over what a people feeds first — `Fields, Works, Mind, Road, Arms` with nothing pressing — and `mind.Direct` reorders it from circumstance with the reasons attached (`Direction.Why()`: "arms first: threatened", "the road before the mind: greedy"). `flow.Direct` then feeds down the order and sheds the rest, reporting what went dark.

So the circuses are **a seventh category, `Hearth`**: what a people spends on itself — games, festivals, monuments, the dole, the temple. Real uses with real upkeep, placed low in the default order, fed after the works and before or after the road.

**A category rather than a use inside `Mind`**, for three reasons. It must be sheddable, and the order is exactly the machinery for what gets cut when the income runs short. Shedding it is legible: `Allocation.Dormant` already carries the phrase, so "the fleets were fed and the games went dark" writes itself. And it must *compete* — the hearth is paid out of the same income as fleets and colony ships, which is the whole point.

**Discontent moves it up.** `DirectionInput` gains a discontent term; below one bar the hearth passes the road, below a deeper one it passes the works; `Why()` gains "the hearth first: the people are angry". That is a council deciding today to provision the circuses, and the failure is the honest one — a people that cannot pay sheds it, the discontent stands, and the break that follows is a bill it could not meet rather than a clock nobody could see.

**This is where an empire's size costs it, without a shot fired.** The hearth's upkeep scales with population, which is carrying capacity (below), so a large realm pays more to keep itself content out of the same income that buys its expansion. A non-military contraction lever of exactly the kind `empires.md` wants, and the Roman one: the bigger the city, the more grain, the further the fleet that carries it.

**Physical art is a hearth use.** A monument is a `Work` with an upkeep in `Hearth` that gives contentment while it is fed. An empire in trouble lets its monuments go dark, and a monument gone dark is on its way to being a remain.

The council's *appraisal of other peoples* still only gets the tilt: discontent lowers the bar for a war of gain and raises it for a pact. That is `leaders.md`'s rule — it moves the disposition, never the decision procedure — and it stands, because the direction is the domestic procedure and the appraisal is not.

### Population is the carrying capacity, and it is already computed

There is no population *field*, but there is a quantity that stands in for one exactly: the organic matter a people's holdings yield. `worldYield` in `sources.go` is biomass by archetype — lush 9, ocean 8, superterran 7, arid and twilight and low-gravity and ice-shell 6, hothouse and floater and volcanic and dim 5 — with terraforming at 4 and comets and nebulae beside it. Over a thousand-year tick a people fills whatever its land will carry, so **population is the organic capacity of what it holds**, with no lag in either direction.

Better than counting worlds, for three reasons. It tells a garden from a rock. It already falls when a world is bombarded, which is `empires.md`'s war that spends what it is fought over — so damage to a world is damage to a population without a second mechanism. And it rises with terraforming and with the tree, which is what actually happened.

**Imported food raises the ceiling, and that is a lever.** Organic matter moves in tribute and in trade, and it feeds people standing on land that could not feed them. An imperial core fed by its provinces carries a population above its own capacity, and when the tribute stops it does not shrink gracefully — it starves, at the centre, all at once. This is expressible in the flow as it stands, it is the annona, and it is one of the better non-military contractions available to step 10.

The die-back is worth writing where the refill is not: a capacity that falls below the population it carried is a shortage, a fact, and a hit to contentment.

### An art is a made tale

A `Tale` is already: about a fact, held by a people, carrying a provenance, a slant, a wear, a weight and a rank; spread by contact; inherited by heirs; forgotten in a dark age; pruned at five hundred; rendered by the tellings. An art differs in three ways and no more — it was **made** rather than witnessed (a new `Provenance`), it carries a **potency** where a tale carries wear, and it does not degrade. Spread, inheritance, loss, the pruning and the whole rendering path come for free, and the codex gets art without a second parallel body of remembered material about the same events sitting beside the first.

The gain beyond cheapness: **an art is a memory that does not wear.** A people with much art about its own past forgets its past more slowly. That is a continuity term, not only a contentment one, and it is why this belongs in this plan rather than beside it.

### What an art is about

**One to five references**, each a fact the maker's telling already holds. The poem and the tale it was drawn from are the same event seen twice: one worn by retelling, one fixed at the moment of making — which is also the honest account of why we know what we know about the past.

**A word of circumstance tags** taken from the maker's condition when it made the thing: at war, besieged, enslaved, holding slaves, plagued, in a dark age, newly great, scarred by a filter, led, bereft of a leader, starving, far from home. A bitmask, not a list.

**Potency** rolled at making, tilted by Social, by Wisdom, by era, and by hardship. The *rate* of making scales with population — the organic capacity above — and to a lesser extent with Social and Wisdom. A set people makes **less** art, not worse art: the rate falls and the potency does not, which is the truer claim and the cheaper one.

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
- The tellings need one new shape, which is the codex step's to write.
- `flow` gains a category and uses, which is what it is for; art itself is not a commodity and does not gain a `Kind`.
- `DirectionInput` gains one term, and `Direction.Why()` one reason. Both are pure and testable in `mind` without a world.

## Open questions

1. **Where does `Hearth` sit in the default order, and what is in it?** Below the works is the obvious place, but a people that feeds its games before its industry is a recognisable kind of failing state and may be worth allowing through a trait or a fixation — `Morality.fixation()` already names a category a people is fixed on, so a people fixed on the hearth is one line. The uses themselves (the dole, the games, the temple, the monument) are a data table with upkeeps, and their calibration is the same calibration as the rest of the flow: a cradle alone should afford a modest hearth.
2. **Does potency fade at all?** "Does not wear" is clean, but it means the oldest arts are always the best attested. A slow fade past some great age may be wanted, or may be exactly the wrong thing.
3. **Art about other peoples**: a conquered people's art about its conqueror. Nothing stops it — the references are facts and a people holds facts about others — but it changes what fit means.
4. **Can an alien read it?** A fit penalty across species, or across the herd-to-solitary spectrum. `Wavers` already says an unconscious people feels nothing from art; whether a merely very different people feels less is a separate and more interesting question.
5. **Does the aftermath text say what a people was content with?** Probably yes, and it is the cheapest legibility in the whole plan.
6. **Does the population above capacity starve gradually or at once?** At once is truer to a tick this long and is the more interesting event; gradually is kinder to the tuning. The tribute-fed core is where it matters and the answer decides how sharp a lever it is.

## Stages

1. **Contentment and the hearth.** Morale's rising side, the new readers, the bound; population as organic capacity; the `Hearth` category, its first uses and its place in the order; discontent moving it up in `mind.Direct`; the tilt on the appraisal's bars. No art. *Gate:* the dial rises as well as falls across the batch; the hearth observed both fed and shed, with `Why()` naming discontent as the reason for the order at least sometimes; the civil-war and separatism rates move in the direction claimed and by an amount step 10's numbers still tolerate.
2. **Art.** The made tale, potency, references and tags, the fit gift, cultural and physical, the Find and the loot. *Gate:* the share of peoples that ever make one; the fit distribution not degenerate; an art observed outliving its maker in the hands of the people that ended it.
3. **What art keeps.** The continuity term and the small ossification term. *Gate:* tuned with step 10's numbers, not against them.
