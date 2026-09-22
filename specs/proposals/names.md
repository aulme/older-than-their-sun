# Names: ids in the simulation, translations and transcriptions after it

**Status:** Implemented (2026-09-22): stage 1 (names-ids) and stage 2 (names-translated) both built; absorption into the spec waits for `specs/plan.md` step 9, with the other open proposals. Adjusts `structured-output.md` (see "Adjustments to structured-output" at the end).
**Last updated:** 2026-09-22

## Problem

Every name in the galaxy is minted by the simulation, once, as a string, and used by everyone after. Four conventions coexist and none of them is the fiction's:

- **One syllabary for every people** (`internal/names/names.go:10-12`): Zaushtsaosheil, Khuukhuyiath, Daimyaokgein, unreadable and indistinguishable. An eldritch thing, a hive, a machine and a planetary mind all get a name of the same shape, though none of them speaks. Heirs share nothing with their line (the Khemohon's heirs are the Lianiatroth and the Vroda). A star is renamed permanently by the first people to arise there (`civ.go:53`) and everyone after, including the player, uses that name.
- **Universal capitalised English** for the miracles (`miracle.go:20`: the Door, the Voice, the Sight, the Unmaking, the Flesh, the Chorus, the Ember, the Manna): the same word in every mouth across sixty million years, vaguely liturgical, and the simulation's voice.
- **Descriptive names minted once and shared globally**: plagues (`names.Plague`), ruler titles (`names.Title`), finder names for elders (`find.go:13`, set on `l.Elder.Name` by the first finder and read by all, against the spec's own "two current civilisations may name the same legacy differently"), war names.
- **Per-people coined words** for the state beneath (`beneath.go:31`): the only convention that matches the fiction.

Underneath: `names.Civ(w.R)` and its siblings draw from the history's RNG (`civ.go:116`, `contact.go:383`, `sunder.go:23`, `expedition.go:432`, `plague.go:533`, `civ.go:715`), so any change to how names are made reshuffles every later roll and rewrites every history. The syllable tables cannot be tuned without changing the wars.

The frame agreed on 2026-09-21: **every name the output carries is a human translation, transcription or approximation**, since the player is human and everything reaches them through translation. Things carry different names for different peoples. Peoples without language coin no names and are known only by what others call them. Things with a plain human term (a star, a black hole, a Dyson swarm) need no alien name at all.

## Scope

**In.**

- The simulation uses **ids only**. No `Name` field on any object in `internal/history`; no call into `internal/names` from the history; the history's RNG never draws for a name.
- A **name record**: `{object, by, name, mode, coined, from}` rows, several per object, one per culture with a relationship to it, plus the human catalogue's row where one exists.
- **Three classes**: proper names (a culture's voice), descriptive names (translations, from recipes), technical terms (plain human words, from the lookups; never named by anyone).
- A **voice per culture**, derived from its profile: none, transcribed, translated, or both; a phonology family with knobs, inherited down a line with drift; a rare unpronounceable family.
- **Grounded translations**: an epithet inventory whose slots are filled from the referent's real properties as the namer knew them, with the tone set by the namer's regard.
- A **names pass** after the run: a pure function of the seed, the namer, the object and the relationship, so a name never changes because another name was added or a table was edited.
- The debug view and `techstats` reading names from the pass; a default-name rule for readability.
- The inventories as data files, with size targets.
- A **plague profile**: a structured description of every plague generated with it (a symptom list, onset, course, whom it takes; or a form and an effect list), keys the names and the flavour read.

**Out.**

- Names changing over time (a renamed seat after a renaissance, a conqueror's name for a taken world replacing the old). One name per (culture, object, tone), coined at the relationship's year; a name outliving the memory of its reason is a feature.
- Exonyms for every pair that met, in every tone. Three per met pair at most (see the entitlement table).
- Human-language variety in the transcriptions. There is one human transcription convention (English-reader, ASCII plus apostrophe and hyphen).
- The narrator's choice of which name the player sees. The record carries all of them and the kind of each; the policy is the engine's. A recommended rule is given.
- Human names for anything but the catalogue. `by: "human"` rows are designations only; what humans call a people they meet is coined in play, by the engine.
- Names for what cannot be perceived. An anti-memetic people gets no row from any culture that cannot hold it in mind; the tellings' "something nameless" is its name there.

## Design

### Ids in the simulation

Every object the fiction can name has an integer id already: peoples (`Civ.ID`), species, stars (`Star.ID`), elders, remains (`Legacy.ID`), plagues, wars, the objects the Ember and the Manna make (`Source`). The history refers to them by id, always. What goes:

| today | after |
|---|---|
| `Civ.Name`, `Civ.HomeName`, `Civ.CradleName`, `Species.Name` | ids; the log lines and the view look names up |
| `Star.Name` overwritten by a people's name (`civ.go:53`) | never overwritten; `Star.Name` is the human catalogue designation or empty |
| `Legacy.Name`, `Elder.Name` (finder names) | gone; a finder's relationship is a fact already (`FFind`) |
| `Civ.Title` | gone; the remnant's status is a fact; the title is a descriptive name from the pass |
| `Plague.Name`, `War.Name` | gone; the pass names them from their facts |
| `Civ.Word` (the word for the state beneath) | gone; the pass coins it from the miracle held |
| `names.Civ(w.R)` and every other draw from the history's RNG | removed; the history's random sequence changes once, for every seed, at this proposal's landing (see "Implementation") |

The sim still prints prose lines until `structured-output` lands. In the meantime `w.log` takes ids where it took names, through typed arguments (`civ(41)`, `star(88)`) rendered as tokens (`{civ:41}`) that the view resolves through the names pass. Everything the sim decides keeps working on ids, which is what it does already.

**Every entitlement is a fact.** The pass reads the chronicle, not the live state, so every relationship it needs must be written as a fact. Two are not today: "met in the flesh" (`Civ.Reached`, set without a fact) becomes `FMet` with a `how` of `touch` or `signal` (the second meeting kind writes a second `FMet`); a sky-name needs no fact, being geometry over the catalogue. Everything else in the entitlement table is a fact already.

### The name record

```json
{"object":{"kind":"civ","id":41},"by":7,"name":"Kesh","mode":"transcribed","tone":"self","coined":2010000,"from":-1}
{"object":{"kind":"civ","id":41},"by":12,"name":"the Salt-Born","mode":"translated","tone":"stranger","coined":2410000,"from":-1}
{"object":{"kind":"civ","id":41},"by":12,"name":"the Burners of Ath","mode":"translated","tone":"enemy","coined":2900000,"from":-1}
{"object":{"kind":"star","id":88},"by":"human","name":"HIP-56601","mode":"designation","tone":"none","coined":0,"from":-1}
{"object":{"kind":"star","id":88},"by":41,"name":"Ur-Kesh","mode":"transcribed","tone":"self","coined":2010000,"from":-1,"gloss":"the Hearth"}
```

- `by` is a culture (a people's id) or `"human"`.
- `mode` is `transcribed` (a rendering of a sound the culture makes), `translated` (a rendering of a meaning), `designation` (a human catalogue label), or `adopted` (learned from another culture, `from` set: a transcription of their word, mangled through the adopter's phonology).
- `tone` is `self` (an endonym or a name for one's own thing), `stranger`, `friend`, `enemy`, `monster`; for stars and remains `self` (the namer's own) or `stranger` (another's, seen).
- `gloss` is present for a transcription whose meaning the culture would also give ("Ur-Kesh, roughly 'the Hearth'"); a translated name needs none.
- The human row's `mode: designation` carries `kind` (`proper`, `bayer`, `flamsteed`, `hip`, `gaia`, `2mass`, `planet`), from `CatStar.Proper()` and the catalogue's source; a proper name (Sirius, Ran, Tupã) is a strong human name and a designation a weak one, which is what the narrator's rule reads.

### Three classes of name

1. **Proper names** get a culture's voice: peoples, star systems, the objects the Ember and the Manna make, and elders as their finders call them. Only these are ever transcribed.
2. **Descriptive names** are translations always, so a recipe from the epithet inventory: exonyms, finder names, plagues, ruler titles, war names, the coined word for the state beneath, endonyms of translated-voice cultures. The pass renders the recipe to a string with the words of `data/banks.json`; the record carries the string and the recipe (`recipe: {pattern, slots}`), so a consumer can re-render or translate.
3. **Technical terms** are plain human words a stranger understands or can look up: star, black hole, neutron star, Dyson swarm, faster-than-light travel, ansible, precognition, directed evolution, memetic engineering, a captive singularity. Nobody names them. They live in the lookups as each key's `term`, and the rule for a `term` is written at the top of the file: a phrase a reader who never saw this project understands at once or can search for, never a capitalised coinage. The Door, the Voice, the Sight, the Unmaking, the Chorus, the Ember and the Manna fail the rule and become terms ("faster-than-light travel", "the ansible", "precognition", "remote matter annihilation", "memetic engineering", "an exotic energy source", "a self-sustaining food organism"); a specific Ember or Manna, once held, is an object with a proper name from its holder, and a cutting carries the giver's name down the trade chain as an adopted name.

### A voice per culture

A culture's **voice** is derived from its species profile, not rolled on its own, so it is a consequence of what the people is:

| the people | voice | why |
|---|---|---|
| biological, conscious, with hearing and sound (the default) | `transcribed`, with a gloss one time in three | it speaks; a human can approximate the sound |
| biological, conscious, `deaf`, or a communication channel other than sound (see below) | `translated` | nothing to transcribe; the meaning is what reaches us |
| machine | `translated` (functional: "Third Foundry", "the Maintenance") unless made by or uplifted from a speaking people, then the maker's voice with drift | a made people inherits a tongue or has none |
| parasite riding a host | the host's voice, adopted | it speaks through the host |
| hive, unconscious, planetary, eldritch, replicator | `none`: coins no names; known only by exonyms | no language, or none we could render |
| anti-memetic | `none` for anyone who cannot perceive it; its own names exist only for the resilient | nothing unperceived is named |

A **communication channel** is added to the species: `sound` (default), `light`, `electromagnetic`, `chemical`, `touch`, `none`; rolled at generation from substrate and senses (an eyeless people that sees by sound is acoustic; a people that hears light or senses the living current tilts to light or electromagnetic; a chemical-sensing people to chemical; a sessile one to touch), stored on the species as a key, shown in the portrait through the lookup. The channel decides the voice above and is a fact the player can find out.

A transcribed voice carries a **phonology**: a family and a few knobs (onset, nucleus and coda subsets of the family's inventory; a syllable range; a diphthong cap; whether apostrophes or hyphens appear). It is **per people**: a first people of a species rolls it from a hash of the seed and the species id; **heirs, branches and shards inherit** their parent people's family and knobs with a small drift (one subset changed, the range shifted by one), so a line sounds like a line and a schism like a dialect; an uplifted or bred people takes its maker's phonology with a larger drift; a family is light-biased by the body (`families.json` carries the bias: a sessile people toward `glottal`, a swarming one toward `reduplicating`), so the sound hints at the body without deciding it.

### The names pass

`internal/names` is rewritten as a pure function over the record: `Name(seed, by, object, relationship) → row`. It runs after the simulation (or lazily, whenever the view asks, with the same result), walks the relationships, and emits the rows. Every random choice inside it is drawn from a small generator (`rand.NewPCG`) seeded by an FNV-64 hash of the canonical string `seed|by|object.kind|object.id|tone`, so:

- a name never changes because another name was added, a relationship was removed, or the history's rolls moved;
- a table edit changes only the names that read the edited entry;
- the same function serves the view during the run and the writer after it.

The pass reads only the record: species and traits, facts, the catalogue. The facts it reads for grounding are those **before the coining year**, so a name coined at first contact knows nothing of the war that came later.

### Who names what: entitlement

A culture names an object when it has a relationship to it, and only then. The relationships, from facts the sim already writes:

| object | relationship (fact) | tone | coined at |
|---|---|---|---|
| itself | arising (`FArise`) | `self` | its birth |
| another people | first knowing of them (`FMet`, or a first sighting) | `stranger` at contact; a friend is called by its endonym, adopted | the meeting |
| a voiceless people | the first bond with it (`FPact`, `FTrade`, `FRelief`) | `friend`: there is no endonym to adopt | that fact |
| another people | the first grudge, war, crime remembered or monster reckoning against them | `enemy`; `monster` if the monster reckoning holds at coining | that fact |
| a star | its cradle (`FArise`), a settling (`FSettle`), a taking (`FTaken`), a finding there (`FFind`), a war named for it, a survey lost there | `self` for its own worlds; `stranger` for a star seen or fought over | the fact |
| a star | seen from home before it was reached: within telescope range of a holding, a real star of magnitude bright enough from there | `sky` | the first tick it was in range |
| an elder | a finding of its work (`FFind`, `FMastered`, `FSealed`, `FUnleashed`) | `stranger` | the find |
| a remain of this age | a finding, when the maker is not known to the finder | `stranger`, a finder name for the makers | the find |
| a plague | catching it (`FPlague`); or holding a tale of another's catching it (the heavy facts travel by word) | `self` for the sufferer, `stranger` for one who only heard of it | the catching; the learning of the tale |
| a war | being a side | `self` per side | the declaration |
| an object (an Ember or a Manna) | holding it; receiving a cutting | `self`; `adopted` for the cutting | the holding |
| the state beneath | a causal miracle held (`FMiracle` for `ftl`, `ansible`, `foresight`, `unmaking`; the wound) | `self`, the coined word | the reaching in |

Bounds: nothing else is named. A star a culture never reached, saw or fought over has no name from it; a people it never met has none of its own (a people known only by tale is called by the teller's row, adopted with the teller's slant, as the tellings already pass regard). At most three rows per pair of peoples (`stranger`, `enemy` or `monster`, and `friend` for the voiceless). Rows per run are in the low thousands.

**Adopted names.** At a first meeting a people learns the other's endonym (if it has one) as an `adopted` row, transcribed through its own phonology: the source name is re-syllabified into the adopter's family inventory with a per-family substitution table for the sounds it lacks (a cluster-heavy people mangles an open-syllable name: Kailoa → Kaloth; a translated name is copied unchanged, being already in our words); for a `none`-voiced people there is nothing to adopt and the exonym stands alone. A people reading a wall or a relic (`FFind`, a testament read) adopts the makers' names for the star and for themselves, if the makers had a voice: this is how a ruin keeps a name three cultures after its builders. Adoption is bounded to these two events.

### Transcription: families

A **family** is a sound pattern recognisable from one name. Each has an inventory (onsets, nuclei, codas), a syllable range, cluster rules and the human-transcription devices it uses. A starting set, with signatures; the inventories live in `data/families.json`:

| family | signature | examples | weight |
|---|---|---|---|
| open | CV syllables, no codas, long | Kailoa, Manu, Tehoa | 20 |
| cluster | consonant clusters, hard codas, short | Krv-thak, Drosk, Vrenth | 15 |
| glottal | one or two syllables, apostrophes | Tsu'o, K'ai, Qo'ra | 10 |
| agglutinative | three to five short syllables, hyphens | Ulan-teru-mek, Iri-vata | 10 |
| sibilant | s, sh, z, th throughout | Ssesh, Sirith, Zhassa | 10 |
| nasal | m, n, ng, doubled nasals | Ngoma, Nnun, Emmen | 10 |
| reduplicating | a syllable doubled | Tiki-tiki, Rorro, Kalka | 5 |
| vowel-rich | vowels outnumber consonants, pronounceable and strange | Aeioa, Ouei, Iaun | 5 |
| liquid | l, r, w, y and long vowels | Lirala, Waru, Yelu | 10 |
| unpronounceable | no vowels, or clicks and stops, or a repetition with a tempo; a human cannot say it | Xkrrth'qv, !Kth, K-k-k-th | 3 |

The unpronounceable family is **motivated**: it is drawn only for a people whose bio or sense traits put its speech far from ours (`eyeless`, `sessile`, `manysexes`, `swarming`, `chemical`), and never for more than one family draw in thirty. Its names are the case the human designation is for ("HIP-56601, whose name in their tongue no human has said").

Every transcription passes a **convention** check: ASCII letters plus apostrophe and hyphen; at most one diphthong per name; no doubled vowels; a coda cap; a length cap of twelve letters. The unpronounceable family is exempt from the diphthong and vowel rules and subject to the length cap. A test generates a thousand names per family and requires a small letter-class classifier (vowel ratio, cluster length, apostrophes, hyphens, doubled letters, length) to assign them to their family well above chance; if two families are confused, they are not distinct enough.

### Translation: grounded recipes

A translated name is a **recipe**: a pattern from `data/epithets.json` with its slots filled. The slots are filled from **properties of the referent**, not from a free bank, and only from properties **the namer could know** at the coining year:

| slot source | what it draws on | known to the namer when |
|---|---|---|
| body | bio traits (`robust`, `fragile`, `sessile`, `manysexes`, `swarming`, `longlived`, `shortlived`, `eusocial`, `symbiosis`, `amphibious`), senses (`eyeless`, `thermal`, `electric`...), substrate | met in the flesh (`Reached`) or fathomed |
| world | the referent's cradle archetype (`ocean`, `arid`, `iceshell`, `floater`, `volcanic`, `twilight`...), and the world traits (`skyless`, `fireless`, `threesuns`) | the star charted, or met in the flesh |
| way | organisation and stance (`herd`, `caste`, `collective`, `conqueror`, `pacifist`, `nomadic`, `submissive`...), morality (the fixation: conquest, spawning, knowing, the old things, holding) | fathomed, or a war fought |
| deed | a fact between the two before the coining year (a world burned, a pact broken, relief sent, a plague given, a yielding), the heaviest first | witnessed by both, so always |
| voice | the channel and the voice (a `light` people is "the Flickering", an `electromagnetic` one "the Hum") | met by signal |
| kind | substrate and modifiers as seen (machine, eater, a world that woke, a thing with no one home) | met at all |
| sky | for stars: the real class and colour, brightness from the namer's home, whether it is a pair or a triple | seen |
| place | for stars: what the namer did there (settled, fought, lost surveyors, found a work) | the fact |
| work | for elders and remains: the node of the work (star lifting, a swarm, a transmitter, a vault) and the legacy's kind | the find |

**Tone** selects the phrasing: the same body trait `robust` yields "the Steady" from a friend, "the Heavy Ones" from a stranger, "the Brutes" from an enemy, "the Crushers" from one that counts them a monster. An entry in the inventory is a pattern with a `tone` set, a `requires` list of property predicates, and a `notice` bias (which cultures tend to pick it: an eyeless people never names by colour; a skyless one names by touch and sound; a hive by number; a caste society by rank). A culture's `notice` weights are derived, not rolled: a table in the header of `epithets.json` maps its senses, world traits and organisation to weights per semantic field (`eyeless` → colour 0, sound 3; `skyless` → sky 0, touch 2; `caste` → rank 2), and an entry's `notice` names the fields it uses. The pass scores the entries whose `requires` hold against the namer's `notice`, picks by weight from the top few, and fills the slots. When nothing grounded is available (a people met only by signal, nothing yet between them) it falls back to relationship-only entries ("the Voices from the Dark", "the Ones Beyond Ath").

**Endonyms** of translated-voice cultures and the glosses of transcribed ones come from the same inventory with `tone: self`: what a culture values (the morality: "the Rightful" for a conqueror fixation, "the Many" for herd, "the Free" for individualists, "the Keepers" for the old things) and where it comes from (the cradle: "the Salt-Born", "the People of the Deep Water", "the Sky-Born"; a `skyless` people never "the Sky-Born").

**Stars** by relationship: hearth words for the cradle ("the Hearth", "the Cradle", "Home-fire"), sky words for a star seen before it was reached, from its real colour and brightness ("the Red Eye" for a red dwarf, "the Twins" for a bright binary, never "the White Eye" for a red star), settlement words for a colony ("New-Kesh", "the Second Hearth", "Far-Landing"), war words for a contested star ("the Wound", "the Field of Ath"), dread words for a star surveyors did not return from ("the Silent Star", "the Mouth").

**Elders and remains** from the work: a star-lifter's makers are "the Ones Who Moved the Star", a swarm's "the Ones Who Wrapped the Sun", a transmitter's "the Ones Who Would Not Stop Talking", a vault's "the Careful Dead", a law's "the Ones Who Changed the Rules"; the current fixed lists (`find.go:13`, `legacy.go:49`) become entries with `requires` on the node.

**Plagues** from their **profile**. Every plague gets a structured description when it is born (`plague.Plague.Profile`), keys only, so the names and the flavour have something real to read:

| kind | keys | drawn from |
|---|---|---|
| biological | `symptoms`: a list of one to four, in the order they come (fever, cough, sweat, pox, bleeding, blindness, madness, sleep, wasting, rot, bloom, fade; and the visible marks among them: grey skin, weeping eyes, glass growths, black tongue), `onset` (quick, slow), `course` (days, years, generations), `takes` (the young, the old, everyone, one caste, one species) | lethality picks the banks (over 0.6 draws the last symptom from the deadly ones, under 0.3 from the wasting ones) and the list's length; contagion picks the onset (quick above 0.5); `Engineered` keeps the visible marks out and `Band` sets `takes` to one species; the rest by hash of the plague id |
| memetic | `form` (question, doctrine, laugh, silence, number, song, dream, argument, certainty, prayer), `effects`: a list of one to three, in the order they come (what believers do: stop working, stop speaking, build, burn, leave, count, kneel, refuse the cure, spread it), `carrier` (words, images, music, a proof) | `Conscious` draws from the forms that want a host (question, argument, prayer); lethality picks the last effect's severity and the list's length; the rest by hash |

The profile is drawn by hash of the seed and the plague id, not from the history's RNG, so it costs no rolls; it is a property of the thing and lives on the record. `data/plagues.json` carries the flavour per key, attached after the fact by the interpretation layer: what a symptom looks like, what an effect does to a city. Names read the profile: the sufferer names by the symptom it noticed (usually the first or the most visible: "the Grey Rot", "the Glass Sleep") or the form ("the Silent Question"), a neighbour by the host ("the Kesh Sweat", "the Fever of Wolf 359"), a cult by the form's dignity ("the Certainty"). The flavour of a plague is the whole list read in order: a fever, then the sweats, then the sleep nobody wakes from.

**Titles** from the remnant's fate and morality: the fate's cause (a lost war, a plague, a dark age) and the fixation pick the adjective and the noun ("the Last Warden" for a holder that lost everything, "the Sleepless Archon" for a knower). **Wars** per side, from the cause and the contested star ("the War of Ath" for one side, "the Betrayal" for the other). **The word for the state beneath** from the miracle held, as now, with the entries moved into the inventory (`ftl` holders lean to place words, "the Between", "the Elsewhere"; `foresight` holders to time words, "the Long Now", "the Interval"; `unmaking` holders to absence words, "the Blank", "the Nothing").

### The epithet inventory

`data/epithets.json` and `data/banks.json` are data, grow without a code change, and are the main lever on how the galaxy reads. Structure:

```json
{"id":"heavy_stranger","about":"civ","tone":["stranger"],"requires":["trait:robust"],
 "pattern":"the {Adj} Ones","slots":{"Adj":"bank:heavy"},"notice":{"sense:masssense":2,"world:lowg":2},"weight":1}
{"id":"burners","about":"civ","tone":["enemy","monster"],"requires":["deed:FBurned"],
 "pattern":"the Burners of {Star}","slots":{"Star":"deed.star"},"weight":3}
{"id":"red_eye","about":"star","tone":["sky"],"requires":["class:M","seen"],
 "pattern":"the {Red} Eye","slots":{"Red":"bank:red"},"weight":1}
```

`banks.json` holds the word lists by semantic field (`heavy`, `red`, `salt`, `deep`, `quiet`, `fire`, `teeth`, `sleep`...), each with a dozen or more words, so the same entry does not render the same twice. Size targets at landing, with a test that counts:

| about | entries | of which grounded (a `requires` on a property) |
|---|---|---|
| civ, `self` | 60 | 50 |
| civ, `self`, machine functional names | 20 | 20 |
| civ, `stranger`/`friend` | 120 | 100 |
| civ, `enemy`/`monster` | 120 | 100 |
| star, by relationship | 80 | 80 |
| elder and remain, by work | 40 | 40 |
| plague, by profile | 40 (plus banks) | 40 |
| title | 30 | 20 |
| war | 20 | 20 |
| the state beneath | 30 | 30 |
| **total** | **560** | |

and every property in the slot-source table above must be required by at least eight entries across the tones, so no trait is nameable in one way only. A second test renders a thousand exonyms from a reference batch and requires no entry to account for more than three percent of them.

### What the player sees: a recommended rule

The record carries every row; the engine chooses. The rule the debug view uses, and the one recommended to the narrator:

- **A people** is called by the name the player learned first: the endonym if the player met them or a friend of theirs, an exonym if the player met an enemy first, with the other revealed later ("the ones the Vroda call the Hollow Ones call themselves Kesh").
- **A star** with a strong human name (a proper name in the catalogue) leads with it and follows with the alien name ("Sirius, which the Huutriatrok call…"); one with only a designation leads with the alien name and trails the designation ("Ur-Kesh, HIP-56601 on human maps"); one with neither, and no culture's name, is described ("a red dwarf eleven light years from Ath").
- **A technical term** is used as a term; if the fiction wants the alien word, it is "a word that translates as *Dyson swarm*".
- In a **telling**, the teller's own names are used, chosen by the tale's `slant`: `self` and `stranger` rows for friends and strangers who have been heard, `enemy` and `monster` rows below zero, the myth-wear archetypes and "a star whose name is lost" as the tellings grammar says.

The **debug view's default** for a name, when the view must print one line per thing: the human proper name; else the earliest `self` row; else the earliest row of any tone; else the designation; else the id. Seeds stay readable for tuning.

### Tests

- The history package imports nothing from `internal/names` and holds no `Name` field (a source test).
- Determinism of the pass: the same record gives the same rows; adding a relationship changes no other row.
- Every transcription passes the convention; the family classifier separates the families.
- Every voice-`none` culture has no `self` row and every met one has an exonym.
- Every entry's `requires` names real property keys; the size and grounding targets; the three-percent cap; the eight-entries-per-property floor.
- Every row's slots were filled from properties the namer could know at the coining year (a test that recomputes the entitlement from the facts).
- Every plague has a profile whose keys exist in `plagues.json`; a deadly plague's last symptom is from the deadly bank; an engineered one has no visible mark.

## Stages

Two stages on one design. The record, the classes, the voice and the pass are stage 1's; the grounded translations, the inventory and the plague profile are stage 2's. Stage 2 adds rows to a table stage 1 defines and changes nothing stage 1 built.

**Stage 1: names-ids.** Ids in the sim; the record; transcribed proper names. **Done (2026-09-21).**
- *Delivers*: no `Name` field in `internal/history` and no draw for a name from its RNG; `FMet` with `how`; typed log tokens; the communication channel and the voice on the species; the phonology families in `data/families.json` behind the first `embed.FS`; the pass for `self` rows (endonyms of transcribed-voice peoples, cradle and settled stars, held objects) and `designation` rows; adopted endonyms at meetings; the view and `techstats` reading names through the pass with the default rule. Translated names are stubbed as the old descriptive generators on the hash stream (a plague is still "the Grey Rot", a title "the Last Warden", an exonym the endonym) so the legends read as before.
- *Testable*: the source test (no names in history); the pass's determinism (adding a relationship changes no other row); the convention and the family classifier; every voiceless people has no `self` row; the reference batch regenerated and `TestDeterminism` green on it.
- *Depends on*: nothing. Shifts every seed once; done first and alone.
- *As built*: `internal/names` (`hash.go`, `families.go`, `book.go`, `stubs.go`), `data/families.json` behind `data.FS`. Tokens are `{kind:id}`, `{^kind:id}` (first letter raised) and `{kind:id@by}` (one people's name for a thing; the tellings use it); the resolver runs to a fixed point so a name may hold a token. Kinds: `civ`, `star`, `plague`, `war`, `elder`, `makers` (a finder's name for a remain's unknown makers), `species` (the first people of a blood), `title`, `word`, `source`. The "how" of `FMet` is its `What` (`touch`, `signal`, or `noticed` for a one-sided finding the other never knew of); `Fact` gained `Plague`; `FWord` is the reaching-in the word for the state beneath is coined from; `War.Named` is the star a war is named for. `Intel.Sick` is a plague id. Rows the translated pass will replace carry `stub: true`; the stubs are the old generators on the hash stream, and a translated-voice people's endonym is the syllabary until then. A voiceless people nobody voiced has met prints as `unnamed #id`. The mangling of an adopted name is by the family's `subs` table with a glide breaking a long vowel run; the classifier separates the ten families at 75–99% (glottal 62%) against 10% chance. `TestSameHistory` re-pinned; the reference batch regenerated.

**Stage 2: names-translated.** Grounded translations, the inventory, the plague profile. **Done (2026-09-22).**
- *Delivers*: the plague profile at generation with `data/plagues.json`; `data/epithets.json` and `data/banks.json` to the size targets, with the `notice` header table; the pass for translated rows: exonyms in their tones, endonyms and glosses of translated-voice peoples, machine functional names, stars by relationship and sky, elders and remains by work, plagues by profile, titles, wars per side, the words for the state beneath; names read from walls; the technical-term rule applied to every lookup `term` (the miracle and object names replaced).
- *Testable*: the size, grounding, eight-per-property and three-percent tests; the entitlement recomputation; the profile tests; the batch's legends read through the new names with no change in the history's numbers.
- *Depends on*: stage 1; `structured-output` stage 2 (lookups), which owns the `term` fields and the lookup coverage test the rule is enforced through.
- *As built* (2026-09-22): `internal/plague/profile.go` draws the profile (`plague.Plague.Profile`) at every birth by an FNV hash of the seed and the plague id into a PCG, and `data/plagues.json` carries the symptom, form, effect, onset, course, takes and carrier rows. `data/epithets.json` holds 1,561 entries (the size table's 560 is a floor: elders and makers have entries per legacy kind and per portrait, enemies per deed and per war cause) with the `notice` header table and the `properties` list the floor test counts; `data/banks.json` the word lists, a dozen or more each. `internal/names/translated.go` is the pass: what a namer knows is recomputed from the facts between the two before the coining (kind and voice at any meeting; body and world at a meeting in the flesh or an understanding; ways at an understanding or a war; every deed, the heaviest first for the slots); an entry's requirements must hold, its score is its weight times the namer's notice weights, and the draw is by score from the top six with ties kept, so the file's order never decides. The row carries the recipe (`entry`, `pattern`, `slots` as filled). Rows made: endonyms of translated voices and glosses (one in three) of transcribed ones; machine functional names; `stranger` at a meeting, `friend` for a voiceless people at the first bond (a voiced people's friend row is its adopted endonym, coined at the meeting), `enemy` at the first war or crime and `monster` when the crime weighs four or more; stars as cradle and colony (translated voices), taken, lost, burned, fought over, silent, and the sky names, the twelve brightest naked-eye real stars from the cradle (magnitude from the catalogue and the distance), coined at the people's birth; elders and unknown makers by kind, work and portrait; names read from walls (`find` with `known`: the makers' own rows for themselves and the star, adopted); plagues by the sufferer at the catching and by a hearer that holds a tale of another's catching; titles at the fall; wars per side; the words for the state beneath; the objects by form, a cutting adopting the giver's name. The tellings name a party in the tale's regard through `{civ:id@by:tone}` tokens, with the ladder monster → enemy → stranger → friend → self. The technical terms replaced the miracle names in `tech.json`, `miracles.json` (`name` became `term`), `filters.json`, the lines and the tellings; a kindled or grown object is named in its line. Tests: the sizes and grounding, the eight-per-property floor, the three-percent cap over four seeds, the well-formedness of the inventory against the other tables, the entitlement recomputation, the profile's rules, every voiceless met people named, no token unresolved. The history's digest is unchanged; the legends' digest re-pinned.

## Implementation / migration notes

1. **Ids in the sim.** Remove every `Name` field and every `names.*(w.R)` draw from `internal/history`; `FMet` with `how`; typed log arguments rendered as tokens; the view resolves tokens through the pass, which at this step is the old syllabary on the hash stream (no attempt to reproduce the old names). **This changes every seed once**: the history's RNG no longer draws for names, so every later roll shifts. The reference batch (`reports/tech`) is regenerated at this step and the byte-identity guard restarts from the new batch. Do it first and alone.
2. **The channel and the voice** on the species; the phonology families and the convention; the pass for transcribed proper names (endonyms, cradle stars). Classifier test.
3. **The plague profile** at generation, with `plagues.json`. **The inventory and the banks** to the size targets; the pass for translated names: exonyms in their tones, stars by relationship, elders and remains, plagues from the profile, titles, wars, the state beneath. Grounding tests.
4. **Adoption**: endonyms learned at meeting, names read from walls.
5. **The view and techstats** read names only from the pass, with the default rule; the technical terms replace the miracle and object names in the lookups.
6. Absorb into the spec; `structured-output` implements the record as adjusted below.

## Adjustments to structured-output

This proposal lands first and changes `specs/proposals/structured-output.md` as follows; those edits are made there:

- **Names are not fields, they are rows.** `state.json` gains a `names[]` table (the record above); `civs[]`, `stars[]`, `elders[]`, `remains[]`, `plagues[]`, `wars[]` carry no `name`. The exception is `stars[].designation` (the human catalogue label and its kind), which is the substrate's, not a culture's. The scope line "Names are data the sim generates" becomes "names are generated after the run by the names pass from the record; the sim holds ids".
- **`words.json` splits** into `families.json` (phonologies), `epithets.json` (recipes) and `banks.json` (word lists); the miracle, object, archetype, title and beneath-word tables the proposal listed under `words.json` move there, and every lookup key's `term` obeys the technical-term rule.
- **Testaments need no frozen `names` map.** The maker's names are its rows, coined at their years, and a testament references the maker's `by` rows; wear-based name loss stays a rule of the tellings grammar. The sharpening item "Names at wear" is amended accordingly.
- **The fold test's durable table** does not include names: they are derived from the record by the pass, and the pass's own determinism test covers them.
- **The byte-identical guard** through the structured-output stages holds against the reference batch regenerated at this proposal's step 1, not the current one.
- **Stage 1 of structured-output** (events replace log lines) starts from log lines that already take ids as tokens, which makes the conversion smaller: the arguments are already typed.
- **`codex/tellings.md`** gains the name-choice rule (which row a teller uses, by slant), and `FORMAT.md` documents the `names[]` rows and the human designation kinds.

## Open questions (sharpening)

- [x] **Plague.** A profile at generation (a symptom list in order, onset, course, takes; a form, an effect list, carrier), keys derived from the numbers where they can be and hashed otherwise; flavour per key in `plagues.json`; names read it. → "Plagues".
- [x] **The transition stub.** Step 1 uses the old syllabary on the hash stream and does not try to reproduce the old names; the RNG shift changes every history anyway.
- [x] **Every entitlement a fact.** `FMet` gains `how`; sky-names are geometry; the rest already are facts. → "Ids in the simulation".
- [x] **Phonology per people**, seeded from the species for a first people and from the parent with drift for heirs; a light body bias on the family. → "A voice per culture".
- [x] **`friend` tone** coined only for voiceless peoples, at the first bond; at most three rows per pair. → entitlement table.
- [x] **The unperceived** get no row from those who cannot perceive them. → scope.
- [x] **Humans** are designations only; human names for peoples are the engine's, in play. → scope.
- [x] **`notice` weights** derived from senses, world traits and organisation by a header table. → "Translation".
- [x] **The hash**: FNV-64 over a canonical string into a PCG. → "The names pass".
- [x] **Adopted-name mangling**: re-syllabification with a per-family substitution table. → "Adopted names".
- [x] **Machine functional self-names** in the size table (20).
- [x] **Third-party plague names**: entitled by holding a tale of the catching, not by signal range; peoples known only by tale take the teller's row. → entitlement table.
- [x] **Gloss as a field** on the transcribed row; a consumer may promote it.
- [x] **Sky-name threshold**: naked-eye from the namer's cradle (magnitude 6 at that distance), since sky-names are the old names.
- [x] **Size.** Split into two stages (see "Stages"): **names-ids** and **names-translated**. Absorb after the second.

## Deferred

- Names that change over time (a renamed seat, a conqueror's name replacing the old).
- A second human transcription convention (a non-English reader).
- Human names for peoples coined in play (the engine's).
