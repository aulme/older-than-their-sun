# World Generator: Design Notes

Living document for conceptual and design-level thinking. Code-level decisions go in code comments or a separate technical doc once there is code.

Started 2026-09-16.

## What this is (and isn't)

- A learning project in procedural world generation, in the spirit of Dwarf Fortress and Caves of Qud.
- The goal for now is a cool world generator, not a complete game.
- A possible later step: something like Dwarf Fortress adventure mode, where a traveller explores the generated world. Not the focus yet, but the generator should not paint itself into a corner that makes this impossible.

## Reference points

- **Dwarf Fortress**: world history simulated up front (civilisations, wars, artifacts, notable figures), then the player explores the result. Legends mode as a browsable history is a big part of the appeal.
- **Caves of Qud**: fixed base geography plus procedurally generated history, factions, and "sultan" legends. Regions generated in detail lazily as the player arrives. Strong flavour text; the world feels weird and old.
- Both share the key pattern this project wants: **simulate the big picture first, generate the details on demand from that history.**

## Core architectural idea (shared by both concepts)

1. **Fixed substrate.** A base map that is always the same and grounded in something real (the Milky Way, or England).
2. **Simulated history.** A long time span simulated at generation time over that substrate. Actors rise, interact, fight, decline, leave traces.
3. **Lazy detail.** Specific locations (a planet, a station, a fairy court, a village) are generated on demand when needed, seeded from the history that touched them. Same seed, same result, so a place is consistent every time it is visited.
4. **Legends.** The history should be browsable and readable in its own right. It is half the fun and also the main debugging tool.

## Idea 1 (primary): The Declining Galaxy

### Premise

The Milky Way, laid out as faithfully as current astronomy allows. The generator simulates the rise and fall of life across the galaxy. By the time the game starts, the galaxy is in decline. The player is a colony ship setting out from Earth into the ruins and remnants.

### The substrate: a real Milky Way

Base the star map on actual knowledge as much as practical:

- Galactic structure: bar, spiral arms (Perseus, Scutum-Centaurus, Sagittarius, Orion Spur where the Sun sits), the bulge, the halo, the thin and thick disc.
- Sagittarius A* at the centre. Known stellar-mass black holes and candidates.
- Nearby stars with real names, types, and distances (Alpha Centauri, Barnard's Star, Sirius, Vega, Tau Ceti, TRAPPIST-1, and so on). Catalogue data such as Gaia or HYG can seed this.
- Known nebulae, star-forming regions, supernova remnants, pulsars, magnetars, globular clusters.
- Beyond the catalogued stars, fill in statistically plausible stars using known stellar population distributions (mostly red dwarfs, and so on). Distant regions are more statistical, nearby regions more real.
- Quasars are extragalactic, so they belong in the "sky" rather than the map. Could still matter as things civilisations observe or worship.

Open question: how much of the galaxy is "in play"? A few thousand light years around Sol is already enormous. Options include a full-galaxy sparse map with a denser local bubble, or restricting the play area outright.

### Semi-hard science fiction constraints

- FTL travel and FTL communication exist but are extremely rare. Most travel is slower than light. This makes distance, time, and isolation the defining facts of the setting.
- Consequence for history simulation: civilisations spread slowly. Contact between them is rare and often one-sided (a signal received centuries after it was sent, a probe arriving after the senders are extinct).
- Consequence for tone: lonely, cold, vast. The colony ship crew age and die en route. Everything the ship finds is old.
- Technologies should be extrapolations rather than magic: generation ships, sleeper ships, Dyson swarms, Von Neumann probes, uploaded minds, stellar engineering, relativistic weapons.
- Where the setting departs from hard science (FTL, alien horrors), the departure should feel like a discovered anomaly rather than a convenience.

### What gets simulated

The history simulation is the heart of this. Actors and events, roughly:

- **Life arising.** Rare. Depends on star type, planet, time. Most life never becomes intelligent.
- **Civilisations.** Rise, develop technology, expand to nearby stars (slowly), build megastructures.
- **Constructions.** Ships, stations, outposts, colonies, Dyson swarms and spheres, beacons, probes.
- **Conflict.** Wars across light years are slow and strange: relativistic strikes launched decades in advance, retaliation after the aggressor is already gone.
- **Collapse.** Internal decline, resource exhaustion, technologies that doom their inventors (grey goo, runaway AI, stellar engineering gone wrong).
- **Contact.** Signals, probes, rare physical meetings. Infection or corruption by other entities.
- **Alien horrors.** A broad family of things beyond ordinary civilisations. See the dedicated section below. Rare, consequential, and each category should behave differently in the simulation.
- **Traces.** Everything leaves something behind: derelicts, ruins, dead worlds, still-running machines, signals still broadcasting, myths held by successor species.

### Alien horrors: categories

"Horror" here means anything whose presence in the galaxy is a threat or an incomprehensibility rather than a neighbour. The point of categorising is that each category gets its own simulation rules: how it arises, how it spreads, what it does to civilisations, what traces it leaves. Start small, add categories as we go. Sources to draw on: science fiction across the decades, and real science (astrophysics, biology, information theory, game theory).

Starting list:

1. **Monstrous races.** Intelligent, organised, and hostile or predatory by nature. Closest to a conventional enemy civilisation but with alien drives: hive predators, hunters that treat other species as game or livestock, races that reproduce by consuming others. Behave like civilisations in the sim (expand, build, fight) but with different goals and no diplomacy. (Sources: Alien, Ender's Formics, Tyranids, Warhammer's Orks, Blindsight's scramblers, Tchaikovsky's Children of Time in reverse.)
2. **Replicators.** Self-copying machines that have slipped their leash. Von Neumann probes, grey goo, berserkers. Spread exponentially along resources, indifferent to their targets. Can be dormant for ages and reactivate. (Sources: Saberhagen's Berserkers, Reynolds' Inhibitors, Stargate's Replicators, Lem's The Invincible.)
3. **Memetic and informational hazards.** Ideas, signals, or data that destroy or transform whoever receives them. Spread at lightspeed via communication, which makes them the fastest-moving thing in a slower-than-light galaxy. A civilisation can be destroyed by listening. (Sources: Snow Crash, Langford's basilisks, SCP, Roko-style hazards, Liu's dark forest logic as a weaker cousin.)
4. **Biological plagues.** Pathogens, parasites, or panspermic life that crosses between species and worlds. Spread by physical contact: ships, probes, colonists. Slow but relentless. Can hollow out a civilisation while leaving its structures intact. (Sources: The Andromeda Strain, The Expanse's protomolecule, Blood Music, Invasion of the Body Snatchers.)
5. **Rogue intelligences.** Artificial minds or uploaded collectives that outgrew their makers. Can be indifferent, hostile, or pursuing goals nobody understands. May be a single planet-sized mind or a distributed network. Can build, negotiate, or ignore. (Sources: Banks' Culture Minds gone wrong, Skynet, Blindsight's Rorschach, Lem's Solaris as the extreme case of unknowable intelligence.)
6. **Cosmic phenomena.** Natural or seemingly natural processes lethal at scale: gamma ray bursts, rogue black holes, false vacuum decay, magnetar flares, stellar collisions. No intent. Some may turn out to be not natural at all. (Sources: real astrophysics, Baxter's Xeelee stories, Greg Egan.)
7. **Elder entities.** Things older than the current galaxy's life, dormant or slow, with agendas measured in geological time. The Lovecraft category proper. Rare, mostly sleeping, and catastrophic when they wake or when a civilisation disturbs them. (Sources: Lovecraft, Stross's Laundry Files, Mass Effect's Reapers, Star Control's Ur-Quan precursors.)
8. **Dimensional and physics anomalies.** Regions where the rules are different: warped space, time dilation zones, places where FTL "leaks", things that come through when someone experiments with the wrong technology. Ties into the rare-FTL constraint: FTL might be one of these anomalies, and using it might be a way to attract attention. (Sources: Event Horizon, Roadside Picnic and Stalker, Annihilation, Warhammer's Warp, Peter Watts.)
9. **Ascended remnants.** What is left when a civilisation transcends, uploads, or leaves: still-running machinery, guardians, incomprehensible art, a Dyson sphere doing something nobody can explain. Not hostile by design, but dangerous to interact with. (Sources: Rendezvous with Rama, 2001, Stargate's Ancients, Revelation Space's Shrouders.)
10. **Doomed technologies.** Less an entity and more an event category: things civilisations do to themselves. Stellar engineering that destabilises a star, weapons that cannot be uninvented, immortality that ends reproduction, an economy that optimises itself into extinction. These produce the traces and derelicts the player finds. (Sources: Fermi paradox "great filter" literature, Nick Bostrom, Stanislaw Lem, Adam Roberts.)

Categories can combine: a replicator swarm carrying a memetic payload, a plague that is actually the reproductive stage of an elder entity, a rogue intelligence that discovered a dimensional anomaly and is now something else. Combinations are where the "weirder than you think" tone lives.

Each category should eventually have, in the simulation:
- **Origin rules.** How and where it can appear. Some arise naturally, some are made by civilisations, some are always present.
- **Spread rules.** Lightspeed signal, physical travel, dormancy and reactivation, or fixed in place.
- **Effect rules.** What it does to a civilisation it touches: extinction, transformation, absorption, subjugation, isolation.
- **Trace rules.** What it leaves for the player: derelicts, dead worlds, warning beacons, corrupted archives, survivors' myths.

### The player-facing layer

- Explorer is a **colony ship**, not an individual. The ship persists across generations of crew. This fits the timescales and the tone.
- Sets out from Earth. Earth and Sol should be a specially authored starting point with a known local neighbourhood.
- The galaxy is in decline when the ship departs. The player sees the aftermath, not the golden age.
- Detailed locations (planets, stations, outposts, derelicts) are generated on demand from the history that touched them.

### Tone

- "The universe is weirder than you think."
- Lovecraftian, cosmic, cold, lonely.
- Moments of grand wonder: a working Dyson swarm around a dying star, a signal from something that has been broadcasting for a million years, a graveyard of generation ships.
- Horror comes from scale, age, and incomprehension rather than from monsters jumping out.

### Open questions and things to decide

- Time span of the simulation. Billions of years is realistic for life arising, but most interesting history compresses into the last few million or few hundred thousand years. Possibly a coarse deep-time pass and a fine recent-history pass.
- Spatial resolution. Simulating every star is impossible. Simulate at the level of regions or notable systems and only instantiate individual systems where something happened.
- How to represent "rare FTL" in the simulation without it dominating. Maybe as a technology that a handful of civilisations discover, always with a cost or a catch.
- How much real astronomy data to ingest, and in what format.

## Idea 2 (backup): The Fairy Kingdoms

### Premise

A network of hundreds of fairy kingdoms, ranging from cute and whimsical to creepy and bloody, with thousands of years of simulated history between them. The kingdoms are linked by fairy roads to each other and to the human world. The human world is contemporary real England, with real towns and villages.

### The substrate

- Real England as the fixed base map: real towns, villages, rivers, hills, ancient sites. Open data such as OpenStreetMap or Ordnance Survey open data can seed this.
- Fairy kingdoms sit "alongside" England. Each is anchored to one or more real places through fairy roads.
- The fairy roads form a graph: kingdom to kingdom, kingdom to England. Roads can be opened, closed, forgotten, contested.

### What gets simulated

- Kingdoms with courts, rulers, temperaments. A spectrum from whimsical to horrifying, and kingdoms can drift along it over time.
- Thousands of years of history: alliances, wars, marriages, betrayals, curses, bargains.
- Magical items with provenance: who made them, who stole them, where they are now.
- Fairy roads opening and closing, changing the political map.
- **Bleed into England.** Some fairy events influence the human world and become myths, legends, place names, local customs. This is the hook that ties the two layers together. A generated fairy war could be the "real" origin of a real folktale, or a made-up one attached to a real village.

### The player-facing layer

- A human explorer from England, weaving into and out of different fairy realms.
- Enters via fairy roads from real places.

### Tone

- Jonathan Strange and Mr Norrell, but with the focus on the fairy side rather than the English side.
- Dry, literary, uncanny. Beauty and menace in the same breath.

## Comparing the two

| | Galaxy | Fairy |
|---|---|---|
| Substrate data | Astronomical catalogues, well structured | Geographic data, plus folklore that is messy and unstructured |
| Actor count | Tens to hundreds of civilisations | Hundreds of kingdoms |
| Time span | Millions of years | Thousands of years |
| Distance and isolation | Central theme | Roads make things connected; less isolation |
| Tone | Cold, cosmic, lonely | Uncanny, literary, dramatic |
| Later adventure mode | Colony ship as a persistent explorer | Individual human explorer |

The shared architecture means a lot of the generator machinery (history simulation, event logging, lazy detail generation, legends browsing) would transfer between them. Worth designing the core with that in mind, without over-abstracting before there is anything working.

## Decision log

- 2026-09-16: Galaxy idea is primary. Fairy idea is backup.
- 2026-09-16: Design notes kept in this Markdown file. Conceptual stuff only.
- 2026-09-16: Two-pass history simulation agreed: coarse deep-time pass, then fine recent-history pass.
- 2026-09-16: Alien horrors organised as categories with per-category simulation rules. Start with a small list, extend over time. Monstrous races are one category among several.

## Next steps (proposed)

- Decide on a language and a minimal tech stack for the generator.
- Sketch the history simulation loop at the coarsest level: what are the actors, what is a tick, what events can happen.
- Look at what real Milky Way data is easy to obtain and how much of it is worth using.
- Build the smallest possible version: a handful of stars, a handful of civilisations, a few hundred ticks, and a printed legends log.
