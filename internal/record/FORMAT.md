# The run directory: format 1

A run of `worldgen` writes one self-contained directory. Nothing in it is a sentence the simulation wrote; everything is an id, a number, a key that resolves in a lookup under `data/`, or a name from the names table. This file is the contract: every file, every record, every field. `internal/record` holds the Go types, and a test (`format_test.go`) fails when a field the writer emits is not in this document.

| file | what it is |
|---|---|
| `dossier.json` | the run's header: format version, code revision, seed, the place and its laws, the cycle, the present, the wall, the counts |
| `state.json` | the galaxy at the present as objects by id: stars, peoples with their knowledge, species, ages and elders, remains, sources, traces, plagues, wars, pacts, fleets, contracts, the batch readings, the names |
| `chronicle.jsonl` | what happened, one record per line, in the order it is told (by year) |
| `tellings.jsonl` | what each people holds of what happened, one record per tale, and every remain's testament as frozen tales |
| `data/*.json` | the lookups every key resolves in, copied from the binary's embedded copy |
| `codex/*.md` | the fiction for whoever narrates, copied likewise |
| `FORMAT.md` | this contract |

Conventions:

- **Ids** are small integers stable within a run, `-1` for none. Peoples, species, elders, stars, remains, sources, plagues, wars, pacts, fleets, contracts, events, tales and sightings are numbered from 0 in the order they came to be. A map keyed by an id (JSON object keys are strings) is written `{"12": ...}`.
- **Years** are integers of years since the dawn of the current age; the ages of myth are negative. The dossier's `present` is when the run stopped; "ago" is `present − year`. The tick is `step` years.
- **Keys** are strings that resolve in `data/`: an event kind in `events.json`, a trait in `traits.json`, a node in `tech.json`, a filter in `filters.json`, a scar or a boon in `scars.json` and `boons.json`, a miracle and its objects' forms in `miracles.json`, a portrait or a condition in `portraits.json` and `conditions.json`, a cause in `causes.json`, an origin in `origins.json`, a source kind in `sources.json`, a work in `works.json`, a world archetype in `worlds.json`, a kind or a power in `kinds.json`, a feature in `features.json`, a plague's symptoms, onsets, courses, forms, effects and carriers in `plagues.json`.
- **Names** appear in one place only: the `names` table of the state. Every other record refers to things by id. A star carries its human `designation` as the substrate's own label.
- **Durable and snapshot.** In the state, a field marked **D** is durable: it changes only when an event of the chronicle says so (the kinds named beside the mark), and the fold test (`internal/writer/fold_test.go`) rebuilds it from the chronicle alone. A field marked **S** is a snapshot: derived every tick (levels, flows, wear), kept for the batch reports, or the substrate's, and the chronicle does not carry its history. A field with no mark is a snapshot. The durable changes the prose never remarked are the **silent kinds** of `events.json` (`civ_born` onward): no line in the view, no weight, ids after every told event's.

## dossier.json

### Dossier

| field | type | meaning |
|---|---|---|
| `format` | int | the version of this contract: format 1 |
| `code` | string | the git revision the binary was built from, `-dirty` when the tree had uncommitted changes, `unknown` outside a checkout |
| `seed` | integer | the world seed |
| `present` | year | when the run stopped: the present |
| `truncated` | bool | the run was stopped at `-until`, before the age's own end |
| `waning` | year | when the waning was declared (the present, if never) |
| `deep_start` | year | where the ages of myth begin |
| `dawn` | year | the dawn of the current age: 0 |
| `step` | year | the tick of the current age, in years |
| `ticks` | int | ticks run in the current age |
| `capped` | bool | the age never wound down on its own; stopped at `max_fades` fades |
| `max_fades` | number | the cap, in fades |
| `config` | Config | what the run was asked for |
| `place` | Place | where in the galaxy the field is |
| `cycle` | Cycle | the galaxy's cycle of dawns |
| `fertility` | number | fertility now, as a fraction of the dawn's |
| `next_dawn` | year | when the galaxy will wake again |
| `hazard` | number | the galactic hazard |
| `wall` | Wall | the wall between this and what is beneath it |
| `counts` | Counts | the numbers the aftermath's header prints |

### Config

| field | type | meaning |
|---|---|---|
| `at` | string | the place as given: a preset, a feature name, or `x,y,z` in kpc |
| `stars` | int | stars asked for |
| `radius` | number | the field's radius, light years |
| `thickness` | number | the field's thickness, light years |
| `tuning` | object | every number the decisions use (`internal/mind.Tuning`), as the run had it |

### Place

| field | type | meaning |
|---|---|---|
| `name` | string | the place's name |
| `code` | string | the designation prefix of the field's synthetic stars |
| `anchor` | string | what the field is centred on: Sol, a feature, or the centre of the field |
| `pos` | {X,Y,Z} | galactocentric position, kpc |
| `has_sol` | bool | the field is the Sun's neighbourhood |
| `sol` | id | the star that is the Sun, or -1 |
| `real` | int | how many stars are catalogued ones |
| `zone` | key | the zone of the galaxy |
| `laws` | Laws | the numbers of the place |
| `arm` | string | the spiral arm the place is in, if any |
| `arm_dist` | number | kpc to that arm's ridge |
| `near` | [NearFeature] | catalogued features near the field, nearest first |
| `beyond` | [key] | the sky features, beyond the galaxy |
| `mean_gap` | number | mean spacing of the field's stars, light years |
| `field_radius` | number | the field's radius, light years |

### Laws

Every value but `metals` is relative to the Sun's neighbourhood, which is 1.

| field | type | meaning |
|---|---|---|
| `r` | number | galactocentric radius, kpc |
| `z` | number | height above the plane, kpc |
| `density` | number | stellar density |
| `youth` | number | star formation: how many young suns |
| `metals` | number | [Fe/H] |
| `glare` | number | the hardness of the sky |
| `crowd` | number | close stellar passages |
| `exotic` | number | dead and collapsed stars near |

### NearFeature

| field | type | meaning |
|---|---|---|
| `key` | key | the feature, in `features.json` |
| `dist` | number | kpc from the field |

### Cycle

| field | type | meaning |
|---|---|---|
| `period` | year | from one surge to the next |
| `fade` | year | e-folding time of fertility after a surge |
| `surges` | [year] | every surge in the history, oldest first; the last is the current age's |
| `floor` | number | fertility below which an age is over |
| `ends` | number | fertility below which the present may fall, drawn per world |

### Wall

| field | type | meaning |
|---|---|---|
| `value` | number | how worn the wall is |
| `stage` | int | its stage, a row of `levels.json`'s wall |

### Counts

| field | type | meaning |
|---|---|---|
| `civs` | int | peoples that arose |
| `standing` | int | still rising at the present |
| `remnants` | int | remnants still living |
| `eaten` | int | stars held by things that eat them |
| `speaking` | int | transmitters live |
| `complex` | int | worlds with complex life |
| `remains` | int | remains of this age |
| `crumbled` | int | of them lost |
| `traces` | int | traces left |
| `events` | int | records in the chronicle |
| `tales` | int | records in the tellings |
| `names` | int | rows of the names table |
| `plagues` | int | plagues born |
| `wars` | int | wars declared |
| `fleets` | int | fleets ever launched |
| `contracts` | int | contracts ever offered |

## state.json

### State

| field | type | meaning |
|---|---|---|
| `stars` | [Star] | every star of the field, by id |
| `civs` | [Civ] | every people, by id |
| `species` | [Species] | every blood, by id |
| `ages` | [Age] | the earlier ages |
| `elders` | [Elder] | the civilisations of the earlier ages |
| `remains` | [Remain] | everything the ages left, by id |
| `sources` | [Source] | everything with a yield, by id |
| `traces` | [Trace] | what was left for whoever comes next |
| `plagues` | [Plague] | every plague born, by id |
| `reservoirs` | [Reservoir] | plagues waiting in dead cities |
| `wars` | [War] | every war, by id |
| `pacts` | [Pact] | every pact, by id |
| `betrayals` | [Betrayal] | promises broken, and kept |
| `fleets` | [Fleet] | every fleet ever launched, by id |
| `contracts` | [Contract] | every contract ever offered, by id |
| `fathomings` | [Fathoming] | every understanding reached |
| `battles` | [Battle] | every battle at a world |
| `meetings` | [Meeting] | every battle in the dark |
| `sightings` | [Sighting] | every fleet seen in flight, by id |
| `names` | [Name] | the names table: the only place a name appears |
| `voices` | {id: voice} | how each people names: `none`, `transcribed`, `translated` |

### Star

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `designation` | string | the human catalogue's label for a real star, a code for a synthetic one |
| `kind` | key | what the designation is: `proper`, a catalogue's key (`hip`, `hd`, `gj`, `2mass`, ...), `catalogue`, or `code` |
| `real` | bool | a catalogued star, or a named feature |
| `alt` | string | another designation |
| `class` | string | spectral class O B A F G K M; W a white dwarf, N a neutron star or black hole (**S**: the substrate's; a death is a `star_class` event, but the first class is the galaxy's) |
| `note` | string | giant, supergiant, white dwarf, brown dwarf, subdwarf |
| `remnant` | key | for class N: `neutron_star`, `magnetar`, `black_hole`, `great_hole` |
| `x`, `y`, `z` | number | light years from the field's centre |
| `hab` | number | habitability weight |
| `mult` | int | 1 single, 2 binary, 3 trinary |
| `lifetime` | number | main-sequence lifetime, years |
| `dies_at` | year | when the star leaves the main sequence |
| `failing` | bool | the star has begun to die (**S**) |
| `mag` | number | apparent magnitude from Earth; 99 if not visible |
| `system` | System | what orbits it: `planets` (each `name`, `kind`, `mass`, `radius`, `period`, `sma`, `teq`, `ecc`, `known`, `method`, `year`, `temperate`, `moons`, `tag`; a number the generator could not compute is `null`), `belts` (AU), `disc`, `comp`, `home` (the habitable world's index), `arch` (its archetype key), `missed` |
| `bio` | key | life on its worlds: `none`, `simple`, `complex` (**D**: `bio`) |
| `held` | id | the people holding it, or -1 (**D**: `world_held`, `world_lost`) |
| `guns` | int | guns standing over it (**S**) |
| `grid_broken` | bool | its defence grid shot to nothing (**S**) |

### Species

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `sub` | key | the substrate, in `kinds.json` |
| `mods` | [key] | the modifiers, in `kinds.json` |
| `channel` | key | how the people communicates: sound, light, electromagnetic, chemical, touch, none |
| `powers` | [key] | the eldritch pool's powers held, in the order gained (**D**: `civ_born`, `power_held`) |
| `world` | key | the archetype of its cradle world, in `worlds.json` |
| `traits` | [key] | its traits, in `traits.json` (**S**: a drift is a `drifted` event, but the fold does not carry traits) |
| `made` | Making | how it was made when it did not arise: `key` (in `origins.json`), `by`, `from`, `legacy`, `plague` (ids, -1 for none); the zero key for a cradle blood |
| `parent` | id | the species it was made from, or -1 |
| `first` | id | the first people of the blood, whose name the blood goes by |

### Civ

A people. The stage and the fate are the two halves of its status.

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `species` | id | its blood (**D**: `civ_born`, `blood`) |
| `origin` | Making | how it came to be when not by arising (see Species `made`) |
| `home` | id | its seat now (**D**: `civ_born`, `seat`) |
| `cradle` | id | the world it arose on; never changes (**D**: `civ_born`) |
| `born` | year | |
| `ended` | year | when it ended, for one that did |
| `fell` | year | when it stopped rising |
| `stage` | key | `emergent`, `interstellar`, `zenith`, `remnant`, `dead` (**D**: `stage`) |
| `fate` | key | `active`, `extinct`, `transformed`, `contracted`, `sundered`, `shattered` (**D**: `fate`) |
| `cause` | key | why it fell or ended, in `causes.json`; the fall and end facts carry the parties (**D**: `fate`) |
| `into` | key | what it became, if transformed: a becoming in `causes.json` |
| `into_civs` | [id] | the peoples (or, for `species`, the blood) it became |
| `fall_event` | id | the fall fact, or -1 |
| `end_event` | id | the end fact, or -1 |
| `named` | bool | has a word for the state beneath (**D**: `named`) |
| `hosts` | int | for a parasite: peoples ridden or fighting its plague (**S**) |
| `morality` | Morality | what it counts as wrong |
| `lifted` | [key] | world blocks lifted by a colony: sea, sky, fire |
| `systems` | [id] | the stars it holds (**D**: `world_held`, `world_lost`) |
| `peak` | int | the most it ever held |
| `voyages` | [Voyage] | colony ships in flight |
| `known` | [key] | the nodes it holds, sorted (**D**: `node_held`, `node_lost`) |
| `dormant` | [key] | of them, gone dark this tick (**S**) |
| `learned` | {key: year} | when each node was learned by pursuit or find; inherited nodes are absent |
| `era` | int | its era, 0 to 4 |
| `pursuit` | key | the node being worked toward |
| `progress` | number | research banked toward it (**S**) |
| `miracles` | {key: route} | the miracles held, each with how it was gained: born, leap, found, wielded (**D**: `miracle_held`) |
| `remade` | {key: year} | when the object of a miracle was last lost |
| `levels` | Levels | the derived levels this tick (**S**) |
| `reach` | number | light years it can cross (**S**) |
| `speed` | number | years per light year (**S**) |
| `envelope` | int | stars within reach (**S**) |
| `morale` | number | the dynamic part of the social level (**S**) |
| `dials` | Dials | temperament as numbers: `aggression`, `risk`, `greed`, `fear`, `loyalty`, `hunger`, `patience`, `hate` |
| `order` | [key] | the direction of the flows this tick, by category (**S**) |
| `income`, `upkeep`, `surplus`, `want` | Income | this tick's yield, needs, spare and shortfall by kind, an object of the symbols O (organic matter), E (energy), M (metal) (**S**) |
| `structures` | {key: int} | the works standing, by key |
| `works` | [Work] | where they stand |
| `wielded` | [id] | the remains it wields |
| `found` | [id] | the remains it has attempted |
| `heard` | [id] | the beacons it has faced |
| `uplifts` | int | peoples it raised |
| `ruled` | int | peoples it has held as slaves or vassals |
| `docks` | int | docks at work this tick (**S**) |
| `salvage` | int | ships of others' make in its guards |
| `wars` | [id] | the peoples it is at war with (**D**: `war_opened`, `war_over`) |
| `trade` | [id] | its trade partners |
| `master` | id | the people that holds it, or -1 (**D**: `civ_born`, `master`) |
| `vassal` | bool | held as a vassal rather than a slave (**D**: `master`) |
| `pacts` | [id] | the pacts it is in |
| `truce` | {id: year} | no new war with each before this |
| `fought` | {id: int} | wars fought with each |
| `grudge` | {id: number} | what each other people has done to it |
| `ridden` | [id] | for a parasite: the peoples taken as hosts |
| `contracts` | [id] | every contract it was party to |
| `taught` | {key: id} | nodes bought, each with the people that taught it |
| `sellsword` | bool | lives on contract pay |
| `dependent` | [id] | partners whose sending keeps the fed uses fed (**S**) |
| `embargo` | [id] | partners it has closed its ports to |
| `refused` | {id: year} | since when each partner has been refused while it wanted |
| `barred` | [id] | the peoples whose goods it refuses for good |
| `infections` | {id: Infection} | the plagues it has, by plague (**D** for the keys: `infected`, `cleared`; the Infection's own fields are **S**) |
| `immune` | [id] | the plagues it cannot catch again |
| `suspect` | [id] | the peoples it thinks are sick |
| `closed` | [id] | the suspects it has closed its ears and ports to |
| `own` | id | for a parasite: the plague it is; -1 for a people that is not one |
| `weapons` | {key: Weapon} | the plagues it has made and holds, by the craft that made them |
| `stiff` | number | how far its ways have set (**S**) |
| `ossified` | bool | set: acting every other tick (**S**) |
| `still` | year | when something new last happened to it |
| `line` | [id] | the peoples it came out of by a sundering or a shattering, oldest first (**D**: `line`) |
| `claim` | [id] | the worlds of the old realm an heir holds itself owed |
| `asleep` | bool | the long sleep (**D**: `sleep`) |
| `slept` | year | when it last went to sleep |
| `aloft` | bool | a nomad people living as fleets (**S**) |
| `rested` | bool | a nomad people that came to rest |
| `drifts` | int | for an evolver: how many times its shape has drifted |
| `searching` | bool | the Sight is turned outward |
| `starfaring` | year | when reach first touched another star; 0 if never |
| `faced` | [key] | the filters faced |
| `scars` | [key] | the scars it carries, in `scars.json` (**D**: `scar`) |
| `boons` | [key] | the boons, in `boons.json` (**D**: `boon`) |
| `record` | [Record] | what it did that its record keeps, in order |
| `dark_ages` | int | dark ages suffered |
| `knows_cycle` | bool | learned the shape of the cycle (**S**; the `cycle` fact says when) |
| `ascended` | year | when a miracle was last gained |
| `renewed` | year | when it was last renewed |
| `renaissances` | int | |
| `dying` | bool | its home star is failing |
| `endure` | number | thousand years left under the failing star |
| `rare` | [key] | the rarities had this tick (**S**) |
| `last_dark` | year | the last dark age |
| `last_taken` | year | when a world last changed hands with it on either side |
| `last_unmade` | year | when the unmaking was last turned on a world |
| `knowledge` | Knowledge | what it knows, believes and has forgotten |
| `batch` | Batch | the counters the batch reports read (**S**) |

### Morality

| field | type | meaning |
|---|---|---|
| `kind` | key | `amoral`, `individual`, `herd`, `fixation` |
| `object` | key | the fixation's object: conquest, spawning, knowing, old things, holding |

### Voyage

| field | type | meaning |
|---|---|---|
| `target` | id | the star |
| `arrive` | year | |
| `blind` | bool | sent to a star nobody has read |

### Levels

| field | type | meaning |
|---|---|---|
| `military`, `survival`, `social` | number | the three levels; `levels.json` gives the bands |
| `wisdom` | number | the capacity to see the counterintuitive |

### Work

| field | type | meaning |
|---|---|---|
| `key` | key | the structure, in `works.json` |
| `node` | key | the node it stands for |
| `star` | id | where |
| `legacy` | id | the remain it was inherited as, or -1 |
| `dark` | bool | shed this tick (**S**) |

### Infection

| field | type | meaning |
|---|---|---|
| `since` | year | |
| `from` | id | the people it came from, or -1 for born or woken |
| `road` | key | how it came, in `plagues.json`'s roads |
| `contained` | bool | this tick's cure roll held it |
| `held` | int | ticks contained in a row |
| `carrier` | bool | carries it without toll and never fights it: a cult |

### Weapon

| field | type | meaning |
|---|---|---|
| `plague` | id | |
| `target` | id | the people it was made for, or -1 |
| `node` | key | the craft |
| `made` | year | |

### Record

| field | type | meaning |
|---|---|---|
| `kind` | key | faced, foresaw, mastered, bounty, crewed, wielded, sealed, unleashed, sky, rest |
| `filter` | key | for faced and foresaw |
| `outcome` | key | for faced: `overcame`, `scarred`, `declined` |
| `narrow` | bool | the margin was slight |
| `legacy` | id | the remain, where the kind names one |
| `known` | bool | whether the people could place the remain's makers then |

### Knowledge

What a people knows, believes and has forgotten: the only place ignorance lives in the output. The record and the chronicle are omniscient; a consumer reads the player's view through these.

| field | type | meaning |
|---|---|---|
| `met` | [id] | peoples known of, by touch or by signal |
| `reached` | [id] | peoples met in the flesh |
| `fathomed` | [id] | peoples it understands |
| `fathom_tried` | {id: year} | when it first tried to understand each |
| `fathomed_at` | {id: int} | the drift count of each people fathomed, at the fathoming |
| `alien` | {id: number} | how alien each people tried is, as the fathoming counts it |
| `perceives` | [id] | the peoples it can hold in mind at all, of every people there is (the anti-memetic are missing) |
| `charted` | {id: year} | stars read, with the year of the reading; a stale chart is a belief |
| `marked` | [id] | stars the Sight showed something at |
| `scouted` | {id: year} | when a scout last reported on each people |
| `intel` | {id: Intel} | what it believes of each other people |
| `regards` | {id: int} | how it holds each people now: -2 a monster, -1 an enemy, 0 a stranger (absent), 1 a friend, 2 its own line |
| `monsters` | [id] | peoples it remembers as things that do harm |
| `dread` | [id] | stars it will not go to |
| `scapegoat` | id | the enemy of the day, who gets the old blame, or -1 |
| `watched` | [id] | peoples looked at hard and left alone |
| `sightings` | {id: id} | fleets it has seen, each to its sighting record |
| `memory` | Memory | the standing count of its tales |
| `lore_dials` | Dials | what the telling does to its temperament |

### Intel

| field | type | meaning |
|---|---|---|
| `mil` | number | the level seen, with the noise of the seeing |
| `ships` | int | the guard seen at `star` |
| `guns` | int | the guns seen over `star` |
| `total` | int | the ships seen manned everywhere |
| `relief` | number | ships of others seen standing with them |
| `star` | id | where the look was taken |
| `year` | year | when |
| `sick` | id | a plague seen raging in them, or -1 |

### Memory

| field | type | meaning |
|---|---|---|
| `held` | int | tales held |
| `myth` | int | of them worn to myth |
| `forgot` | int | tales forgotten over its life |
| `revised` | int | tales rewritten because a regard changed |
| `blamed` | int | times blame was moved |

### Batch

Counters kept for the batch reports (`cmd/techstats`); snapshot only.

| field | type | meaning |
|---|---|---|
| `tally` | Tally | what the people did in war and peace |
| `high_income`, `high_upkeep`, `high_want` | Income | income, upkeep and want at the height of its means |
| `peak_trade` | [id] | the partners at the height of means |
| `shed_ticks` | {key: int} | ticks each node has spent dark |
| `built` | {key: int} | structures raised, by key |
| `had` | [key] | every rarity ever had |
| `harnessed` | [key] | every source kind ever harnessed |
| `granted` | [key] | nodes learned while their grant was had |
| `fell_dependent` | bool | depended on a partner at the moment of its fall |
| `peak_wis` | number | the most Wisdom ever had |
| `wis_from` | [5 numbers] | where the Wisdom comes from: species, tech, experience, boons, scars |
| `peak_ships` | int | the most ships ever in being |
| `first_plague` | year | when it first had one |
| `sire` | id | the people that raised or bred it, or -1 |
| `dock_rate` | {id: number} | the rate each dock worked at last tick, by star |
| `want_ships` | int | the need of the campaign the council last could not man |
| `garrison_want` | int | what the garrison policy asked for last tick |

### Tally

Every field is a count, but where said; `sent` and `got` are Incomes, `searched`, `sighted`, `building`, `acted_gap`, `slights` and `oss_stiff` are sums.

| field | meaning |
|---|---|
| `declared`, `fought`, `taken`, `lost`, `glassed` | wars declared and fought, worlds taken, lost, glassed |
| `fleets`, `native`, `scouts`, `relief` | fleets sent, of them from the home world; scouts; relief fleets |
| `pacts`, `refused`, `betrayals`, `called` | pacts sworn, refused, betrayed, called |
| `capitulated` | ever yielded |
| `surveys`, `charted`, `blind`, `blind_lost` | surveys sent, stars read, blind colony ships, of them lost |
| `find_survey`, `find_settle`, `find_chance`, `find_own` | remains found by survey, by settling, by chance, of its own |
| `met_touch`, `met_heard`, `met_survey`, `met_ship` | meetings by touch, by signal, by survey, by ship |
| `searched`, `sighted` | kyr with the Sight turned outward; kyr holding it |
| `tales`, `witnessed`, `told`, `read`, `inherited` | tales learned, by provenance |
| `forgot`, `myths`, `revised`, `blamed`, `testaments`, `restored` | tales forgotten, worn to myth, rewritten, blamed; testaments written; memories restored from a relic |
| `ticks`, `lean` | ticks lived, ticks shedding |
| `ticks_at`, `lean_at`, `alone_at`, `lean_alone_at` | the same by era, and alone (one system) |
| `sent`, `got` | what was sent and what came, over the life |
| `partners`, `fed` | partners ever, partners ever sent to |
| `built`, `ships_lost`, `rotted` | ships built, lost in battle, rotted laid up |
| `building` | ship-kyr of flow spent building |
| `star_ticks`, `at_want`, `laid_tick` | ticks starfaring, at the want, with ships laid up |
| `battles`, `won`, `empty_sky` | battles fought as the attacker, won on the roll, worlds taken with nothing in the sky |
| `garrisons`, `musters` | garrison moves and musters ordered |
| `sightings`, `intercepts`, `meetings`, `caught`, `pickets`, `salvaged` | fleets seen, interceptors sent, meetings fought, own fleets caught in the dark, pickets sent, ships crewed from fields |
| `fathomed`, `unfathomed`, `brokered`, `misunderstood`, `leaps` | peoples fathomed, fathomings lost to a dark age, brokered attempts, wars that ended unfathomed, remains sealed by looking before the leap |
| `dropped`, `judged`, `acted_gap` | messages dropped unread; councils' verdicts; the sum of the acted-on odds' distance from the mean |
| `hired`, `sold`, `broke`, `bought_off`, `tributes`, `sold_sightings` | contracts bought, sold, broken, sold out of; tributes paid; sightings sold |
| `slights`, `deterred` | slights taken in all; councils the offence alone held back |
| `sickened`, `cured`, `contained`, `worlds_sick`, `cults` | plagues caught, cured, ticks contained, worlds lost, cults formed |
| `refusals`, `shut` | senders closed out, messages dropped for a plague |
| `attempts`, `poisoned`, `detected`, `breakouts`, `leaks` | plagues made and tried: attempts, ones that took, ones seen, ones that got out at discovery, ones that leaked while held |
| `ridden`, `risen` | peoples ridden by this parasite; risings against a rider |
| `oss_faced`, `oss_renewed`, `oss_set`, `oss_broke`, `oss_stiff` | facings of the ossification filter, renaissances, times set, breaks, the stiffness summed at each facing |
| `appeared`, `deepened`, `tithed`, `sleeps`, `wakings`, `demands`, `unmade` | the kinds' doings: appearances, deepenings, tithes, sleeps, wakings, demands, worlds unmade |
| `eaten`, `consumed` | ships grown by eating; worlds stripped and held empty |
| `hunts`, `drifts` | hunts declared on a hole in the ledger; times the shape drifted |

### Age

| field | type | meaning |
|---|---|---|
| `index` | int | |
| `start` | year | the surge |
| `end` | year | fertility below the floor |
| `ender` | key | what swept up the remains, in `portraits.json`'s enders |
| `elders` | [id] | its civilisations |

### Elder

A civilisation of an earlier age: no traits, only a portrait. Finders' names for it are rows of the names table.

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `age` | int | |
| `portrait` | key | in `portraits.json`'s elders |
| `rose` | year | |
| `fell` | year | |
| `legacies` | [id] | what it left |

### Remain

Something an age left on the substrate.

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `age` | int | the age it is from, -1 for the current one |
| `elder` | id | its elder, or -1 |
| `maker` | id | the people that made it, -1 for the elder ages (**D**: `remain_left`) |
| `kind` | key | `artifact`, `structure`, `threat`, `sleeper`, `law`, `bounty`, `field` (**D**: `remain_left`) |
| `star` | id | where (**D**: `remain_left`, `remain_moved`) |
| `node` | key | the art it stands for, for artifacts and structures |
| `portrait` | key | what it looks like: an elder list of `portraits.json` by kind, a work's key for a remain of this age, a form's key for an object |
| `state` | key | `undisturbed`, `sealed`, `wielded`, `mastered`, `unleashed`, `lost` (**D**: `remain_left`, `remain_state`) |
| `people` | id | the people a threat or a sleeper is, or -1 |
| `payload` | key | what a transmitter carries: `corruption`, `seed` |
| `listeners` | int | peoples a transmitter has taken (**D**: `listened`) |
| `woken` | int | parasites a transmitter's seed woke |
| `finder` | id | the people that last acted on it, or -1 (**D**: `remain_left`, `remain_finder`) |
| `level` | key | for a wielded artifact: which level it lifts, or `miracle` |
| `cond` | key | `abandoned`, `derelict`, `wreck`, `ruin` (**D**: `remain_left`, `remain_cond`) |
| `hardy` | number | multiplier on its rate of decay; 0 never decays |
| `source` | id | the source it is, for a bounty or a wielded artifact, or -1 |
| `testament` | [id] | the tales of its testament, in `tellings.jsonl` |
| `plague` | id | the sickness of the mind its makers had when they wrote, or -1 |
| `wrecks`, `derelicts` | int | for a field: the ships in it |
| `at` | [x,y,z] | for a field adrift: where, light years |
| `adrift` | bool | between stars: `star` is the nearer end |

### Source

Everything with a yield: a world, a belt, a giant, a star, a feature's reach, a work, an elder's bounty, an object of a miracle.

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `key` | key | what it is, in `sources.json`; `bounty:<kind>`; `artifact`; a miracle's key for an object |
| `kind` | key | world, belt, giant, star, cosmic, structure, elder, made |
| `star` | id | where, or -1 for a ranged source |
| `feature` | key | for a ranged source: the feature whose reach it is |
| `radius` | number | the reach of a ranged source |
| `yield` | Income | per tick |
| `needs` | [key] | any one of these harnesses it |
| `with` | [key] | and all of these |
| `cradle` | bool | the habitable world of the people that arose on it |
| `rarity` | bool | had or not had |
| `grants` | [key] | nodes that cost half to whoever has it |
| `levels` | [3 numbers] | what it adds to the levels of whoever has it |
| `reach` | number | light years it adds to the reach of whoever has it |
| `mobile` | bool | moves with its holder |
| `holder` | id | the people that holds it, or -1; for an immobile source, the last to harness it |
| `carried` | id | the fleet carrying it, or -1 |
| `legacy` | id | the remain it is, or -1 |
| `since` | year | when it was first put to use |
| `wear` | year | years from `since` until the yield is gone; 0 never wears |
| `form` | key | for an object of a miracle: the form it rolled, in `miracles.json` |
| `sentient` | bool | a food organism that thinks |
| `maker` | id | the people that made it, or -1 |
| `made` | year | |
| `given` | int | cuttings given |
| `fate` | key | for an object: `rose`, `loose`, `doom`, `through`, or none |

### Trace

| field | type | meaning |
|---|---|---|
| `star` | id | |
| `kind` | key | in `conditions.json`'s traces |
| `civ` | id | the people, or -1 |
| `species` | id | the blood that lost the world, or -1 |
| `year` | year | |

### Plague

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `kind` | key | `biological`, `memetic` |
| `contagion`, `lethality` | number | |
| `band` | int | the era band it needs, or -1 |
| `engineered` | bool | shaped on purpose |
| `conscious` | bool | a memetic plague that can wake a mind |
| `profile` | Profile | what it does: `symptoms`, `onset`, `course`, `takes`, `form`, `effects`, `carrier`, keys of `plagues.json` |
| `born` | year | |
| `first_host` | id | the people it was born in |
| `cause` | key | what the birth reads as: clean, dirt, siege, dark_age, relic, born_rider, signal, made, breakout |
| `hosts` | int | peoples that have it now (**D**: the peoples still rising with an `infected` and no `cleared`) |
| `peak` | int | the most at once |
| `caught` | int | peoples that have had it, in all |
| `worlds` | int | worlds lost to it |
| `peoples` | int | peoples ended or brought low by it |
| `cults` | int | peoples that formed around it |
| `cures` | int | |
| `refusals` | int | ears and ports closed for fear of it |
| `woken` | int | times a reservoir or a wall gave it again |
| `last_host` | year | when it last had a host |
| `extinct` | bool | |
| `wildfire` | bool | had ten hosts at once, once |
| `maker` | id | the people that made it, or -1 |
| `made` | bool | shaped on purpose |
| `rider` | id | the parasite people it became, or -1 |
| `transmitter` | id | the transmitter it came down from, or -1 |
| `poisonings` | int | peoples it was put in by stealth |

### Reservoir

| field | type | meaning |
|---|---|---|
| `star` | id | the dead cities |
| `plague` | id | |
| `until` | year | when it is gone |

### War

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `sides` | [id, id] | the declarer and the target (**D**: `war_opened`) |
| `began`, `ended` | year | |
| `over` | bool | (**D**: `war_over`) |
| `cause` | key | in `causes.json`'s war causes (**D**: `war_opened`) |
| `cause_of` | id | the people the cause names, or -1 |
| `named` | id | the star the war is named for, or -1 |
| `nth` | int | the nth war between these two |
| `will` | [2 numbers] | each side's will |
| `taken`, `glassed`, `lost` | [2 ints] | worlds taken, destroyed and lost by each side |
| `result` | key | how it ended, in `causes.json`'s war results (**D**: `war_over`) |
| `pact` | id | the pact it was joined under, or -1 |
| `principal` | id | the ally whose war it is, or -1 |
| `hire` | id | the contract it was declared for, or -1 |
| `hunt` | Hunt | for a hunt: the region fought |

### Hunt

| field | type | meaning |
|---|---|---|
| `x`, `y` | number | the hole's centre |
| `radius` | number | light years |
| `losses` | int | losses inside it |
| `star` | id | the star nearest the centre |
| `since` | year | when the hole was deduced |
| `empty` | year | since when the region has held nothing unaccounted for; 0 while it does |

### Pact

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `members` | [id] | |
| `kind` | key | defence, war, defence and war |
| `target` | id | the people it is against, or -1 for whoever comes |
| `formed`, `ended` | year | |
| `over` | bool | |

### Betrayal

| field | type | meaning |
|---|---|---|
| `by`, `against` | id | |
| `year` | year | |
| `shape` | key | in `causes.json`'s betrayals |
| `weight` | number | negative for faith kept |

### Fleet

An expedition: ships in flight or at a star.

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `owner` | id | |
| `target` | id | the people fought, stood with or looked at, or -1 |
| `kind` | key | campaign, relief, scout, roam, survey, guard, intercept |
| `star` | id | its destination |
| `from` | id | where it set out from |
| `ships` | int | |
| `launched`, `arrive` | year | this leg |
| `base` | id | where it operates from, or -1 in flight |
| `returning` | bool | |
| `over` | bool | |
| `laid_up` | bool | the flow does not keep it |
| `laid`, `manned` | year | when last laid up, when last manned again |
| `held` | [id] | worlds it took and its people still hold |
| `seen` | [id] | the peoples that have seen it |
| `battles`, `wins` | int | |
| `turned` | bool | turned back |
| `out` | year | when it first set out |
| `back` | id | for a campaign fleet that fell back: the world it returns to, or -1 |
| `sieges` | int | times it fell back and came again |
| `drive` | number | years per light year on this leg |
| `warning` | year | the most warning any people had of it |
| `quarry` | id | for an interceptor: the fleet it was sent to meet |
| `picket` | bool | for a scout: it stays and watches |
| `contract` | id | the contract it serves, or -1 |

### Contract

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `buyer`, `seller` | id | the buyer asks and pays, the seller does |
| `by` | id | who proposed it |
| `ask`, `pay` | Term | |
| `length` | number | thousand years the pay runs |
| `offered`, `formed`, `until`, `ended` | year | |
| `state` | key | offered, accepted, running, done, broken, lapsed, refused |
| `broke` | id | who broke it, or -1 |
| `failed` | [2 ints] | ticks running each term has gone undelivered |
| `missed` | bool | the pay has failed at least once |
| `ask_done` | bool | the ask is delivered |
| `taught` | bool | teach: the node arrived |
| `burned` | bool | strike: the work was burned |
| `tribute` | bool | written at a war's end |
| `bought_off` | bool | ended by the seller taking a better offer |

### Term

One side of a bargain.

| field | type | meaning |
|---|---|---|
| `kind` | key | flow, rarity, access, teach, guard, strike, deliver, peace, sighting, broker |
| `res` | key | flow: the commodity's symbol |
| `amount` | number | flow: units per tick; guard, strike, deliver: ships |
| `source` | id | rarity, access: the source |
| `node` | key | teach |
| `star` | id | guard: where to stand; strike, deliver: the target world |
| `work` | key | strike: a work to burn |
| `target` | id | guard: whom against; strike, deliver: whose; peace: with whom; broker: the people to be understood |
| `fleet` | id | sighting: the fleet |

### Fathoming

| field | type | meaning |
|---|---|---|
| `year` | year | |
| `who`, `whom` | id | the one who understands, and whom |
| `how` | key | meeting, kin, chorus, familiarity, war, taught, broker |
| `since` | year | when the two met |
| `diff` | number | how alien the two are |
| `mutual` | bool | this made the pair mutual |
| `reversed` | bool | the other already understood this one |

### Battle

| field | type | meaning |
|---|---|---|
| `year` | year | |
| `star` | id | |
| `attacker`, `defender` | id | |
| `ships` | int | the attacker's |
| `held` | int | ships and guns in the sky |
| `gap` | number | the attacker's levels less the defender's |
| `won` | bool | the attacker won the roll |
| `outcome` | key | taken, withdrew, guns, held, broken, or none for a world taken without a battle |

### Meeting

| field | type | meaning |
|---|---|---|
| `year` | year | |
| `quarry`, `interceptor` | id | the fleets |
| `owner`, `seer` | id | their peoples |
| `ships`, `sent` | int | the quarry's and the interceptor's ships going in |
| `won` | bool | the interceptor won the roll |
| `broken` | bool | the quarry was broken |
| `lost` | [2 ints] | the interceptor's and the quarry's losses |

### Sighting

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `fleet`, `owner`, `seer` | id | the fleet, its people, and who saw it |
| `kind` | key | the fleet's kind |
| `from`, `star` | id | its leg |
| `launched`, `arrive`, `leg` | year | the leg's start and end; `leg` the fleet's launch when seen |
| `year` | year | when it was seen |
| `ships` | int | |
| `mil` | number | the level seen, with noise |
| `speed` | number | years per light year |
| `eye` | key | what saw it: works, fleet, picket |
| `eye_work` | key | the work, for works |
| `eye_star` | id | where the eye was, or -1 for a fleet |
| `feasible` | bool | a meeting point existed |
| `intercept` | id | the interceptor sent, or -1 |
| `offered` | bool | put up for sale, or judged not for sale |

### Name

One name for one object by one culture. A people has at most one row per object and tone. `names.md` gives the rules.

| field | type | meaning |
|---|---|---|
| `object` | Object | what is named |
| `by` | id | the namer: a people's id, or -2 for the human catalogue |
| `name` | string | |
| `mode` | key | `transcribed` (the namer's own sounds), `translated` (a recipe rendered), `designation` (a human catalogue label), `proper` (a human proper name), `adopted` (another's endonym learned) |
| `tone` | key | `self`, `stranger`, `friend`, `enemy`, `monster`, `sky`, `none` |
| `coined` | year | when |
| `from` | id | the culture an adopted name was learned from, or -1 |
| `gloss` | string | the meaning a transcribed name would also be given |
| `recipe` | Recipe | how a translated name was made; absent for a transcription |

### Object

| field | type | meaning |
|---|---|---|
| `kind` | key | `civ`, `star`, `plague`, `war`, `elder`, `makers` (a remain's unknown makers), `source`, `word` (a people's word for the state beneath), `title` (a remnant's ruler), `species` |
| `id` | id | |

### Recipe

| field | type | meaning |
|---|---|---|
| `entry` | key | the epithet entry, in `epithets.json` |
| `pattern` | string | its pattern |
| `slots` | string | the holes as filled, `hole=value; ...` |

## chronicle.jsonl

### Event

One thing that happened, as it happened. The kinds and the parameters each declares are `data/events.json`, with a one-line meaning; a kind with weight is a fact, the kind of thing a people can hold a tale of, and its sort (deed, crime, woe, bond, folly) is the fact's own; a tale carries the teller's.

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `year` | year | |
| `kind` | key | in `events.json` |
| `subject` | id | the people it is about, or -1 |
| `object` | id | the other people, or -1 |
| `star` | id | where, or -1 |
| `legacy` | id | the remain in it, or -1 |
| `plague` | id | the plague in it, or -1 |
| `n` | int | a count: worlds |
| `p` | object | the kind's parameters, by the keys `events.json` declares: ids, numbers, keys, flags, lists; a term as a Term, a morality as a Morality |
| `fact` | bool | the kind is a fact |

## tellings.jsonl

### Tale

One thing a people holds of what happened. A testament is the same record with `remain` set: the maker's tale frozen when the remain was left.

| field | type | meaning |
|---|---|---|
| `id` | id | |
| `civ` | id | the teller |
| `remain` | id | the remain whose testament it is, or -1 for a living memory |
| `fact` | id | the event |
| `learned` | year | when the teller came to know it |
| `source` | key | `witnessed`, `told`, `read`, `inherited` |
| `from` | id | who told it, or whose relic |
| `slant` | int | the teller's regard for the other party when last told: -2 monsters, -1 enemies, 0 strangers, 1 friends |
| `wear` | int | 0 exact, 1 worn, 2 myth |
| `blamed` | id | whom the tale now blames instead, or -1 |
| `revised` | int | times rewritten because a regard changed |
| `forgot` | bool | |
| `believed_year` | year | the year the teller puts on it: exact at wear 0, rounded to a hundred thousand years at wear 1, absent at myth |
| `sort` | key | as this teller judges it through its morality: deed, crime, woe, bond, folly, nothing |
| `weight` | number | as this teller weighs it |
| `rank` | int | its place among the teller's dearest, 1 first; 0 for a forgotten tale. The view tells the first thirty of a living people, the first ten of a dead one, in the order of their years |
| `frozen` | Frozen | for a testament: what the maker's telling read of the world when it wrote |

### Frozen

What a testament's telling read of the world at the moment it was written, of the parties in the tale (its subject, object and blamed) and its star. Everything else a telling reads is the tale's own fields and the names of the maker's time.

| field | type | meaning |
|---|---|---|
| `perceived` | [id] | the parties the maker could hold in mind |
| `met` | [id] | the parties it had met |
| `active` | [id] | the parties still rising then |
| `living` | [id] | the parties still living then |
| `ours` | bool | the star was the maker's own |

## How the view reads it

The legends (`worldgen -legends`, `worldgen -read DIR`) are rendered from these files and nothing else: the chronicle's lines from `events.json`'s kinds and the parameters, the tellings from the tale fields with the teller's knowledge and names, the aftermath from the state. Where the view says a word for a number (a level's band, a stiffness, the wall's stage) the word is in `levels.json`; where it says a name, the name is a row of `names`, chosen by the rule in `names.md`: the namer's own row in the tone the tale's slant asks for, down a ladder of regard, else the human proper name, else the earliest row of any tone, else the designation.
