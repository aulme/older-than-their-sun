# Resources and trade

**Status:** Implemented (stages 1 to 3 at plan steps 4 and 5, stages 4 and 5 at step 6; see the Means bullet in `DESIGN_NOTES.md`)
**Last updated:** 2026-09-17

Assumes [one-tick](one-tick.md): every rate below is per thousand years, and "per tick" means a thousand-year tick, the same thing.

## Problem

Nothing in the sim is a quantity that can be gained or lost. Structures give flat level bonuses and a people raises them at random; a Dyson swarm is a research multiplier. Trade is a boolean with a 1.15 research bonus, cut by war at no felt cost. Colonising, fleets and research draw on nothing, so expansion has no reason beyond the dials, a rich world is not worth a war, and Overshoot is a level test named after a collapse that cannot happen. The galaxy has black holes, magnetars, nebulae, belts and remnants that only feed the ambient laws. The design notes promise resource exhaustion as a collapse cause and outposts and swarms as constructions, and neither exists.

## Scope

**In.**
- Three commodities as per-tick flows with upkeep: organic matter, energy, metal. Nothing is saved between ticks.
- Sources: worlds, belts, giants, stars, structures, cosmic objects, elder remains, and two made objects, one model for all of them.
- Harnessing structures behind research nodes, with a guard against catch-22s.
- A directed allocation: each people fills its uses in an order set by temperament and state, and sheds the rest. Known-but-dormant tech as a state between working and dark age.
- Fleets, scouts, surveyors, colony ships and structures as reservations of flow.
- Rarities: binary access, natural, elder and made, immobile or mobile, with grants, flat bonuses and yields. Wielded artifacts become mobile rarities.
- Two miracles that are objects: the Ember (energy) and the Manna (organic matter), each instance with a rolled form and a literal consequence. A species can be its own Manna.
- Trade as surplus sharing between partners, positive-sum by construction, with willingness, transport caps and dependence.
- Taking sources by force; nomads grazing.
- Facts, log lines, tellings and a `techstats` section for all of it.

**Out.**
- Carrying capacity and population. The commodities define the number (spare organic matter for organic kinds, spare energy for machines), but what it does is a later proposal.
- Prices, money, stockpiles, routes with freighters, per-world building queues.
- Exhaustion of natural sources. One elder source wears down as a flavour exception; nothing else does.
- Changing the ambient laws. Metals, exotic and the rest stay as background multipliers.

## Design

### Commodities

Three kinds, abstract units, per tick, recomputed like the levels:

| Kind | Symbol | Feeds | Capacity for |
|---|---|---|---|
| organic matter | O | biology nodes, medicine, ecologies, living ships, terraforming | standard, swarm, evolver, planetary mind, parasite on organic hosts |
| energy | E | energy, computation, exotic and society nodes, structures, fleets, research speed | machine-born, mind-riders on machine hosts |
| metal | M | industry, weapons and propulsion nodes, structures, fleets and ships | none |

Each tick a people has an **income** per kind, a set of **uses** each with an upkeep per kind, and an **order**. Uses are filled in order until the income runs out; what is left is **surplus**, which trade may move and is otherwise lost. A use that is not filled is **dormant** this tick: a node gives no levels, no reach and no envelope; a structure gives nothing; a launch does not happen. Nothing accumulates. Income and surplus are printed in the stats, never in the legends; the legends only say what went dark.

### Sources

A source is anything with a yield, a grant, or both. One record:

```
Source { Key, Name, Kind (world|belt|giant|star|structure|cosmic|elder|made),
         Star or (Feature, Radius), Mobile, Yield map[kind]float64,
         Needs []node (any one to harness), Grants []node, Levels (mil,sur,soc),
         Holder, Carried (expedition id or -1), Form, Consequence }
```

Worlds, belts and giants are read off the system the galaxy already generates. A source yields to whoever owns the star, or to a nomad fleet based there (grazing, below), and only if the holder knows one of its `Needs` nodes. Ranged sources yield to every people with a holding inside the radius.

**Natural sources, starting calibration.** Numbers are per tick and are the first thing to tune against the batch.

| Source | Yield | Needs |
|---|---|---|
| habitable world: lush 5, ocean 4, superterran 3, arid 2, twilight 2, lowg 2, iceshell 2, hothouse 1, floater 1, volcanic 1, dim 1 | O | agriculture (cradle: none) |
| terraformed world | +2 O | terraforming |
| rocky world (each) | 1 M × metallicity multiplier; superterran and volcanic +1 | metallurgy |
| world tagged heavy-element or uranium-rich | +1 E | atomic |
| belt (each) | 1 M | interplanetary; 2 M with a mine |
| gas or ice giant | 1 E | fusion with orbital habitats; floaters 1 O with high air |
| the star itself | collectors: 1 E (M-class) to 2 E (F and brighter) | orbital habitats |
| comets shaken loose (crowd law) | 1 O per system in a crowded field | interplanetary |
| supernova remnant, radius | rocky worlds inside yield double M | metallurgy |
| nebula, radius | 1 O per held star inside | synthetic biology + orbital habitats; also hides signal: no radio contact into or out of it |
| doomed giant, radius | 2 E per held star inside until it goes, then the blast | orbital habitats |
| globular cluster, radius | collectors yield double, rocky worlds half | as above |

**Structures as sources.** Every structure has an upkeep and a yield or a bonus. It is built only when the spare income covers twice its upkeep this tick, and it is dormant when it is shed.

| Structure | Node | Upkeep | Gives |
|---|---|---|---|
| mine (new) | orbital habitats | 1 E | belt yields 2 M; at most one per belt |
| collectors (new) | orbital habitats | 1 M | star yields as above; one per star |
| arcology | closed ecologies | 1 E | Sur +1, and the world's O is not lost to a failing sun or a hard sky |
| shipyard | orbital habitats | 1 M 1 E | Mil +0.5; fleets launched from this star reserve half M |
| defence grid | planetary defence | 1 M 1 E | Mil +1.5, as now |
| ansible net | the Voice | 2 E | Soc +1.5, as now |
| Dyson swarm | Dyson swarms | 2 M | star yields 6 E (K 4, M 3, F and brighter 8); replaces the collectors there; the flat research ×1.5 goes |
| accretion tap (new) | stellar engineering | 2 M 1 E | a black hole or neutron star yields 6 E |
| lifter (new) | star lifting | 3 M | a star or neutron star yields 3 M |
| vacuum tap | vacuum energy (node, no structure) | 1 M | +2 E per held star |

`build` stops picking at random. A people builds what fixes its deepest deficit, or the largest source in reach when there is none, one attempt per few hundred thousand years as now.

### Uses and upkeep

Every node from era 1 up has an upkeep, by domain and era, with named exceptions:

| Era | Biology | Energy, Computation, Exotic, Society | Industry, Weapons, Propulsion |
|---|---|---|---|
| 0 | none | none | none |
| 1 | 1 O | 1 E (electricity and steam: none) | 1 M |
| 2 | 1 O | 1 E | 1 M, 1 E |
| 3 | 1 O, 1 E | 2 E | 2 M, 1 E |
| 4 | 2 O, 1 E | 3 E | 3 M, 2 E |
| miracle | 3 of its domain's kind | | |

Exceptions: agriculture, atomic, fusion, dyson, vacuum energy and the source nodes cost nothing, they produce; kind nodes (the Gathering, Host-craft, Maintenance, Husbandry) cost nothing; world nodes (the Breach, High Air, Cold Chemistry) cost 1 of their domain's kind.

Other uses:

| Use | Reserves while it exists |
|---|---|
| fleet, relief, scout, surveyor: per level | 1 M, 1 E (evolver with living ships: 1 O, 1 E) |
| colony ship in flight | 2 M, 2 E (swarm seed-clouds: 1 O) |
| structure | its upkeep |
| research | none; the mind runs on spare energy: research rate × (1 + 0.1 × spare E), capped at 1.5, in place of the swarm's flat multiplier |

A reservation is released when the thing ends, is lost or comes home. Nothing is refunded because nothing was spent. A launch needs the reservation covered from spare income this tick, or it does not go.

### Direction: the order

The allocation is one function, `direct`, which returns the order of five categories and fills them; within a category the lowest era is filled first, so the newest things go dark first. It is the shedding rule and the policy in one place, like the scouting policy.

| Category | Contains |
|---|---|
| the fields | agriculture, medicine, ecologies, life extension, arcologies: the nodes that keep people alive |
| arms | weapons nodes, defence grids, fleets in being and in flight |
| the works | industry and energy nodes, mines, collectors, yards, swarms |
| the mind | computation, society and exotic nodes, the ansible net |
| the road | propulsion nodes, colony ships in flight, surveyors |

Default order: fields, works, mind, road, arms. Changes, in this precedence:
- at war, or fear above 0.6 with a hostile neighbour in reach: arms first;
- a fleet in flight is never shed for anything but the fields;
- hunger above 0.6: mind before works;
- greed above 0.6: road before mind;
- a planetary mind and a machine-born people have no fields;
- a nomad has no works and the road first.

The state of a node is `Known` and either working or dormant this tick. `recompute` sums only working nodes. A node dormant for a whole dark age is forgotten with the rest; dimming alone never forgets. The envelope shrinks only when a node has been dormant for ten ticks, so a bad millennium does not empty worlds.

When a people first sheds something in a stretch, one line: "The X let the yards at Y go dark to keep the fleet fed." When a shed node comes back: nothing. A stretch of shedding longer than a hundred thousand years is a fact, `FWant` (Woe 1): "the lean years".

### Rarities

A rarity is a source with `Grants`, `Levels` or both, and access is binary: a people **has** it or not. Having it means: owning or being based at its star, having a holding inside its radius, carrying it with a fleet, or a partner sharing it. One instance is enough; a second is worth nothing to the same people, which is what makes them worth trading.

**Natural rarities** are placed at galaxy generation from what is there. A field always has belts and giants, so nobody starves for want of anything, and the rarities are the extras:

| Object | Rarity | Effect |
|---|---|---|
| black hole | a horizon | grants causal physics and deep time at half cost; a fleet based there is lost 5% per tick |
| magnetar (radius) | the beam | grants the Unmaking at half cost; Mil +0.5; flares: dread for the fearful |
| pulsar | a beacon | reach +3 for whoever holds any star within its radius |
| white dwarf | the diamond star | Soc +0.5; a luxury the tellings name |
| neutron star | the heavy star | with star lifting, 3 M; grants stellar weapons at half cost |
| nebula | the colours | Soc +0.3; hides signal, as above |
| supernova remnant | the ash | grants exotic matter at half cost |
| Sagittarius A* | the Heart | grants transcendence at half cost; one of a kind; a people holding a star within its radius has glare to match |
| a companion note (white dwarf companion, brown dwarf) | a small rarity by kind | Soc +0.3 or 1 E |
| a world tag (rings, a moon of its own, a dust cloud) | a luxury | Soc +0.3 |

**Grants gate softly.** A granted node costs half when had. No node is *impossible* without a rarity, except where the design already makes it so (the Heart is unique, and that is the point). This is the catch-22 guard: harnessing a rarity needs a structure behind a node, and no node in that structure's prerequisite closure may depend on a rarity for anything but its cost. A test, `TestNoCatch22`, walks every source's `Needs` and fails if the closure reaches a node that is gated hard.

**Elder rarities** replace nothing and gate nothing; they are bonuses with a description, placed as elder legacies of a new kind, `Bounty`, and found like the rest (the Find is what it is now; a bounty is easy to use, since it was made to be used):

| Description | Gives |
|---|---|
| a lattice that turns starlight to metal, still turning | 3 M |
| a sea that has been growing since before the age | 3 O |
| a battery the size of a moon, a tenth full | 3 E, wearing to 0 over a million years; the one thing in this design that runs out |
| a seam of a metal that is not on the table | 2 M, Mil +0.5 |
| a garden that tends itself, under a roof of something clear | 2 O, Sur +0.5 |
| a mirror in orbit that never lost its polish | 2 E |
| a ring of black metal around a dead star, still warm | 4 E |

**Wielded artifacts** are mobile rarities with `Levels`. Nothing changes in how they are found; they become tradable in principle and never in practice, since a people trades a mobile rarity only when it has no use for it, and it always has.

**Mobility.** An immobile rarity is at a star or in a radius and changes hands only when the star does. A mobile rarity is carried: it sits at a star of the holder's, and a fleet that takes that star takes it and carries it home with the fleet. A fleet lost with a rarity aboard leaves it as a buried legacy at the star where the fleet died, for somebody to find. A nomad's mobile rarities ride with the greatest fleet.

### The Ember and the Manna

Two new miracle nodes in the style of the six: the Ember (Energy, era 4, from antimatter and exotic matter) and the Manna (Biology, era 4, from synthetic biology and germline). Discovering one **makes one object**, a mobile source. Losing the object does not lose the miracle; the people can make another after a million years. Both can be found as elder artifacts (they join the miracle share) and taken by force. A species can be born to the Manna (`born_manna`, power group, weight 10); nobody is born to the Ember. A people born to it rolls `kinfed` as usual; if it is not kinfed, the Manna is another organism living on the cradle in symbiosis with them from the start, its form and sentience rolled like any other, so the first line of a people can be that it keeps a mould that thinks.

Yields are large but finite, a swarm's worth: the Ember 8 E, the Manna 6 O. Each instance rolls a **form** and carries the form's **consequence**, since a departure from physics has a literal price:

| The Ember's form | Consequence |
|---|---|
| a pocket star | the safe one: dims by 2 E each time it is moved |
| a captive singularity | if its star is taken by force, 30% it gets loose: a blast on the star |
| a tap into a hotter place | glare on the holder's worlds; 1% per tick something comes through, an incursion at that star |
| a hole in the vacuum | +1 E per hundred thousand years; at 16 E it is a Doom for the star, and the tellings say who opened it |

| The Manna's form | What it is |
|---|---|
| a mould | grows on anything; can be given as a cutting |
| a mat of flesh | the floor of every city |
| a brood | insects, or something like them |
| a kelp | the oceans are full of it |
| a fungus that eats stone | 1 M as well, from what it eats |
| the people themselves | a caste bred for the table; see below |

Any form but the last is **sentient** with chance 0.2. A sentient Manna is a fact (`FManna`) that each people judges by its own morality, a crime to some ("they ate what thought") and nothing to a hive; see [morality](./morality.md). Each tick it may rise (0.001): it becomes a people at that star, of a new made trait, "grown for the table", and the eaters' O falls by the Manna's yield at once. Any Manna may get loose (0.0005 per tick, doubled on the holder's dark age): a horror of a new kind, the Manna Loose, a biological replicator that takes worlds and yields O to nobody. The Manna copies: giving a cutting to a partner gives the partner an instance and the giver keeps its own, so the Manna spreads by gift along the trade graph, which is the point of it and the danger of it.

**A species as its own Manna.** A species trait in the bio group, `autotroph` is the wrong word, call it `kinfed`: "who eat their own; a caste is bred for the table" (weight 4, standard and evolver kinds only, needs a caste or hive organisation). It is a Manna of the last form from birth: +3 O at every world, Soc −0.5, and a fact every other people judges by its own morality once known, with slant on top, so friends excuse it, enemies never do, and a hive does not see the question. It cannot get loose and cannot rise.

### Trade

Trade stays the relation it is, but it now moves surplus. Each tick, for each kind, in `sortedInts` order:

1. Each people computes its surplus S and its want W (the upkeep of its shed uses).
2. For each partner pair, the transfer of a kind from A to B is `min(S_A × cap, W_B × share_B)`, where `share_B` is B's want as a fraction of the wants of all A's partners, and `cap` is the transport fraction: 0.25 base, 0.5 with beamed sails or near-light, 1.0 with the Door or wormholes, 0 if no holding of A is within A's reach of a holding of B. The Voice does nothing to it; goods are not words.
3. B refills its shed uses in its order with what arrives. What A sent was surplus, so A loses nothing, and the sum of working uses rises. That is the positive sum, with no formula.

**Willingness.** A sends nothing to B if A holds B a monster or a grudge above 0.3, and sends half to a people its xenophobia counts as different. A fearful people (fear above 0.6) sends no M to a partner its intel reads as stronger and hostile in posture. A refusal that lasts is an **embargo**, a fact (`FEmbargo`, Crime 1 from B's side), and a cause for war the appraisal can use: a wanting people covets the partner's sources.

**Rarities in trade.** Partners share access to each other's immobile rarities, grants and levels only, never yield; yield is a commodity and goes through the surplus. Mobile rarities are not traded, except the Manna as a cutting, which a partner gives to a partner in want of O with chance 0.1 per tick, and never takes back.

**Dependence.** If B's shed uses this tick are covered by what a partner sends, B is dependent on that partner. When the trade breaks, by war, a pact left, the partner's fall or an embargo, B's uses go dark at once and a fact is written: `FCutOff` (Woe 2, Object the partner): "the X went dark when the Y stopped sending". The war appraisal counts what A would lose in trade against what it would take, so a war on a partner has a felt cost and the war count has something to move with.

**Nomads** graze: a fleet based at an unowned star takes half its natural yield, at a partner's a quarter, at anyone else's nothing, and stripping takes the rest once as now. They carry as before, and a nomad partner's cap is 0.5 regardless of drive.

### Force

Taking a star takes its sources: the yields, the immobile rarities, the structures (which come back to work if the taker knows the node, as now). Mobile rarities are carried off by the fleet. The appraisal gains a term: the yield and rarities of the front's stars, weighted by the attacker's want, so a rich neighbour is worth more to a wanting people and a hungry conqueror aims at the belts. Losing a star removes its yield the same tick; the direction function sheds; the line is logged.

### Kinds

| Kind | Difference |
|---|---|
| standard | as written |
| swarm | nests yield half; the Gathering costs nothing; seed-clouds cost O |
| planetary mind | no fields: it is the world; its cradle yields double O to it |
| parasite | while riding, its hosts' sources are its own; free-living as standard; mind-riders on machines count E as capacity |
| machine-born | no fields, no O use at all: every O cost is paid in E instead; Maintenance costs nothing |
| evolver | living ships reserve O; biology upkeep halved, industry upkeep doubled |
| nomad | grazing; no structures; road first |

### What it writes

Log lines: first harnessing of a source kind ("The X put mines in the belt at Y", "The X ring the black hole at Y with a tap and draw on what falls in"); shedding, once per stretch; a cut-off; the making and each consequence of the Ember and the Manna; a Manna rising or getting loose; a rarity named when first had, and in the gazetteer. Facts: `FHarness` (Deed 1), `FWant` (Woe 1), `FCutOff` (Woe 2), `FEmbargo` (Crime 1), `FManna` (Crime 2), `FRise` for a Manna that thinks (Deed 3 for it, Crime 2 for the eaters). Tellings: a blamed `FCutOff` reads "the Y starved us"; the scapegoat sweep picks it up like any woe.

`techstats` gains a section, **Means**: income and upkeep by kind at zenith, share of ticks shedding, what is shed most, trade moved by kind, dependence at the moment of a fall, rarities had by kind, gated nodes reached with and without the grant, sources harnessed by kind and by node, the Ember and Manna forms and what came of them.

### Per-tick order

```mermaid
flowchart LR
  A[sources: income by kind] --> B[direct: order by temperament and state]
  B --> C[fill uses; mark dormant]
  C --> D[recompute levels from working nodes]
  D --> E[trade: surplus to partners' wants]
  E --> F[refill; log what went dark]
  F --> G[build, expand, launch: need spare]
```

## Implementation notes

Five stages, each verified by the three-run determinism check and the ten-world batch, each committed on its own.

1. **Income, upkeep, direction.** `Source` type, natural sources read from systems and features, node upkeep table, `direct`, dormant state in `recompute`, log lines, `FWant`, the Means section in `techstats`. No trade, no rarities, no new structures. Calibrate here: a cradle should run era 2 in comfort, strain at era 3, and need colonies, a swarm or trade for era 4. This stage alone changes the batch most and is where the numbers get set.
2. **Reservations and structures.** Fleets, scouts, surveyors and colony ships reserve; `build` chooses by deficit; mines, collectors, taps and lifters; the swarm as a source; research on spare E; grazing.
3. **Rarities.** Natural and elder placement, access, grants at half cost, `TestNoCatch22`, mobility, force, wielded artifacts as rarities, the appraisal term.
4. **Trade.** Surplus sharing, caps, willingness, embargo, dependence, `FCutOff`, the war cost.
5. **The Ember and the Manna.** Nodes, objects, forms, consequences, `born_manna`, `kinfed`, the Manna Loose, rising.

Absorb into `DESIGN_NOTES.md` as a new v1 section, "Means", and a decision-log line; the *As built* figures from the batch go with it.

## Open questions

- The numbers. Every yield and upkeep above is a first guess; stage 1 exists to replace them. Is the cradle-alone target right: era 2 comfortable, era 3 strained, era 4 out of reach?
- Should natural sources at a star a people has only *read* yield anything? Proposed: no, only owned or grazed.
- Does the Exotic law stay as a research multiplier once the horizon and the ash exist as rarities, or does it fold into them? Proposed: stays, as ambient.
- Capacity: the number falls out of stage 1. Does it get its own proposal, or a stage 6 here?
- Soft gating at half cost, or a hard gate for one or two nodes with a mobile fallback? Proposed: soft everywhere; the Heart is the only unique thing and it gates nothing hard.
- Where the direction order lives when the player takes a people: it is the obvious first lever to hand over.
- Should the Manna's rising use the uplift path or a new one? The made trait "grown for the table" is new either way.
- Judgment of the Manna and of `kinfed` depends on the [morality](./morality.md) proposal; without it, both are a plain crime of weight 2.
- This repo has no `CLAUDE.md` and its canonical spec is `DESIGN_NOTES.md`. The workflow wants `CLAUDE.md` to link the spec; adding one is a one-line change to make when this is accepted.
