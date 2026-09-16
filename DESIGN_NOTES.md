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

Decided 2026-09-16: the whole galaxy is the map, but a history runs in one *field* of a few hundred stars at a chosen place in it. See "The galaxy as a map" under Simulation v1.

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
- **The simulation produces it, not a guard.** (Revised 2026-09-16.) The cycle's fading fertility and the weight of the age push nearly every civilisation to an end on their own. The few still standing at the present (one to five per run) are reported as such: the last ones, in the waning of the age. The earlier "Long Dusk" pass that forced them into decline has been removed; if many stand at the present, tune the fading, do not bring the guard back.
- **Humanity is young and late.** Humans never saw the galaxy alive. Sol is special-cased: the simulation cannot sterilise it or let a horror consume it, because the player must exist. Other civilisations can still have visited Sol and left traces, which is a feature.

### Ages of the galaxy

Discussed 2026-09-16. The galaxy's history is a series of **ages**, each a burst of civilisations rising and falling, separated by **interregna** of hundreds of millions to billions of years in which only natural cosmic events happen. The fine simulation is the current age, and the present is its tail end. Everything before it is myth.

- **Each age ends by attrition, not by a blow.** (Revised 2026-09-16, superseding the earlier "age-ender" idea.) See "The cycle" under Simulation v1. The galaxy's fertility for new spacefaring species surges at the start of an age, then declines a little every tick until it is near zero. Cosmic events and horrors only mop up what the fading leaves. The Long Dusk guard is gone. The player arrives knowing, from the legends of prior ages, that ending is the rule.
- **Prior ages are played very coarse.** One tick per rise, one per fall. Each age produces a handful of **elder civilisations** with a vague portrait rather than traits: something that thought in the convection cells of a red giant, a mind spread through the magnetic field of a nebula, a species that lived in the dark between stars. They are deliberately weirder than the weirdest current alien, and they left no trace of themselves, only their works. First age no earlier than about 7 billion years ago, when the galaxy had enough heavy elements for worlds.
- **Names are given by the finders.** Elder civilisations have no surviving name. The current age calls them by what was found: the Ones Who Moved the Star, the Makers of the Hollow Sun. Two current civilisations may name the same legacy differently. Optional but cheap.
- **Legacies** are what an age leaves, placed on the real substrate. Kinds:
  - *Artifacts.* Mystery tech with unknown function. Finding one grants a node the finder never researched, with its filter attached, which is why they are dangerous.
  - *Structures.* Things that should not exist: a moved star, a black hole in a ring, a corridor of darkness, a hollowed sun. Dungeons for the player.
  - *Threats.* What ended an age, or the weapons of its wars, still around. Some of the current horrors originate here.
  - *Sleepers.* Elder civilisations that did not die but withdrew into slumber. More or less lovecraftian gods. This is now the backstory of the dormant elder entities the sim already has, and some wake in the current age.
  - *Laws.* Things an age left in the universe itself: a jump network, a standing signal, a region where minds do not work. Rare.
- **Legacies act on the current age through the Find**, a filter that fires when a civilisation's reach touches a legacy. Unlike other filters it has four outcomes, and the species chooses which one to attempt:
  - *Mastered.* Very hard. The finder understands the thing and gains the entire tech tree up to that node. Every tech filter attached to the nodes gained fires, so mastering a late artifact cashes in several filters in quick succession. A shortcut with the bill attached.
  - *Wielded.* Moderate. Only this artifact is usable, as a level bonus (an artifact weapon adds military, an artifact engine adds reach) with no node gained. They do not know how it works: if it is lost or broken it stays broken. The node's own filter still fires, because the thing is in use.
  - *Sealed.* Easier. The artifact is left dormant, and it stays on the substrate to be found again by a successor, by the same species after a dark age, or by the player. The only outcome that triggers nothing. Requires social cohesion to hold: someone always wants to poke it.
  - *Unleashed.* The failure. The artifact acts uncontrolled and forces a worse filter: the legacy's own threat, at raised difficulty. A sleeper wakes, a replicator wave starts, a law is broken locally.
  - Which outcome a species attempts depends on traits (curious and expansionist species reach for mastery, cautious and xenophobic ones seal, pragmatic ones wield). Whether it succeeds depends on levels, and failure slides down the list: a failed mastery becomes a wielding or an unleashing, a failed sealing is an unleashing.
- **The current age becomes the next myth.** The legacies of current civilisations (Dyson remnants, beacons, rogue minds, made species) are exactly what an elder age leaves, at finer grain. One legacy model at two granularities: the coarse age generator and the fine sim should write the same record type, so a later age generator could run on this age's output.
- **Interregna** are pure substrate time: star deaths, supernovae, a gamma-ray burst, a nebula born from an elder war, drift. The current deep pass (life arising, gamma-ray bursts, precursors, elders) folds into this: precursors become the last elder civilisations of the previous age, elders become sleepers.

**Three passes.**

1. *Ages.* Previous ages, extremely coarse: the age generator, one tick per rise and fall, output is legacies on the substrate and the myth log.
2. *The youth of the current age.* Medium grain. The v1 civilisation engine with large ticks (tens of thousands of years) and rates scaled to match, so civilisations rise, face filters, and end, but ships and wars are not followed in detail. Sets up the world the fine pass inherits: remnants, horrors, legacies of this age.
3. *The waning.* Fine grain, thousand-year ticks, the same engine. Everything the legends follow closely happens here. Where it ends is decided by the state of the galaxy, not by a date; see "The cycle".

The middle and fine passes should be one engine with a tick size, not two engines, so that a civilisation that spans the boundary carries over without translation. How long the current age is, and where the medium/fine boundary sits, is tuning.

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

**End states.** Extinct, transformed, or contracted (a remnant with a ruler title on one world, tech decaying toward a floor, which can later fade or be destroyed). Whatever is still active at year 0 was pushed through the Long Dusk (removed 2026-09-16).

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

## Simulation v1 (built 2026-09-16)

Discussed and built 2026-09-16. This replaces the scalar "tech" and the hand-weighted filter odds of v0 with a small set of interlocking systems. The sections below are the design; "What the first v1 runs showed" at the end of this section records how the build went and what the first runs look like.

### Species and home worlds

Every civilisation starts as a **species on a home world**, and the world explains the species.

- **Star(s).** Class from the substrate. Multiplicity rolled (about half of sunlike stars are in multiples): binaries give cyclic climates and hardy life, trinaries are rare and strange. Star lifetime matters: an F star dies within the simulation window, so a species born there faces a "dying sun" filter eventually. Massive stars nearby are future supernovae.
- **World archetype**, weighted so lush worlds are most common and extreme ones rare but interesting. Starting list: temperate lush; ocean; arid; ice or tidally locked twilight band; high-gravity superterran; low-gravity; hothouse under a thick atmosphere; subsurface ocean under an ice shell; gas giant aerial; tidally heated volcanic; dim world around a brown dwarf. Each archetype sets **physical facts** (gravity, heat, atmosphere, orbit, major biomes) and **base traits**: high gravity breeds robust bodies and strong survival; an ocean world delays flight and metallurgy but breeds cooperation; an ice-shell world has never seen a sky, so astronomy and the urge to leave come late; gas giant floaters cannot make fire, so the industrial path is locked until an exotic route opens.
- **Random traits** on top of the base ones, from a pool in rarity tiers. Keep three to five traits per species so each one reads in the legends. Common tier is social organisation and drives (individualist, collective, hive, caste, non-conscious intelligence, expansionist, contemplative, xenophobic, submissive, fight-to-the-death, pacifist). Uncommon tier is biology quirks (short-lived, very long-lived, cyclical dormancy, many sexes, sessile adults). Very rare tier (under one percent of species) is being **born to a miracle**: see "Miracles". Machine symbiosis and unbroken memory moved to the uncommon tier (2026-09-16).
- **Traits are asymmetric against filters.** A hive mind cannot schism, so the Distance and the Weight of Ages barely touch it, but it is one mind, so a beacon that catches it catches everything. Unbroken memory makes the Long Silence harsher. This asymmetry, not raw bonuses, is what makes species feel different.
- **Kinds.** Beyond traits, a species has a body plan: standard; swarm; planetary mind (a living ocean, a forest that thinks); parasite that needs host species; machine-born; evolver that directs its own flesh. Since 2026-09-16 the kinds are mechanically different, not only in vocabulary. A **swarm** spreads and does not hold: nests are cheap and twice as likely, each counts half for research, there is no centre for the Distance to pull on, the burning sky is easier and the Signal harder, and a swarm cannot be enslaved, only burned out. A **planetary mind** is the world: no Distance and no schism, a taken home is death, plague is harder, research a third faster, and reach a third until it learns to graft itself onto other worlds. A **parasite** grows by hosts: it starts riding a local host species, settles almost nothing on its own until it learns to live free, takes an infected people as slaves and half of what they knew, uplifts readily, and is as rich in Social as its hosts; an empty field weighs on it. **Machine-born** peoples never arise on their own: they are what is left when someone builds a mind that outgrows them (half of Thinking Machines falls). They live on rock and vacuum, do not sicken, endure a failing sun three times longer, ossify under the Weight, and have computers, machine minds and hibernation from birth. An **evolver** breeds what others build: biology at half price, industry dear, one step wider envelope, the Brood easier and Overshoot harder. Each kind has nodes of its own (the Gathering and Seed-clouds; Grafting and Deep Root; Host-craft, Broodline and Free-living; Maintenance and Forking; Husbandry of the Self and Living Ships).
- **Senses and the herd.** A species may have senses beyond or instead of the usual five (eyeless and seeing by sound, deaf, heat-sight, a feel for the world's magnetism, for living current, for radiation, for the chemistry of the air), and its social organisation runs from solitary through individualist, collective and herd to hive. These matter through the aptitude table below, not through raw bonuses.
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

A dying sun is a **slow filter**, not a single roll. Once the star begins to fail, the home world degrades every tick and the species keeps living there as long as its Survival holds out: a high-Survival species endures under a swelling or fading star for a very long time, a low-Survival one has a few ticks. Survival buys time; reach decides whether the time is enough to leave. A species that endures to the end without leaving is extinct with its star, and its world is a distinctive trace: a burned or frozen cradle around a white dwarf or a giant.

### War, submission, enslavement

Built 2026-09-16 as designed below, with the tuning noted at the end. Enslavement, vassalage, revolt, uplift and breeding stay as they were:

- **Enslavement** is an outcome alongside destruction. An enslaved civilisation keeps its species and home, loses its reach, researches slowly. **Dominion** (2026-09-16): the held feed the master's armies, works and confidence and lose more than the master gains: per slave the master has +0.4 Military, +0.2 Survival, +0.3 Social (four slaves' worth at most) and the slave −0.8, −0.4, −0.6; a vassal gives half and loses half. When the master faces a decline, the slaves face a **revolt filter**: overcome and they are free, often inheriting the master's ruins and becoming the remnant the player meets; decline and they fall with the master. Slaves can also be bred into a made species, which is a transformation.
- **Uplift** is the benign version: a strong civilisation makes a new species from complex life within its reach, with a client relationship that can later go the same way as vassalage.

**The war object.** A war holds its two sides, the year it began, its cause, a will to fight per side, a tally of worlds taken, glassed and converted, and a name once it has taken a world (after the most contested star). Peace writes a record into both peoples: who took what, the truce, the grudge.

**The front is the overlap.** Each tick the war recomputes, for each side, its worlds that lie within the other's reach measured from the other's nearest holding. Only those worlds can change hands. A home is in the front only when the enemy's reach covers it, and it always carries the home defence bonus. Peoples whose homes lie beyond each other's reach fight a border war over colonies, which is the standard war; the other kind is the expedition.

**Strikes take time.** A strike goes to the nearest front world and its rate per thousand years falls with the crossing time, distance times the attacker's ship speed. Slow arks at twenty light years give a strike every few thousand years, relativistic ships one or two a millennium, the Door none. A won strike is conquered (if the winner can live there and is not xenophobic), glassed, or converted (parasites and swarms, which never glass); the Unmaking unmakes, the Chorus takes worlds whole. The loser faces the strain of war on the first loss and every third after.

**Will decides when it ends.** Each side starts with a will from posture and cause. Will drains every tick, faster past fifty thousand years; a world taken raises the taker's will and lowers the loser's, unless the loser is a conqueror or unyielding, where it rises. Both wills gone, or one gone and the other with nothing left in the front to take, is peace. One gone and the other still able to take is capitulation: the front ceded, and if the home lies in enemy reach vassalage for the submissive or where the winner is uninclined, slavery otherwise. Extinction stays what it is: the home taken by a xenophobe of the different, or an unyielding people that would not surrender.

**Peace leaves a grudge** and a truce of a few thousand years. After it a hostile side with a grudge and a changed balance renews the war at a raised chance and the legend counts them: the second war between the two, the third. Causes beyond first contact: an expansion aimed inside the other's reach, a neighbour weakened by a dark age or plague or an army away, the church scar for a zealous people, revenge, extermination of the different.

**Postures** replace the stance trait group outright, so nobody is martial and pacifist at once:

| Posture | Declares on | Will | Peace and terms |
|---|---|---|---|
| Pacifist | nobody | only while the home is in the front | cedes the front at once |
| Defensive (the commonest) | nobody; answers border friction | low | quick, status quo |
| Submissive | nobody | lowest | capitulates early, vassalage over ruin |
| Opportunist | anyone in reach weaker by two levels or weakened | high while winning, folds on the first real loss | takes what it holds and stops |
| Conqueror | any neighbour the appraisal allows | highest, climbs with each world | wants homes: vassal or slave |
| Vengeful | nobody first; anyone who ever struck them | fed by the grudge, which never decays | renews at every truce's end |
| Confederate | a common threat, with allies | shared | binds allies to one peace |
| Unyielding | as defensive | never runs out | never sues; the home dies fighting |

**Xenophobia stays a drive** layered over any posture. A difference score between two species (another kind counts most, then order such as hive against individualist, then senses and body, then home world) decides whom a xenophobe treats as its posture says and whom as a thing to end: no trade, no pact, glass instead of conquer, scour instead of enslave, will that does not drain. Everyone else feels difference a little, as less trade and readier war. Parasites read as different to nearly everyone.

**Parasites fight by infection.** A parasite always opens on a host species in reach. The host's Infection filter is its opening response: overcome fights with full will, scarred fights while burning its own infected worlds, declined does not resist and the front falls world by world. A parasite strike converts a front world to a host world (a new host species on the first); a host strike on a converted world is a quarantine burn, and each burn drains the parasite's will. "They burn their own worlds to stop it" is a peace, earned rather than rolled. The home falling is enslavement with the tech gain.

**Expeditions** are the army far from home. Any number at once, each at least a tenth of the people's Military, which leaves home for the duration; the home level drops by that much and neighbours see it. An expedition crosses at ship speed, then fights battles (about one a millennium) from a base, the nearest enemy world, against the defender's Military plus home and defence bonuses. A win takes the world and moves the base; a loss costs it thirty percent of its strength; away from a world it can live on it thins by attrition. Below one level it is spent. A spent or stranded force, the more so the farther and longer away, goes native: it keeps the world it holds as a successor state with the parent's tech and a new name; otherwise the survivors come home with what is left. Taking the enemy's home ends the war on the winner's terms with the expedition as garrison. Allied expeditions at one target strike as one, levels summed, a loss shared in proportion. Conquerors do this by nature, the vengeful for a distant enemy, xenophobes to end the different; opportunists, defenders and pacifists never.

**Scouts** are expeditions of one level that do not thin and cannot be lost; the level is gone while the scout is out. A scout makes a round trip to a world and reports the strength there as it was: the home level if it is the home, otherwise the garrison and any stationed relief, plus the defence grid. With the Voice the report comes at arrival; with the Sight a people reads one world per tick for free and sees every expedition on its way to any of its worlds, allied, returning or hostile. Intelligence has a date and its noise widens with age (half a level per five thousand years, up to three). Trade partners see each other every tick; war reports strength as of the last strike, but not relief stationed nearby; a people that watched a young species from orbit keeps watching. **Scouting policy** starts simple: scout when the appraisal's spread straddles the posture's bar, so a report would change the decision, and the level can be spared. It lives in one function so a better policy can replace it later: proactive scouting by posture (opportunists forever looking for targets, defenders measuring their neighbours), since a scout takes thousands of years and a smart people does not wait for the question.

**Watching the sky.** Any people sees an expedition, but not a scout, that passes within its watch range of any of its holdings: fleets seen by telescope, a few light years out. Watch range comes from the tree like reach (astronomy a little, computers and orbital habitats more, a defence grid most; the Sight sees everything bound for you) and a relativistic fleet is seen from twice as far, since its drive is a torch. A fleet seen inbound is the last-minute call: the people can recall its own forces, tilt to defence, call allies (who will rarely arrive in time unless already stationed) or sue for terms before the first battle.

**Exploration.** Every star is seen: its place, its colour, whether it is dying. What a star holds is known only by reading it. Three states per star, per people: *seen* (always); *read* (worlds known: a surveyor's visit, a colony ship's arrival, a nomad fleet basing there, the Sight turned outward, or a telescope within watch range, which shows worlds but not who is on them); *known taken* (a visit, a reading, or an owner loud enough to hear from a holding). About half the stars of a friendly colour have no temperate world at all, and a people without a wide envelope cannot live there, and before reading nobody knows which. Colony ships go to read stars; when nothing read is free and livable, one ship in five goes on a guess to a star of the right colour, and may find nothing to live on, or somebody already there, which is a meeting. Settling a star finds whatever is buried there with certainty. **Surveyors** are the scouts of exploration: expeditions of one level, out for a tour of up to six stars or forty thousand years, the nearest unread star first and then the next within a hop, home when the tour ends or a war recalls them. A star held by a horror keeps half the ships that reach it. Hunger and Greed set how many a people keeps out, up to three; a level must stay home; wartime sends none; a people with nothing read to settle sends one out of necessity. Readings go stale after a million years and one ship keeps the charts current. **The Sight is a mode.** Turned outward it reads a couple of stars a millennium out to twice the reach, five surveyors' worth, and marks remains for the surveyors to visit; while it does, the free reading of rivals and the sight of fleets bound for one's worlds are off. A war, or for the fearful a hostile neighbour in reach, turns it back to the borders. The chance lottery survives at a fortieth of its old rate, plus a people's own lost works, which it knows where to look for. **Contact follows.** Territories that overlap meet in the flesh only once one side has heard the other or read one of its holdings; surveyors and colony ships that come upon a people meet it on the spot; a quiet young species stays unfound until somebody arrives.

*As built* (reference batch, ten worlds): 63% of starfaring peoples ever surveyed, 16 surveys each, about four readings per survey. Finds: 59% by surveyors, 13% by settling or a ship arriving, 4% a people's own works, 24% chance. Sight-holders spend 16% of their time outward. Meetings by survey or ship are a few percent of meetings; most flesh meetings follow hearing. Never-met fell from 60% to 51% of peoples. Median life rose from 1.2 to 2.0 Myr and unleashings fell by more than half, since remains are now found later, by stronger peoples, instead of by anyone within reach; wars and fleets rose with the longer lives. Pacifists survey least, since the cost is a military level.

**Diplomacy travels at light speed.** Proposals, declarations, peace offers, calls for help and shared intelligence are messages that arrive after distance in years and are answered after as long again; the Voice skips the lag. At the fine tick this only bites over hundreds of light years, so its real work is narrative: the call that reaches the ally after the home has fallen.

**Pacts** have a kind, defensive, aggressive or both, and optionally a named enemy, a people or a horror. Confederates propose defence against a threat both can see; conquerors and the vengeful propose aggression against the people they mean to fight. Answers come from the appraisal with posture on top: pacifists accept only defence, opportunists join the stronger side, xenophobes refuse the different, conquerors accept aggression only against someone they already want. Defence is an expedition sent to the member's home or threatened world; once stationed its level adds to the host's defence of every world within its reach from the station, and at the war's end it goes home, stays as garrison, or goes native. Each ally keeps what it takes. Intelligence is *usually* forwarded to allies but never automatically: forwarding is a choice, and the faithless withhold.

**Honour** is a new trait group: true to their word (30, betray almost never), practical about promises (50, when the profit is large and keeping faith costly), faithless (20, whenever it profits). Posture multiplies it: opportunists and conquerors twice as ready, the vengeful never against a faithful ally and always against a betrayer. Betrayal has three shapes: not coming when called, a separate peace, and turning, a stationed force seizing the world it guards. **Reputation is galactic history**: every betrayal is recorded on the world with year, victim and shape, every promise kept under real cost as renown, known to everyone, weighed in every pact with a half-life of two hundred thousand years. A betrayed vengeful people carries it as its deepest grudge.

**The decision layer** is rebuilt before the war build as four layers, so every later decision costs one scoring line instead of a special case. *Beliefs*: a people reads only its own dated intelligence about anyone else, as an estimate and a spread; fog is the only way anyone sees anyone. *Dials*: every trait sets a handful of numbers at birth (aggression, risk tolerance, greed for worlds, fear for home, loyalty, hunger for information, patience, hate for the different), scars and morale nudge them, and decisions read dials, never traits; the trait-to-dial table is the one place a temperament is tuned. *Appraisal*: one estimate of a fight against a target from beliefs, returning odds with a spread: effective strength (Military at home minus expeditions away, plus stationed relief, plus allies who would actually join, minus the enemy's other wars, dark age or plague), terrain (home bonus, defence grid, lag: a war prosecuted only by a four-thousand-year crossing is discounted steeply). A risk-tolerant people acts on the mean, a risk-averse one on the pessimistic tail, which is the whole of deciding on incomplete information. *Options*: at a council cadence and on events (a contact, a strike, a message, a call, a fleet seen), a people scores its candidate actions (declare, expedition, scout, pact, relief, forward a report, turn on a host, sue, renew) as gain times the dial that wants it minus cost times the dial that fears it, then a lottery over those above their bars, one new war at most per council. Posture sets the bars: opportunists above three in four, conquerors above two in five with a small chance honour compels them lower, the vengeful above three in ten against the grudge target and nobody else (the cost term dropped), xenophobes above a third against the different. Nobody launches an expedition the appraisal says cannot take its first world; a conqueror that scouted and stayed home gets a line and watches the target until its beliefs change. Forwarding a report scores loyalty times the ally's need minus greed times the forwarder's own designs on the subject. The same appraisal decides peace (when the odds of taking more fall below the cost) and steers research before any war: a people that sees a neighbour it cannot beat tilts toward weapons and defence, seeks a pact if confederate, bends the knee early if submissive. The research lottery, leap weights and filter adjustments stay on traits as they are; the encounter switch in contact becomes the first council after a meeting. Tuning: a decision-trace flag that prints each council's options with odds, spreads, bars and scores, and batch counters by posture and honour (wars, expeditions and their fates, scouts, pacts, betrayals, restraint).

**Contact** is measured from holdings, not homes: a colony at the edge of one reach can touch a colony at the edge of another. Roughly doubles contact, makes border wars the common kind, and gives parasites hosts.

**As built (2026-09-16), where it differs from the design above.** The defender fights with the whole world's industry behind it (+1) and the home with everything it has (+2.5 more), and the home is struck last whatever its distance, since everything between defends it. Strikes are campaigns, not raids: a quarter per thousand years at no lag, falling with the crossing time. Will drains 0.05 a thousand years, twice that past fifty thousand, not at all while a fleet is in flight (nobody tires of a war that has not begun), and judgement waits for the fleet. Capitulation cedes the winner's gains plus two worlds, not the whole front, and a fleet at base counts as a front for it; a xenophobe takes no terms from the different and scours them. Xenophobia's difference threshold is 2.5 (another kind, or a hive and strange senses), not 1.5, which had xenophobes exterminating nearly everyone. **Signal contact**: two peoples at the atomic age or later hear each other by radio across twenty light years, forty for the interstellar, from any holding; they know of each other, trade, and may go to war by fleet, and meet in the flesh later when territories touch. Allies are called once per war. Wars get a name from the most contested star when a world is taken, or from the home that fell.

**What the reference batch shows (seeds 1 to 10, `reports/tech`)**: 35 wars per world, most in the coarse pass where a war resolves within a tick; fine-pass wars last 6 thousand years at the median, quartiles 2 to 15; one world changes hands per war at the median. Endings: enslaved 42%, peace 31%, the fall of a side 23%, vassal 4%, capitulation 1%. A quarter of wars are the second or later between the same two. Conquerors declare about one war each and send a fleet in a third of their lives; five fleets in ten worlds went native. Contact is still the limit: a people meets 1.5 others and half never meet anyone, so pacts (0.1 per people), relief (13 fleets in ten worlds) and betrayals (7) are rare. Postures read as intended: pacifists never take a world and are the most often capitulated, conquerors take the most and rule the most, opportunists lose the most, the unyielding live shortest.

**Parasites, settled (2026-09-16).** They stay a normal people that rides the beaten. Two sub-types, one per parasite: riders of the flesh and riders of the mind, the second an idea that lives in others' heads and needs no ship, so it infects by signal. Rising against a parasite master is four levels harder than against any other, and the cure is a science: for the flesh, medicine and genetics; for the mind, memetics, with a small help from religion, since the old prayers turn out to be worth something. The revolt lines say which: the drug that kills what rides them, or learning to unthink it.

### Nomads

Built 2026-09-16. A people in twelve, any kind but a planetary mind, is born with the way: nomads, who will not stay. On the ground they are a wandering culture, a tilt to propulsion, industry dear, the Weight and Overshoot easier, plague harder, the Distance nothing to them. They **take to the sky** at the first interstellar reach, or sooner if the sun fails, or rather than yield or lose the home in a war, since a nomad never capitulates; it leaves. The cradle is abandoned and stays on the map as a trace, the empty cradle of the X, for the next people to arise among.

From then on the people is a set of **fleets**, each with a strength and a base, any star. The fleets are the people's Military; the tree sets the quality they grow toward. A star feeds a fleet for four thousand years, then it moves on to a star within a hop, unowned or a trade partner's by preference, and a fleet with nowhere to go thins. Fleets above four split, fleets at one base merge, a scout is a level off the largest. The seat moves with the greatest fleet, so contact, the front and the Find work from wherever the horde is. Research is pooled across fleets. Nomads trade readily and **carry what they know**: now and then a node passes between the fleets and a settled partner within a hop, in either direction, which is how the tree travels a sparse field.

At war a nomad people's front is the settled worlds within a hop of its bases, and it strikes with the fleet nearest. A won world is **stripped**: its share of the loser's Military becomes a fleet of the horde and the world is left empty, home included, the loser swallowed by the horde if it was the last. Settled peoples strike at the fleets in their reach and break them by a third at a time; a people is extinct when its last fleet is gone, and cannot be enslaved. A horde's will is half anyone's, and its peace is leaving; a settled loser's capitulation to a horde is stripped worlds and the horde moving on. Schism among the aloft splits the fleets into two peoples. A cosmic blast kills every fleet in its radius.

**Refugees** (2026-09-16): a settled people whose last world is lost while it has any reach takes to the sky under the same rules, without the way: a fleet of a third of what it had, at the nearest empty star, and a people no longer of any world. Swallowed by a horde, sterilised by a failed city, half of those scoured by a xenophobe. Unlike the born nomads they want a home: a refugee people settles the first empty world it can live on at two percent a thousand years, and on any scar at half. In ten worlds three peoples got away this way; the rule is rarely reached because most extinctions leave nothing to get away.

**Coming to rest** is uncommon and comes from a scar: one in five scars taken aloft settles the people on an empty world it can live on at or near its greatest fleet, and it is a nomad no longer, for good. A horde that contracts under the Weight rests the same way if it can, else drifts as a remnant with no world of its own.

**What the batch shows** (ten worlds): 161 born with the way, 47 flew (about three in ten, which is the share that reaches the stars), 27 came to rest (mostly the Weight). Those who flew lived 5.9 Myr at the median against 0.7 for nomads who never got off the ground, because the sky has none of the killers a world has. Strips and broken fleets are still rare, as contact is.

### Tech tree

High level only: Industrial Revolution, not the internal combustion engine; Slow Interstellar Travel, not a particular drive. Around eight **domains**: energy, industry, computation, biology, society, propulsion and communication (reach), weapons, exotic physics. Start with 50 to 100 **nodes** with prerequisites; the tree can grow later, but richness should come from how nodes, traits, scars and filters combine, not from node count.

A node has: domain, prerequisites, effects on the three levels and on reach and envelope, changes to discovery weights in other domains, structures it unlocks, a chance to proc a filter on discovery, and flavour text per kind.

**Discovery** was random each tick in the first build. Rebalanced 2026-09-16 into a **pursuit**: a civilisation picks one node it can reach and banks research points toward it until the price is paid. The pick is drawn from weights that are species tilt times current filter focus times **momentum**, a bonus for every node already known in that domain, so a people specialises and its build shows. Prices rise steeply with depth (era 0 to 4: 0.6, 2, 6, 30, 250 points; the deep spine nodes 500 to 600) while the rate stays around 0.15 to 0.5 points per thousand years, so the whole tree takes far longer than a lifetime. **Necessity steers**: a people with no habitable star within reach drops what it was doing and pursues the cheapest propulsion node, which is what keeps specialists from dying at home. The old scalar tech is a derived era label.

Era 4 is the deep tree: each domain has a spine that ends in something only a long-lived people reaches (exotic matter, causal physics, world engines, substrate minds, panspermia, the long thought, near-light travel, stellar weapons). Deep Time sits at the end of the exotic spine with a patience gate and a 30% chance per finished pursuit.

Observed after the rebalance (eight seeds): civilisations past the interstellar era know a median of 8 to 12 of the 44 deep nodes and about half of that in one domain; 0 to 4 per run know the whole tree, all of them ten-million-year survivors; relativistic travel is known by a quarter of civilisations rather than two thirds; regional powers of 5 to 14 systems number 16 to 33 per run and empires of 15 or more only a handful, most of them miracle holders.

**Structures** are built within reach, require nodes, and give levels: orbital defences give military, arcologies and closed ecologies give survival, an ansible net gives social across distance, a Dyson swarm multiplies energy and so discovery. Flavour by kind.

### Aptitudes: what a people can and cannot learn

Traits, home world and kind act on the tree through one table (`internal/history/aptitude.go`), a row per node and condition. Five words:

- **Dear or cheap.** A price multiplier that also divides the pick weight, so cheap nodes come early and dear ones drift to the back. Multipliers stack without a cap: an eyeless people on a hothouse world takes an age to find the sky, and should.
- **Innate.** Had from birth with its effects, its filter never faced. Networks for a hive; computers, machine minds and hibernation for the machine-born.
- **Moot.** Counted as had for prerequisites, otherwise not on their tree: no levels, no tilt, no filter. States and mass politics for a hive or a swarm, burial for a planetary mind, medicine for a machine. Nothing is closed to a hive by this; the Long Thought and the Chorus stay reachable.
- **Never.** Impossible because of what they are, counted as had so the tree goes on, with the dependents made dear by their own rows. Star Gazing for the eyeless, with astronomy at three times and physics at double, forever. Only body-given rules are permanent.
- **World-blocked.** The cradle lacks something: a sea on an arid world, fire on a cloud deck, a sky under ice or poison. The block holds while the people live on the cradle alone, and lifts at three times price once a colony offers the need, with a line in the legends about the bafflement of home. Two worlds get a node of their own instead, since their people cannot reach a colony while blocked: the Breach for the ice shell and High Air for the hothouse.

**Stand-ins** let a prerequisite be met by another node: Cold Chemistry for Fire (and Cold Chemistry is only for the fireless now), the Gathering for States, Seed-clouds or Living Ships for slow interstellar travel, Broodline or Forking for genetics. A stand-in must really be held; what one can never have stands in for nothing (the bug that made Life Extension free for everyone on the first run).

Star Gazing (era 0) and Astronomy (era 1, behind the Scientific Method) are separate, so a people that cannot see the sky still does science. The Scientific Method needs philosophy, printing and mathematics. Culture is on the tree: Burial, Song, Religion, Philosophy, Law, Organised Religion at era 0, and Doubt at era 1, which fires the Wars of Faith whenever it comes. The cultural nodes carry little level, since a level every people gets is inflation, not character: the first batch with them gave everyone 2.4 Social and quadrupled lifetimes.

At birth the legends say what the table does to a people: what they will never do, what they take to as if born to it, and what will come hard.

### Miracles

Agreed and built 2026-09-16. The science-fiction powers (an ansible, directed evolution, faster-than-light travel, a matter-unmaking weapon, memetic dominance, foresight) used to be ordinary tech nodes that half of all civilisations learned and that added a point or two to a level. Now they are a class apart, called **miracles** in the legends: the Voice, the Flesh, the Door, the Unmaking, the Chorus, the Sight. In code they are tech nodes flagged `Miracle` so that knowledge, closure, relics and the Find work unchanged.

**What they do.** A miracle is the dominant fact about whoever holds it, and it is most potent from a low or middling position: an end-game empire without one can hold its own, since levels are capped and its reach and arts are already high, but a young people with one becomes a regional power within a million years. The Voice adds three to Social, cancels the Distance, multiplies reach by one and a half and research by one and a half, and holds vassals at any distance. The Flesh adds three to Survival, widens the habitable envelope by two, gives reach twelve without ships (they breed bodies that cross the dark), doubles expansion and ends plague. The Door gives reach 45 and near-instant voyages, and its colony hop is its whole reach. The Unmaking adds four to Military, wins strikes at plus three, unmakes worlds rather than glassing them, and makes hostile peoples bend the knee on contact. The Chorus adds four to Social, turns most contacts into vassals within a generation, makes revolt against the holder nearly hopeless and shrugs off beacons. The Sight adds one to every level, steps around half of all filters outright, and adds two to strikes. Every miracle also multiplies research by one and a half.

**The surge.** Gaining a miracle renews a people (the Weight of Ages counts from the gain, not the birth) and starts a three-million-year surge in which expansion runs at two and a half times and research at one and a half. The surge comes once. For the born it begins when they first have reach, since a pre-industrial people cannot use a Door it cannot walk through.

**Three ways in.**
1. **The leap.** After climbing the miracle's spine (two deep era-4 nodes) a people may choose the miracle as its pursuit, at a price of 1500 points, five or six million years of work at a strong civilisation's rate. The choice follows temperament (the curious reach, the cautious do not, the martial reach for the Unmaking, the expansionist for the Door) and size, so it is the wide and the old who leap. A people holding one miracle never reaches for another. The legends say "turn everything they have toward the Flesh" when the pursuit begins, and many die still reaching. Two or three leaps per run.
2. **The find.** Elder artifacts stand for a miracle 40% of the time, and this is the bargain of the elder legacy: what is dug up may be a better water purifier or may be the Door. The elders lived where life is, so half of what they left lies under worlds that become someone's cradle and is found beneath their own cities. A miracle artifact is easy to wield (it is plainly a thing to be used) and hard to master; wielding gives the miracle for as long as the artifact lasts, mastering gives it for good along with its whole spine. Failure fires the miracle's own filter at a penalty rather than a generic beacon. This is the common route, five to fifteen per run, and the wielded Door is the usual story of a young people that suddenly holds twenty systems.
3. **Born.** Under one percent of species are born to a miracle (minds that speak across any distance, masters of their own flesh, touched by foresight, whose thought takes root in any mind, who walk between the stars), with no tree beneath it. What is evolved was never reached for, so there is nothing to fall from: the born never face the miracle's filter. One or two per run.

**Every miracle carries a filter**, faced on the leap or the find. The Open Line (the Voice): other, older voices on the line; at worst the people become one voice and a beacon goes out from their home. The Brood (the Flesh): a change nobody meant, or a remaking into a successor species that holds the same worlds. The Door: what comes back through it (the existing filter). The Unmaking: a test takes a world, or the weapon turned too close unmakes the home. The Chorus: one thought forever, or the thought gets loose as a beacon. The Sight: a fatalism that never lifts, or a people that saw what was coming and sat down to wait for it.

**Tuning lessons.** Costs alone cannot make a finite tree rare; with the pursuit model the first cut let half of all civilisations leap, and the fix was steep prices plus one leap per people. Miracle holders that leapt at the end of life just died of the Weight, hence the renewal. Elder miracle artifacts were being unleashed by young finders into beacons, and beacons feed hazard, which fails more filters, which spawns more beacons; wielding had to be easy and the beacon weight on hazard was lowered. Ordinary expansion collapsed when specialists stopped learning propulsion in a star field where a people needs about 20 light years of reach to have any neighbour; necessity now steers pursuit. A zero-weight fallback in the chooser was quietly handing miracles to civilisations that had exhausted their tree.

### The galaxy as a map

Built 2026-09-16 in `internal/galaxy` (milky.go, features.go, catalog.go, system.go, field.go).

**The model.** The Milky Way is a fixed substrate in galactocentric coordinates (kpc; the Sun at X = -8.2). It is the textbook picture, not a survey fit: a boxy bulge and a long bar at about 30 degrees, four logarithmic arms (Scutum-Centaurus and Perseus rooted at the bar's ends, Sagittarius-Carina and Norma-Outer as the minor pair) plus the Local arm the Sun rides the inner edge of, a thin disc that flares outward, a thick disc, a thin halo, and the central molecular zone around Sagittarius A*. Zones by position: the Heart, the Core, the Bulge, the Bar, the Inner Disc, the Middle Disc, the Outer Disc, the Rim, the Halo.

**The laws.** Every place has a `Law`, all values relative to the Sun's neighbourhood: density (how close the stars stand), youth (star formation: giants, supernovae, nebulae), metals ([Fe/H], richer inward, poorer outward and in the halo), glare (the hard sky: cosmic rays from young stars, the X-ray glow and outbursts of the centre), crowd (how often other stars pass close), exotic (dead and collapsed stars nearby to learn from). What they do in the engine:

- Density sets the field's radius, so a fixed number of stars stand at the right average distance. The Heart is 400 stars in 38 ly; the Rim is 400 stars in 375 ly. The reach a people needs to find a neighbour follows.
- Metals set the share of stars with worlds of rock (`Rocky`), which scales the rate life arises and complex life follows. Metals also speed industry.
- Youth scales the massive share of the class mix (so supernovae), and the gamma-ray burst rate (capped at five times the Sun's; bursts are beamed and rarer than deaths).
- Glare lowers the rate of life and raises the baseline galactic hazard (1 at the Sun, up to about 1.8 at the Heart), which every filter reads. Near the centre the hole flares now and then and every world in the field faces the burning sky at once. A people born under a hard sky is born hardy and finds the burning sky a smaller thing: the law adapts what it does not kill.
- Crowd scales the passing-dark-mass rate, capped at thirty times.
- Exotic speeds research in the exotic domain, up to two and a half times next to a black hole.

What the survey shows (`-map`): going inward from the Sun the stars crowd, the sky hardens, and lives shorten; the Bulge is dense and old and short-lived; the Heart is nearly dead (a couple of civilisations per age, none lasting); the Core has dozens, none past three million years. Going outward the stars thin, the sky softens, metals fall; the Perseus arm and the Outer Disc grow the longest-lived peoples in the galaxy, alone. The Rim and the Halo are lonely, poor and safe. A globular cluster is a ball of ancient metal-poor stars three light years apart with comets shaken loose every few hundred thousand years: a handful of peoples per age, and every one of them meets the others.

**Black holes.** A stellar-mass black hole is a quiet thing unless it is feeding; the law it carries is exotic: the deep physics comes sooner to a people who can study one. A region anchored on a black hole, neutron star or magnetar (`-at "Cygnus X-1"`) has it at the centre of the field as a star of class N, and the legends measure distances from it. Sagittarius A* is different: its glare and flares make the Heart lethal. No tidal or navigational hazards at the field scale; a field is light years across and a hole is kilometres.

**Features.** About seventy named real things with positions and two descriptions each (how a people living near it would tell it, and what is actually known): the centre and its clusters, the bar and the rings, the nurseries of the arms, the doomed giants (Betelgeuse, Eta Carinae, WR 104 with its axis pointed at us), the dead stars near the Sun (Vela, Geminga, RX J1856), the nearest black holes (Gaia BH1, BH3, A0620-00, Cygnus X-1, V404), magnetars, remnants, globular clusters, the Sagittarius stream, the two Clouds and Andromeda. A region lists the features within their reach as "Near" and the extragalactic ones as "Beyond". Dated events (the Vela supernova 11,000 years ago, the Crab in 1054, the magnetar flare of 2004) are added to the legends once the present is known; they are flavour, and only for fields close enough to have seen them.

**The catalogue.** `cmd/mkcatalog` builds `catalog_data.go` from the HYG database and the NASA Exoplanet Archive: about 800 real stars within 150 ly (every proper-named star, every known planet host, everything within 25 ly) with their known planets. The Sun's field seeds 55 percent of its stars from the catalogue (best known first) and fills the rest to the laws; nothing procedural is placed within 20 ly of the Sun. Real stars keep their names when a people arises on them. Nearby regions are real; distant ones are statistical, as the substrate notes always said.

**Star systems.** Every star has a system: known planets from the catalogue, the rest drawn from class, metals and a little astrophysics (snow lines, tidal locking, hot Jupiters where the metals are, small stars making small worlds). The habitable world is a real world in the system, and the species' home archetype (lush, twilight, superterran, floater, iceshell...) is chosen from what the world is: tidally locked worlds around red dwarfs, heavy worlds for super-Earths, cloud decks where there is a warm giant. A world Earth has seen stays the home. At a real star a procedural habitable world is only placed where the surveys would have missed it (no known planet within a factor of two in orbit), and the legends say so. A people seeded on a star with no habitable world gets one made. The legends print each cradle's system when a people arises, and a gazetteer of every star with a history at the end.

**Traps found.** Arm azimuth ranges must include the Sun's azimuth or an arm vanishes from the near side. GRB rate scaled with youth uncapped makes the Core sterile forever (a burst every 300 kyr wipes complex life faster than it can return). Crowd uncapped makes the Bulge a place where nothing lives past half a million years. A brown dwarf host from the exoplanet archive with no spectrum and no magnitude reads as a naked-eye K star unless the tool says otherwise.

### The state beneath

Agreed 2026-09-16, built in `internal/history/beneath.go`. Lore: every miracle that breaks causality (the Door, the Voice, the Sight, the Unmaking) reaches into the same thing. It is not a place but a state, the one the universe is in underneath this one, where distance and sequence are not features. The Flesh and the Chorus are not part of it; they are biology and language pushed very far, which keeps "very advanced" distinct from "wrong".

Rules, all from the user:

- Nobody born to this age understands it, and the code enforces that: there is no name for it anywhere, only the word each people coins on first reaching in (the Grain, the Quiet, the Lull, the Absence...). The legends use the finder's word. The only comprehension there ever was is elder; the law legacies are places where the state shows through, and the corridor along which ships arrive before they leave is one of them.
- Nobody is ever consciously inside it, not while staying alive and sane. A crossing has no duration from inside: one moment here, the next there. No one can say what it was like. The ones who tried to stay awake for it did not come back as one person.
- Perceiving it as something real is always bad. It means something is leaking through.

Mechanic: the wall. Holding a causal miracle wears it (the Door most), and every scar, fall, broken law and unmade world tears it; it heals with a half-life of a few hundred thousand years. It is one number for the whole field, so a people's use of the Door thins the wall for everyone. The stages are whole, worn, thin, torn; each first crossing gets a line nobody in the field could have written. A thin wall raises the galactic hazard, makes the four causal filters harder, and leaks: some of a holder's people begin to see their word for it as a place with a shore and a weather, and face the miracle's filter again; a sleeper wakes because the wall is thin near it; or something speaks from an empty star in a voice that did not cross space to get there. A born miracle never leaks; what is evolved is not a reaching-in.

Tuning: the first cut tore the wall two million years into every age and doubled the horror count. The balance is a few leaks per age and a wall that is worn or thin through the middle of the age and whole again by the present.

### The cycle

Agreed 2026-09-16. The age separator is attrition, not catastrophe.

- At any time there is a chance that a complex biosphere produces a spacefaring species. That chance is scaled by the galaxy's **fertility**.
- When an age begins there is a sudden, unexplained **surge**: fertility is 1. Many civilisations arise almost at once, many fail, everything happens at once. From the surge on, fertility decays every tick (exponential, e-folding time the **fade**, 16–30 Myr per world) until it is close to zero. Cosmic events, horrors and wars only sweep up the remains. The actual age ender is the galaxy becoming less fertile and the remaining civilisations gradually declining: the fading also weighs on the living, as a multiplier on how often the Weight of Ages is faced (up to ×3 at zero fertility).
- After a very long time (the **period**, 0.9–1.8 Gyr per world, with ±10% jitter per turn) the surge happens again and history repeats.
- **The present is found, not fixed** (agreed 2026-09-16). The engine starts at the dawn (year 0 in the code) and runs until decline has truly set in: no more than five civilisations active, and fertility below a threshold drawn per world between 5% and 20%. It then lingers a random 0–2 Myr, and stops the first time no more than five are active again. That moment is the present, and all years are printed relative to it. The age's length is therefore an output: 30–75 Myr across the first ten seeds, with the present at 6–18% fertility and one to five civilisations still standing. The tick switches from 20 kyr to 1 kyr when the waning sets in (twelve or fewer active and fertility below half), which is where the legends' "youth" and "waning" sections split. A cap of eight fades stops a run that never winds down and prints a warning in the header; that would mean an immortal-civilisation bug, not something to hide.
- Nobody knows why. A few very advanced species learn *that* it happens and where in the turn they stand: the tech node **Deep Time** (prereqs star lifting, wormhole physics, the ansible; needs 3 Myr of existence and lands with 0.2% per attempt, so one to three species per run). Its legend line reports how long ago the galaxy woke, the fertility now, and when the next surge comes, "they will not see it". Elder civilisations in the myth occasionally learn the same thing.
- Myth ages are the earlier turns of the same cycle: each is a surge whose elders rise early and fade, an "age wanes" line, and a mop-up event. Legacies erode with deep time (survival exp(−age/5 Gyr)), so old ages leave less.
- The legends header reports the cycle: period, fade, when the age dawned, fertility now, time to the next dawn.
- **Words.** In the code the moment is a *surge* and the decline is the *fade*. In the fiction, the civilisations that discover it call the surge the **dawn of the age** and its early part the **youth of the age**; the present is the **waning**. The legends use the fiction words.

Observed in the first runs with the cycle: spawn histogram falls from ~100 per 10 Myr at the surge to ~10 in the last 10 Myr; total 200–380 per run; the number active peaks around 30 in the youth and hovers between zero and eight through the waning. Deep Time lands with 0–7 species per run. Use `go run ./cmd/worldgen -seed N -stats` for one line of these numbers per seed and `-debug` for the active count every million years.

### Remains: what this age leaves for itself

Agreed 2026-09-16. The current age leaves legacies for its own successors, in the same record and through the same Find as the elder ages. The fiction is the middle ages living in Roman buildings without knowing how to raise them (wielding), and then a renaissance recovering the lost arts from what was left (mastering). The generic word is **remains**; "ruin" is a specific condition.

- **Works stand at stars.** A structure is built at a particular system (a people raises one every few hundred thousand years, at most two of a kind). When the system is lost, the works there become **remains**: structure legacies with the civilisation as maker.
- **Relics.** A civilisation that goes extinct leaves, four times in ten, an artifact of its latest art on its home world. One that falls into a dark age leaves, six times in ten, a relic of the highest art it just forgot. Relics need at least atomic-era arts.
- **The ending decides what survives.** Every filter, and every manner of losing a world (war, blast, dying sun, horror), carries a **wreckage**: the fraction of works destroyed outright, and the **condition** the rest are left in. Walking away (distance, weight, plague, transcendence) destroys a tenth and leaves things abandoned. Overshoot, machines, infection leave derelicts. Atomic war, lost wars, revolt, a failed door destroy over half and leave wrecks. Blasts, dying suns, replicators and a broken star destroy everything.
- **The condition ladder**, best to worst: **abandoned** (whole, everything works, bonus on all attempts), **derelict** (bad shape but usable, plain checks), **wreck** (repairable with much work; penalty on master and wield; a wielded wreck becomes a derelict), **ruin** (nothing usable; only mastery is attempted, at a heavy penalty, and failure simply fails; cannot be wielded, sealed or unleashed). Below ruin the record is lost. Destroyed is not a state, only the verb in the event.
- **Decay.** Every tick a buried remain may drop one step, at a base rate of one step per 2.5 Myr scaled by the thing's **hardiness**: Dyson swarms 0.3, defence grids 0.7, relays 0.9, arcologies 1.2, shipyards in their precarious orbits 1.6; vaults 0.4, workshops 1.3. The hardy outlast the age and become, by survivorship, the elder legacies of the next. Elder legacies have hardiness zero and never decay, which is what makes them elder; they start abandoned or derelict.
- **The Find handles remains.** Remains whose art the finder already knows are not Finds; a usable structure like that is simply put back to work when the star is settled (**takeover**; wrecks and ruins are not). Otherwise the finder attempts master, wield or seal as for elder legacies, at lower difficulty (made to be understood by minds of this age) adjusted by condition. Wielding a structure means moving in and keeping it running without being able to build another; it becomes a work of the finder and returns to the substrate, in whatever shape the next ending leaves it, if that star is lost. Mastering grants the tree up to the node with every filter attached: the renaissance. Unleashing a same-age structure is a small blast.
- **Kinship.** A people that finds its own works from before a dark age looks for them first (found within tens of thousands of years, not by chance), prefers to master, never seals, and gets a large difficulty bonus: the design is theirs, the thinking is theirs, and something in them remembers. A branch of the same species gets a smaller bonus. Others who never met the makers give them a finder name ("the Ones Before", "the Builders of the Halls"); those who met them, or find a remnant still living, know them by name.
- Observed across eight seeds: 500–1100 remains per run, of which about 80% crumble unfound; 7–22 mastered, 1–30 wielded (takeovers included), 3–23 sealed, 11–22 unleashed; at the present 20–160 still buried, spread across all four conditions. Tuning lessons: exclude remains whose art the finder already knows, or the Find floods with trivial masteries; and keep great works few, since every one left standing is a future Find.

### The seat and the cradle

A species is fixed at birth. Its traits, body plan and home-world archetype come from the world it arose on (the **cradle**) and never change afterwards. If the cradle is lost (a dying sun, a burning sky, a broken star) and the people survive on their colonies, the civilisation is re-seated on the nearest remaining world, but nothing about the species changes: an aquatic people stays aquatic on a desert world, and the legends say so ("from Kesh, an ocean world, later seated on Vaurr"). What the new world does to them is a matter of levels and morale, not of traits. A new species only appears through the explicit paths: schism branch, uplift, breeding, transformation.

### Order of building

1. Species and home world generator with traits, levels and reach as data; legends print a species portrait.
2. Filter resolution as level tests.
3. Tech tree driving levels and reach; drop the tech scalar.
4. Contact, war, submission, enslavement, revolt, uplift.
5. Cosmic filters with blast radii. Pairs well with ingesting the real catalogue, since star lifetimes come with it.
6. Kinds as flavour tables, then the few hard kind rules.

### What the first v1 runs showed

Built in one pass as a rewrite of the history package plus two new packages, `species` and `tech`. Everything in the design above exists in some form: species from home worlds with traits in tiers and six kinds; three levels plus reach derived every tick from species, tech, structures, wielded artifacts, scars and morale; filters as level tests with margin bands (overcome at +0.5, declined below -2, scarred between) and a per-trait difficulty table; a 56-node tree in eight domains with filters attached to nodes; contact by reach overlap with peace, submission, war, extermination, enslavement, revolt and uplift; cosmic filters with blast radii, scheduled star deaths and the dying sun as a slow filter; the age generator with elder portraits, five legacy kinds and the Find with its four outcomes; horrors adapted to tests (beacon tests Social, incursions test Military, a waking tests Survival). One engine runs the middle pass (60 to 5 Myr ago, 20 kyr ticks) and the fine pass (last 5 Myr, 1 kyr ticks); every rate is given per thousand years and scaled to the tick.

Tuning findings, in the order they were found:

- **The early age used to burn everything.** Sterilised worlds never recovered, so the late age was empty. Now a blast or a glassing leaves microbial life, which climbs back to complex life in tens of millions of years, and the late age has 15 to 25 new civilisations per run.
- **Beacons chain-reacted.** Each listener that fell became a new beacon and every listener was tested every tick. Now a civilisation faces each beacon once, and hazard is capped.
- **Uplift ran at a hundred times the intended rate** and consumed every world with complex life. Now it is rare and capped at two per civilisation.
- **Immortal giants.** High-Social civilisations overcame every Weight of Ages test and lived the whole 60 Myr. Two causes: a v0 bug (the renaissance timestamp defaulted to zero, so the age term was negative until the end of history), and no cumulative cost of age. Now the Weight's difficulty grows with total age and with every renaissance, and the Distance recurs each time an empire doubles. Lifetimes now spread from under half a million years (about 35 percent, mostly the early filters, wars and being found young) through 1 to 3 Myr to a 3 to 10 Myr tail; almost nothing lives past 10 Myr.
- **Found young.** An empire with 90 ly of reach finds every new species at the stone age. Contact with a pre-atomic species is now its own event: a xenophobic elder occasionally scours the world, an expansionist or martial one sometimes takes them, everyone else watches from orbit and meets them properly later.
- **Star deaths are a fudge.** About 14 percent of F, G and K stars are scheduled to die within a few hundred million years of the present so that supernovae and dying suns happen inside the window. Documented in `galaxy.diesAt`; replace when the real catalogue comes in.

Known oddities to look at next, none of them blocking:

- Levels saturate at 10 for every exotic-era civilisation, so the legends say "overwhelming" three times for all of them and late filters lose their bite. The scale wants either a softer cap or smaller tech contributions.
- The tree lets Life Extension arrive before the Atomic Age because the two branches do not touch. Either add a cross-prerequisite or accept it as an alien research order.
- Remnants persist for tens of millions of years; a remnant that contracted 57 Myr ago is still there at the present. The fade rate should scale with era or the middle pass should end them.
- The myth section is padded with "life arises" lines; the age events should probably print alone.
- Mastery of legacies is more common than "very hard" suggests, because of the level saturation above.
- A run produces 250 to 300 civilisations in 60 Myr and 12 to 18 thousand lines of legends. The middle pass should log less, or the legends should have a "short" mode that prints only rises, falls and filters.

### Reading the tree in use

`go run ./cmd/techstats -seeds 10 -out reports/tech` runs a batch of worlds and writes three files: `report.md` (every node with its one-line meaning from `internal/tech/desc.go`, how often it is reached overall and among those holding its prerequisites, when in a people's life it comes, what its filter did; the common builds by domain pair and by frontier; what the long-lived held), `civs.jsonl` (one record per civilisation for later questions: every node ever held and when, miracles and their route, filters faced, scars, fate) and `stats.txt` (the tuning line per seed). The committed `reports/tech` is the reference batch for seeds 1 to 10 at Sol. Ten worlds is a small sample: the same code with different random paths moved the share of very short lives by six points, so read the percentages as rough.

The first batch found that seeds did not reproduce. Go's map iteration order is random per run, and several loops ranged over a map and drew from the RNG inside (blast victims, wars, forgetting, uplift, remaking). Every such loop now walks a sorted or tree-ordered list (`knownOf`, `sortedInts` in `util.go`), and a seed gives byte-identical legends. Rule: never range over a map in a path that draws from the RNG or picks an element.

## Decision log

- 2026-09-16: Galaxy idea is primary. Fairy idea is backup.
- 2026-09-16: Design notes kept in this Markdown file. Conceptual stuff only.
- 2026-09-16: Two-pass history simulation agreed: coarse deep-time pass, then fine recent-history pass.
- 2026-09-16: Alien horrors organised as categories with per-category simulation rules. Start with a small list, extend over time. Monstrous races are one category among several.
- 2026-09-16: Language is Go. Reasons: the history sim is a graph of mutually referencing entities, which a garbage collector handles cleanly; performance needs are modest; fast iteration matters more than raw speed. Rust's sum types would have been nicer for event modelling, accepted as a cost.
- 2026-09-16: "Filters" is the term for anything that may push a civilisation into decline. Three outcomes: overcome, scarred, declined. Early filters like nuclear weapons included. Scars are lasting traits.
- 2026-09-16: Filter balance is fine for now. The lethality of early filters is accepted; the fix later is scale, not tuning. With hundreds of billions of stars the sim can afford far more civilisations, so a high failure rate still leaves plenty of history. Not for now.
- 2026-09-16: Simulation v1 model agreed in full, including refinements (asymmetric traits, few traits per species, dynamic Social, Survival widens the envelope, kinds as flavour, small tree first). Dying sun is a slow filter: Survival sets how long a species endures it, reach decides whether it can leave.
- 2026-09-16: Simulation v1 model sketched (species from home worlds, three levels plus reach, filters as level tests attached to tech nodes, war with enslavement, high-level tech tree, exotic kinds as flavour). See "Simulation v1". Not built yet.
- 2026-09-16: Galactic history is a series of ages with interregna. Prior ages are coarse myth that leaves legacies (artifacts, structures, threats, sleepers, laws) on the substrate; the fine sim is the current age and its aftermath is the age ending. See "Ages of the galaxy".
- 2026-09-16: Three passes: previous ages extremely coarse, early to mid current age medium, late current age fine. The Find has four outcomes chosen by traits and tested by levels: mastered, wielded, sealed, unleashed.
- 2026-09-16: Simulation v1 built (species, levels, reach, tech tree, tests, contact, cosmic, ages, the Find). Tuned until lifetimes spread and the late age is populated. See "What the first v1 runs showed".
- 2026-09-16: A species never changes because its seat moves. Cradle and seat are separate; traits and archetype are fixed at birth. See "The seat and the cradle".
- 2026-09-16: The present is the aftermath. Every civilisation ends as gone, transformed, or contracted, always with a cause. Sol is protected by fiat.
- 2026-09-16: The current age leaves remains for its own successors through the same Legacy record and the same Find. Wielding is living in the old halls; mastering is the renaissance. A people can recover its own lost arts after a dark age, with a kinship bonus. Every ending carries a wreckage (fraction destroyed, condition of the rest); remains decay a step at a time on the ladder abandoned, derelict, wreck, ruin, at a rate set by their hardiness, so the hardy become the next age's elder legacies. See "Remains".
- 2026-09-16: The present is found, not fixed: the engine runs from the dawn until no more than five civilisations are active and fertility is below a per-world threshold, and that moment becomes now. Age length is an output. See "The cycle".
- 2026-09-16: The Long Dusk is removed. Civilisations still active at the present are reported as still standing. The horror kind formerly called Elder is renamed Sleeper (in code SleeperHorror, as the legacy kind is already Sleeper); "elder" now means only a civilisation of an earlier age.
- 2026-09-16: Research is a pursuit with steep prices by depth and momentum by domain, so civilisations specialise and nobody climbs the whole tree; necessity steers a stranded people toward propulsion. Era 4 deepened with a spine per domain. See "Tech tree".
- 2026-09-16: The science-fiction powers are miracles, a class apart: gamechanger effects, a surge on gaining one, their own filters, three ways in (leap, find, born) with the elder find the common one and the born exempt from the filter. "Miracle" is the word in the legends. An end-game empire without one holds its own; a young people with one becomes a regional power. See "Miracles".
- 2026-09-16: The real Milky Way is the map: a structural model with laws by geography (density, youth, metals, glare, crowd, exotic), a catalogue of named features, real stars and known planets within 150 ly from HYG and the Exoplanet Archive, and every star with a system of worlds the species' home is drawn from. A history runs in one field placed anywhere (`-at`); the whole map is printed by `-map`. See "The galaxy as a map".
- 2026-09-16: The causality-breaking miracles all reach into one state beneath the universe, never understood, never named in common, never consciously entered; perceiving it as real means a leak. The wall between wears with use for everyone in the field. See "The state beneath". Weird literal consequences of the physics are wanted, not sanitised conventions.
- 2026-09-16: Traits, home world and kind act on the tree through an aptitude table with five words (dear, innate, moot, never, world-blocked) and stand-in prerequisites; culture is on the tree; Star Gazing and Astronomy are split; the kinds have hard rules and nodes of their own; senses and the herd-to-solitary spectrum are traits; machine-born peoples are only ever made by a failure; parasites hold hosts as slaves. The leap is priced like Deep Time (600) and weighted to land at one or two per world. See "Aptitudes".
- 2026-09-16: Every node in the tree has a one-line description (`tech.Node.Desc`) and a batch tool reads the tree in use over many seeds (`cmd/techstats`, reference batch in `reports/tech`). Seeds must reproduce: no map iteration on any path that draws from the RNG. See "Reading the tree in use".
- 2026-09-16: Ages are attrition-based. Galactic fertility for new species surges at the start of an age and decays every tick; cosmic events only mop up; the cycle repeats with a period of about a billion years; nobody knows why, but a few advanced species learn where in the turn they stand (Deep Time). Replaces the "age-ender" catastrophe. See "The cycle".
- 2026-09-16: War redesigned (not yet built): wars as objects with a front (the overlap in reach), duration and will; strikes lagged by distance; peace, grudges and series of wars; postures replace the stance group; xenophobia by a difference score; parasites fight by infection; expeditions as armies far from home (any number, each at least a tenth of Military, thirty percent lost per defeat, successor states when stranded); scouts as one-level expeditions; dated intelligence; light-speed diplomacy; pacts, honour and galactic reputation for betrayal; one appraisal that every decision routes through. See "War, submission, enslavement".
- 2026-09-16: War built as designed: wars as objects with a front, will, strikes lagged by distance, peace, capitulation and grudges; postures replace the stance group and honour is a trait group; a decision layer of beliefs, dials and one appraisal; fleets, scouts, the watch; pacts, messages at light speed, relief, betrayal and galactic reputation; parasites fight by infection; signal contact by radio; dominion levels for masters and the held. Tuned on the reference batch; see "As built" under "War, submission, enslavement". Proposed next: parasites as homeless fleets.
- 2026-09-16: Parasites stay a normal people with two sub-types (flesh, mind) and a revolt that needs a cure. The homeless-fleet rule set lives in nomads instead: a way any kind but a planetary mind can be born with, fleets as the people, stripping, carrying tech, coming to rest from a scar. See "Nomads".
- 2026-09-16: Exploration: stars are seen but read only by a visit, the Sight or a nearby telescope; colony ships go to read stars, or on a guess; surveyors are one-level expeditions on tours, the count set by hunger and greed; the Sight is a mode, outward or at the borders; the chance lottery is nearly gone; contact in the flesh needs having heard or read the other. See "Exploration" under "War, submission, enslavement".

## Next steps (proposed)

- Fix the known oddities from the first v1 runs: level saturation, tree cross-prerequisites, remnant fade, a short legends mode.
- Proactive military scouting by posture; the survey policy shows the shape. Tuning of pact acceptance and relief, still rare. Charts do not record who was there when: a stale reading counts as known taken, so a star abandoned since is never tried.
- Read many seeds and adjust until histories feel right: how often each filter is the killer, how often the Find goes each way, whether wars and enslavement read well.
- Real stars are in (HYG, Exoplanet Archive). Next for the map: fit the arms to Reid et al. 2019 rather than the textbook picture; more features (the Gum nebula, the Cepheus bubble, the Vela molecular ridge); let several fields share one history.
- Later: scale up massively. The real galaxy has hundreds of billions of stars, so civilisation counts should go from dozens to thousands or more. This needs a different simulation structure (regions, statistical treatment of the unremarkable, only instantiating stars where something happens). Design for it, but do not build it yet.
- Add more horror categories to the sim: plagues as spreading actors, monstrous races as civilisations with alien drives, dimensional anomalies tied to FTL use.
- Add the lazy detail layer: given a star and its history, generate the system contents (planets, stations, derelicts).
