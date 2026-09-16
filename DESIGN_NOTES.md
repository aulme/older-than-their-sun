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
- **Out There: Ω The Alliance**: tone reference for the galaxy idea. A lone ship, scarce fuel and oxygen, every jump a gamble, alien languages learned a word at a time, ancient mysteries found rather than explained. Small, cold, and wondrous. Not a history simulator, but the feel of being one fragile thing in a very old galaxy is exactly right.
- Both DF and Qud share the key pattern this project wants: **simulate the big picture first, generate the details on demand from that history.**

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

### Filters

A **filter** is any event that may or may not force a civilisation into decline. The name is from the Great Filter idea in Fermi paradox discussion: the steps between dead matter and a galaxy-spanning civilisation, any of which most candidates might fail. Here the term covers everything from a nuclear age to a memetic beacon.

**Every filter has three possible outcomes:**

1. **Overcome.** Passed cleanly. Sometimes leaves a boon: a civilisation that aligned its machine minds grows faster afterwards.
2. **Scarred.** Survived, but changed for good. The scar is a permanent trait that alters behaviour and the odds of later filters. Dune is the model: a civilisation living under a total prohibition on thinking machines, or an atomic taboo, or a quarantine creed that keeps it from ever fully trusting contact.
3. **Declined.** Forced into one or more decline scenarios. Not always terminal: a **dark age** sets a civilisation back and can be climbed out of, though a third one is fatal. The terminal scenarios are the end states: extinct, transformed, contracted. Which scenarios are possible depends on the filter: failing the thinking machines filter produces a rogue mind, failing self-replication produces a swarm, failing transcendence produces silent machinery.

**Filters by era** (starting list, extend as we go):

- *Early, single world:* the Atomic Age (weapons that can end the world), Overshoot (ecological and resource collapse), Plague.
- *Middle, interstellar:* Thinking Machines (AI emergence), The Distance (colonies diverge across light years and generations), The Long Silence (immortality and stagnation), Contact (another civilisation, handled by the war and peace rules), Plague again.
- *Late, megastructural:* Self-Replication (nanotech and von Neumann machines), Stellar Engineering, Transcendence, FTL (which can let something through).
- *Ambient, always:* The Weight of Ages. The old crisis roll, scaled by age, size, galactic hazard and prior scars. It is what makes even a civilisation that passes every named filter eventually decline.
- *External:* horrors, cosmic events, and war are also filters in this sense, but they are driven by their own actors rather than rolled by the civilisation.

**Odds** depend on temperament (curious civilisations handle technology filters better, zealous ones handle stagnation better and machines worse), on prior scars (a scarred civilisation is more brittle, so each scar nudges later filters toward decline), and on galactic hazard.

**Why this matters for the aftermath.** The remnants the player meets are defined by their scars as much as by their fall. An emperor on a single world whose people abandoned machine minds four million years ago is a different encounter from one whose people are under a quarantine creed. The filter record of each civilisation is part of its legend.

### The present is the aftermath

The game's "now" is defined as the aftermath of galactic history. This is a hard rule for the simulation, not a mood:

- **Every civilisation must reach an end state by the present.** No civilisation is still at its zenith when the player sets out.
- **Three end states:**
  - **Gone.** Extinct. Only traces remain: ruins, derelicts, dead worlds, signals still broadcasting from nobody.
  - **Transformed.** Became something weird. Uploaded into a rogue intelligence, ascended and left machinery behind, devolved into a monstrous race, became a cult of a memetic signal. The transformed thing may still be active in the galaxy, but it is no longer a civilisation in the ordinary sense. Often it becomes one of the alien horror categories.
  - **Contracted.** Still alive but shrunk to one planet or a handful. Signs of old grandeur: an "emperor" ruling a single world that once ruled hundreds, crumbling megastructures, ceremonial titles that no longer mean anything. These remnants are the living civilisations the player can actually meet.
- **Every end has a cause.** War, an encounter with one of the horror categories, a doomed technology, a cosmic event, internal decay, or a combination. The cause is part of the legend and determines what traces are left.
- **The simulation guarantees it.** Rising hazard over time (horrors accumulate, they do not go away) should push most civilisations to an end naturally. Any civilisation still active at the present is pushed through a final decline pass, the "Long Dusk", so the rule always holds. Better to tune the sim so this rarely triggers, but the guarantee stays.
- **Humanity is young and late.** Humans never saw the galaxy alive. Sol is special-cased: the simulation cannot sterilise it or let a horror consume it, because the player must exist. Other civilisations can still have visited Sol and left traces, which is a feature.

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

## Simulation v0 (what the code does today)

Written 2026-09-16. Go, in `cmd/worldgen` and `internal/`. Run with `go run ./cmd/worldgen -seed N`. A full history takes well under a second.

**Substrate.** A random disc of a few hundred stars around Sol with a realistic spectral class mix. Placeholder until real catalogue data goes in.

**Deep pass.** 3 billion years in 10 million year steps. Life arises by star class, complex life follows, gamma-ray bursts sterilise regions, precursor civilisations rise and vanish off-screen leaving vaults, Dyson remnants or beacons, and elder entities settle in and go dormant.

**Fine pass.** The last 5 million years in 1000 year steps. Civilisations arise from complex life with a random temperament (curious, insular, aggressive, zealous). They grow tech with diminishing returns, reach the stars at tech 1, send sublight colony ships at 0.01c, build Dyson swarms past tech 2, and very rarely discover FTL past tech 3, which sometimes lets something through. Contact happens by proximity and can start slow relativistic wars. Civilisations face filters (see the Filters section) as they cross thresholds: the Atomic Age, Overshoot, Thinking Machines, the Distance, the Long Silence, Self-Replication, Stellar Engineering, Transcendence, Plague, and the ambient Weight of Ages. Each is overcome, leaves a scar, or forces a decline scenario. Scars change tech growth, expansion, war, and the odds of later filters.

**Horrors as actors.** Replicator swarms spread and eventually fall silent. Rogue minds spread slowly and absorb. Beacons convert or kill listeners within range, and converted civilisations start broadcasting themselves. Elder entities wake when settled too close, unmake everything within 40 light years, and sleep again. Hazard rises with the number of horror-held systems and beacons, which is what drives the aftermath.

**End states.** Extinct, transformed, or contracted (a remnant with a ruler title on one world, tech decaying toward a floor, which can later fade or be destroyed). Whatever is still active at year 0 is pushed through the Long Dusk.

**Output.** A chronological legends log plus a present-day summary: remnants, horrors, trace counts, and a fate line per civilisation.

### Observations from the first runs

- Around 15 to 35 civilisations per run. Roughly half extinct, a fifth transformed, a third contracted.
- The Long Dusk still fires for 2 to 4 civilisations per seed. Should be tuned down with rising hazard, not removed.
- Successor species arising on a dead homeworld happened naturally and is good. Keep accidents like this.
- Deep time log is dominated by "life arises" lines. Should be summarised into eras rather than listed.
- Civilisation lifetimes are around half a million to a million years. Expansion is fast relative to that. May want slower ships or slower tech.
- Sol is protected but nothing else is special about it yet. Nobody has visited.

Observations after adding filters (10 seeds):

- The early filters do most of the killing. Roughly a third of civilisations fall to the Atomic Age or Overshoot, and about a third of those falls are fatal. That matches the Great Filter idea: most intelligences never leave their world.
- The Long Silence produces most remnants. "Stopped dying, and then stopped being born" is the most common way a civilisation ends up contracted. Worth diversifying the contraction causes.
- Around 60 percent of civilisations end extinct, 10 percent transformed, 25 to 30 percent contracted. Five to fifteen dark ages per run.
- Scars stack: remnants living under two or three creeds are common, which is exactly the Dune flavour wanted.
- Filters arrive fast, within a few hundred thousand years of a civilisation arising. Plausible given thousand-year ticks, but a civilisation's whole story can be over in 300 ticks. Slowing tech growth would spread it out.
- The Weight of Ages rarely gets to act because the named filters catch most civilisations first. It matters mostly for the survivors of everything else.

## Simulation v1 (proposed model, not yet built)

Discussed 2026-09-16. This replaces the scalar "tech" and the hand-weighted filter odds of v0 with a small set of interlocking systems. Nothing here is coded yet.

### Species and home worlds

Every civilisation starts as a **species on a home world**, and the world explains the species.

- **Star(s).** Class from the substrate. Multiplicity rolled (about half of sunlike stars are in multiples): binaries give cyclic climates and hardy life, trinaries are rare and strange. Star lifetime matters: an F star dies within the simulation window, so a species born there faces a "dying sun" filter eventually. Massive stars nearby are future supernovae.
- **World archetype**, weighted so lush worlds are most common and extreme ones rare but interesting. Starting list: temperate lush; ocean; arid; ice or tidally locked twilight band; high-gravity superterran; low-gravity; hothouse under a thick atmosphere; subsurface ocean under an ice shell; gas giant aerial; tidally heated volcanic; dim world around a brown dwarf. Each archetype sets **physical facts** (gravity, heat, atmosphere, orbit, major biomes) and **base traits**: high gravity breeds robust bodies and strong survival; an ocean world delays flight and metallurgy but breeds cooperation; an ice-shell world has never seen a sky, so astronomy and the urge to leave come late; gas giant floaters cannot make fire, so the industrial path is locked until an exotic route opens.
- **Random traits** on top of the base ones, from a pool in rarity tiers. Keep three to five traits per species so each one reads in the legends. Common tier is social organisation and drives (individualist, collective, hive, caste, non-conscious intelligence, expansionist, contemplative, xenophobic, submissive, fight-to-the-death, pacifist). Uncommon tier is biology quirks (short-lived, very long-lived, cyclical dormancy, many sexes, sessile adults). Rare tier is science-fiction powers: ansible minds that communicate faster than light, directed evolution, precognition, memetic immunity, machine symbiosis, unbroken memory across generations.
- **Traits are asymmetric against filters.** A hive mind cannot schism, so the Distance and the Weight of Ages barely touch it, but it is one mind, so a beacon that catches it catches everything. Unbroken memory makes the Long Silence harsher. This asymmetry, not raw bonuses, is what makes species feel different.
- **Kinds.** Beyond traits, a species has a body plan that sets the flavour of everything it builds: standard; swarm; planetary mind (a living ocean, a forest that thinks); parasite or memetic virus that needs host species; machine-born; evolver that directs its own flesh. The tech tree is the same for all; the words are different. A standard species builds a space station, an evolver breeds a placid sub-species that survives vacuum and hosts a city, a swarm secretes one. A few kinds get hard rules: a parasite spreads through other civilisations rather than colonising, a planetary mind has very low reach and very high social. Keep the hard rules few and the flavour tables large.
- **Species can be made.** Some filter outcomes produce a new species: an uplift by a stronger neighbour, a directed-evolution schism, slaves bred into something else, a transformation that leaves a successor. Made species carry a trait recording who made them.

### Three levels and reach

Each civilisation tracks three levels, on a small named scale so legends can say "a warlike people" rather than print a number.

- **Military.** Answering threats with force and tech: wars, monsters, machines. Mostly from tech and structures, a little from traits.
- **Survival.** Continuing to live in degraded environments: plague, ecological collapse, a dying star, a scoured colony. Mostly from traits and biology tech. Survival also widens the **habitable envelope**, the set of worlds the species can colonise at all.
- **Social.** Holding together: unity under stress, ability to act as one when it matters. Mostly from traits, a little from tech and structures, and it is dynamic: wars, plagues, dark ages and distance erode it, peace restores it. Scars push it (iron centralism raises it and freezes it).

Rare science-fiction traits can move all three.

**Reach** is how far from home a civilisation can act, in light years, and it is separate from the levels. Structures are only built and worlds only colonised within reach. Two civilisations **encounter each other when their reach spheres overlap**. Reach comes from propulsion and communication tech and a few traits, and it steps: one world, one system, a few light years by slow ship, then more, then far if FTL exists. Ansible minds effectively multiply reach for social purposes because a colony a hundred light years away never diverges. Reach spheres are also what makes the later scale-up cheap: a civilisation is a sphere and stars only need to exist where a sphere touches them.

### Filters resolved as tests

A filter now **tests one or two levels against a difficulty** instead of rolling hand-tuned weights. Roll plus level minus difficulty gives a margin: a clear pass is overcome, a narrow one is scarred, a clear failure is declined, with the decline scenarios still specific to the filter. Difficulty scales with galactic hazard and prior scars, as before.

| Filter | Tests |
| --- | --- |
| Atomic Age | Social |
| Overshoot | Survival, Social |
| Plague | Survival |
| Thinking Machines | Social |
| The Distance | Social (reach makes it harder, ansible traits negate it) |
| The Long Silence | Social |
| Self-Replication | Military |
| Stellar Engineering | Survival |
| Transcendence | Social |
| War | Military, then Social to hold together |
| Replicator swarm, rogue mind | Military |
| Beacon | Social |
| Elder entity | Survival, and reach to flee |
| Cosmic event | Survival, and reach to evacuate |
| The Weight of Ages | Social |

**Filters attach to tech.** Discovering a node can proc its filter: the Atomic Age fires when atomic power is discovered, Thinking Machines when machine minds are, Self-Replication when self-replicating industry is. This replaces the tech-threshold triggers of v0. The ambient and external filters keep their own triggers.

**Filters steer research.** Facing a filter raises discovery weights in the domains that answer it for a while: a plague pushes biology, a war pushes weapons, a dying sun pushes propulsion. Scars can lock domains permanently (a prohibition on thinking machines closes computation).

**Cosmic filters** are the new external category: a supernova, the death of a home star, a passing black hole, a gamma-ray burst. Each has a blast radius and every civilisation inside it faces the filter at once. Overcome can mean **migration**, a new home world and the old one left as a ruin, which is a legend worth having. With a real star catalogue the supernova candidates and the dying stars are known in advance and can be scheduled rather than rolled.

### War, submission, enslavement

Contact happens when reach overlaps. What follows depends on **stance traits and relative military**:

- Peace or trade when neither is expansionist or xenophobic. Trade raises both sides' discovery rates.
- Submission without a war when one side is far stronger and the weaker has a submissive trait: it becomes a vassal.
- War otherwise, decided mostly by military with social deciding whether the loser holds together. Fight-to-the-death species are exterminated rather than enslaved. Pacifists do not fight and are ignored, absorbed or enslaved.
- **Enslavement** is an outcome alongside destruction. An enslaved civilisation keeps its species and home, loses its reach, researches slowly, and adds to its master's military. When the master faces a decline, the slaves face a **revolt filter**: overcome and they are free, often inheriting the master's ruins and becoming the remnant the player meets; decline and they fall with the master. Slaves can also be bred into a made species, which is a transformation.
- **Uplift** is the benign version: a strong civilisation makes a new species from complex life within its reach, with a client relationship that can later go the same way as vassalage.

### Tech tree

High level only: Industrial Revolution, not the internal combustion engine; Slow Interstellar Travel, not a particular drive. Around eight **domains**: energy, industry, computation, biology, society, propulsion and communication (reach), weapons, exotic physics. Start with 50 to 100 **nodes** with prerequisites; the tree can grow later, but richness should come from how nodes, traits, scars and filters combine, not from node count.

A node has: domain, prerequisites, effects on the three levels and on reach and envelope, changes to discovery weights in other domains, structures it unlocks, a chance to proc a filter on discovery, and flavour text per kind.

**Discovery** is random each tick. The domain is drawn from weights that are base times traits times current filter focus times scars, then an available node in that domain is taken. Discovery rate scales with size and social level. The old scalar tech becomes a derived era label.

**Structures** are built within reach, require nodes, and give levels: orbital defences give military, arcologies and closed ecologies give survival, an ansible net gives social across distance, a Dyson swarm multiplies energy and so discovery. Flavour by kind.

### Order of building

1. Species and home world generator with traits, levels and reach as data; legends print a species portrait.
2. Filter resolution as level tests.
3. Tech tree driving levels and reach; drop the tech scalar.
4. Contact, war, submission, enslavement, revolt, uplift.
5. Cosmic filters with blast radii. Pairs well with ingesting the real catalogue, since star lifetimes come with it.
6. Kinds as flavour tables, then the few hard kind rules.

## Decision log

- 2026-09-16: Galaxy idea is primary. Fairy idea is backup.
- 2026-09-16: Design notes kept in this Markdown file. Conceptual stuff only.
- 2026-09-16: Two-pass history simulation agreed: coarse deep-time pass, then fine recent-history pass.
- 2026-09-16: Alien horrors organised as categories with per-category simulation rules. Start with a small list, extend over time. Monstrous races are one category among several.
- 2026-09-16: Language is Go. Reasons: the history sim is a graph of mutually referencing entities, which a garbage collector handles cleanly; performance needs are modest; fast iteration matters more than raw speed. Rust's sum types would have been nicer for event modelling, accepted as a cost.
- 2026-09-16: "Filters" is the term for anything that may push a civilisation into decline. Three outcomes: overcome, scarred, declined. Early filters like nuclear weapons included. Scars are lasting traits.
- 2026-09-16: Filter balance is fine for now. The lethality of early filters is accepted; the fix later is scale, not tuning. With hundreds of billions of stars the sim can afford far more civilisations, so a high failure rate still leaves plenty of history. Not for now.
- 2026-09-16: Simulation v1 model sketched (species from home worlds, three levels plus reach, filters as level tests attached to tech nodes, war with enslavement, high-level tech tree, exotic kinds as flavour). See "Simulation v1". Not built yet.
- 2026-09-16: The present is the aftermath. Every civilisation ends as gone, transformed, or contracted, always with a cause. Sol is protected by fiat.

## Next steps (proposed)

- Read the v0 legends output and adjust rates until histories feel right: civilisation lifetimes, how many civs, how often each end state occurs.
- Look at what real Milky Way data is easy to obtain (Gaia, HYG) and replace the random star field with real nearby stars.
- Later: scale up massively. The real galaxy has hundreds of billions of stars, so civilisation counts should go from dozens to thousands or more. This needs a different simulation structure (regions, statistical treatment of the unremarkable, only instantiating stars where something happens). Design for it, but do not build it yet.
- Add more horror categories to the sim: plagues as spreading actors, monstrous races as civilisations with alien drives, dimensional anomalies tied to FTL use.
- Add the lazy detail layer: given a star and its history, generate the system contents (planets, stations, derelicts).
