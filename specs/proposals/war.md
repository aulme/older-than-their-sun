# War: how peoples fight, and how a war ends

**Status:** Draft (2026-09-23). Placed in `specs/plan.md` as steps 10 and 11, after `leaders.md` (a leader at the front is one of the things a war council reads) and before `empires.md`, whose conquest wave and coalitions ride on the machinery here and whose tuning pass is read against the wars this leaves. Read by `art.md` (a war is most of the circumstances worth making anything about) and `miracles.md`.
**Last updated:** 2026-09-23

## Problem

A war today is a declaration and at most one fleet. Seed 5 of the step-9 reference batch (400 stars, 363 peoples over a 38.5 Myr age) had 398 wars, 415 battles and 228 campaign fleets, against 7207 scouts and 3944 guards. Read war by war:

- **The declarer sends one campaign, and never a second.** The council considers only peoples it is not already at war with (`council.go`, the `c.Wars[eid]` skip), so once a war exists nothing sends another fleet to it. 221 wars had one campaign from the declarer; 177 had none; none had two.
- **The side declared on never strikes back.** It is at war with the declarer, so it is skipped for the same reason: 2 counter-campaigns in 398 wars.
- **Over a third of wars never see a ship by construction** (141 of 398). A war an heir inherits at a sundering (71), one a disturbed sleeper brings (34), one an ally joins by pact (22), a poisoning or an unmaking (14) is declared without any council sizing a fleet for it.
- **The one fleet is small and short-lived.** Median two ships, most fighting one or two battles: it takes its target or falls back a hop, moves to the next enemy world within 20 light years, and goes home when there is none.
- **Will runs on a clock.** Each side loses 0.05 a thousand years, nothing while a fleet is in flight, half while one is out, twice when nobody's is — so a war nobody sends ships to ends in about ten ticks, and one with a fleet out lasts until the fleet comes home. The median war is 16 ticks long and holds one or two battles.

**And the ticks between are staring.** Travel is not what takes the time: a campaign's flight is a median of a hundred years, inside a tick. A war with a campaign runs a median of 9 ticks (quartiles 3 and 21): about 3 from the declaration to the launch (the muster declares at once, then gathers guards from across the realm and waits for the docks to build to the want), the fighting inside one tick (first battle to last, median zero), then a median of one tick, and a quarter nineteen or more, of nothing until someone's will runs out. A war with no campaign runs a median of 19 ticks, a quarter 76 or more, and an unyielding people's will does not drain at all.

That is why battles are rare, why wars take so little, and why step 9's leaders at the front so seldom meet a battle. It is not a tuning problem: the procedure has no step for a war that is already running.

The galaxy should instead show the patterns real history shows, and they should **emerge from the rules** rather than be scripted:

1. **Short wars and long ones.** Many wars over in a tick or three; many running a dozen ticks or more, with fighting in them throughout.
2. **Old enemies.** The same two peoples at war again and again across long spans, each war with its own occasion: England and France, Rome and Carthage.
3. **World wars.** Alliances triggering one another until most of a region is at war.
4. **Border disputes.** Frequent short wars between large realms over a world or two, settled without either staking its existence.
5. **Cold wars.** Two powers building fleets against each other for long spans, with only incidents between them.
6. **Proxy wars.** The clients of great powers fighting each other while the powers stay at peace.
7. **Conquest waves.** A realm that keeps winning and keeps going — Alexander, the Mongols — until it is stopped, spent, or its leader dies.

Today's material for them is there and idle. The cascade exists (a declaration calls the target's defensive allies and brings in the declarer's allies of war), and seed 5's largest system of overlapping wars already takes in 13 peoples, but the wars in it are declared without fleets. Grudges, claims and the count of wars fought exist, and 46 of seed 5's 264 warring pairs fought twice or more, but every war between them is as empty as the first. There are 53 pacts, one of 17 members, and 34 vassals standing at the present, but a held people holds no council, cannot be declared on and is never in a pact (`council.go`, `pact.go`). Shipbuilding reads fear, garrisons, a campaign's want and explorers, and no rival (`ships.go`, `mind.WantShips`).

## Scope

**In.** The council on a running war: press, hold or sue, for each side of each war, from what it believes. Will that follows the war. A war's aim, and the terms it ends on. Allies fighting the wars they join. Old enemies. The arms race and the incident. Clients with a war council of their own. Momentum: the machinery a conquest wave runs on.

**Out.** What makes a conquest wave begin (the marginal periphery, the unification, the hollow target) and coalitions against a hegemon (`empires.md`); war spending what it is fought over (`empires.md`); the battle arithmetic (`internal/battle`, unchanged); deliberately hunting a leader (`empires.md`, per `leaders.md`).

## Design

The rules this keeps from the rest of the plan: **the mind decides** (every choice below is a pure function in `internal/mind` reading numbers the history gathers, tuned in `mind.Tuning`); **decisions read beliefs, not the world** (what a people thinks of its enemy is its `Intel`, aged and noisy, and the appraisal it already makes before a first strike); and **a leader moves the disposition, never the procedure** (`leaders.md`).

### A war has an aim

Every war is declared for something, and the aim decides what winning is, when pressing stops and what terms are asked. The aim follows from the cause and the declarer's posture — a column on `causes.json`'s war causes:

- **A world** — the contested world or two at the front: a border, a claim of an heir, a sleeper's star, an embargo's retaliation. A limited war.
- **Tribute** — the greedy and the opportunist: the enemy's spare, paid for a term. Limited.
- **Submission** — conquest: the enemy as a vassal or a slave. Total in practice.
- **Redress** — revenge and the old quarrel: what was lost taken back, or its price in worlds. Limited, and escalating with each war (below).
- **Ending** — extermination, a monster, the hate: nothing short of the enemy gone. Total.
- **Defence** — an ally called in: the principal's war ended well. Bound to the principal's war.

The side declared on has its own aim, which is at least **to hold**, and may become **to retake** or to carry the war home if its council presses.

Limited aims make short wars; total ones make long ones. That is pattern 1 and pattern 4 without a rule for either.

**What a war is at this scale.** A tick is a thousand years, and a campaign of the kind history books record is over inside one. So a war of a dozen ticks is not one campaign drawn out; it is a string of campaigning seasons, a tick each, with the war councils deciding between them whether to go again. A long war is long because both sides keep choosing to fight it, and the gate reads battles per tick of war, not length alone, so a long war cannot pass by staring.

### The war council

Each side of each running war sits a war council on the ordinary cadence, and whenever an event of that war summons it: a battle lost, a world taken or lost, a fleet broken, a leader fallen, an offer of terms. It reads:

- **The balance as believed**: the appraisal the council already makes before a first strike, against the enemy's next world by the aim, from its own `Intel` — so it can be wrong, and a battle, which is an observation (`observe` on both sides), corrects it.
- **The aim's progress**: the worlds the aim names, taken or not; the home in reach or not.
- **Its will, its fear, its posture and its leader.**
- **Its other wars**: ships committed elsewhere are not available here. A war on two fronts is weaker on each.

And it returns one of three:

- **Press** — size and send a campaign at the next target the aim names (the contested world for a border, the home for conquest, the nearest otherwise), as the first strike's sizing does today, from the guards or a muster. The side declared on presses too: to retake what it lost, or to carry the war to the declarer.
- **Hold** — keep the ships home: guards at the worlds the war threatens and relief where the enemy is (the garrison policy already reads fear; it will read the war's front), no new campaign; the campaigns out go on.
- **Sue** — offer terms (below).

**Both sides can believe they are winning.** Two stale or noisy pictures, or a wise people's appraisal against a foolish one's, and both press; the war runs hot until the battles have taught one of them otherwise. A wiser people's appraisal already has a thinner optimistic tail (`Wisdom.Tail`), so the wise make that mistake less often, which is what wisdom is for.

The skip in the ordinary council stays: it is the question "whom should we start a war with", and the war council is the question "what do we do about the one we are in". Every war gets the second, however it began — the heir's inherited war, the sleeper's, the ally's, the hunt's — so none is fleetless by construction.

### Will follows the war

Will starts where the posture puts it, as now, and then **moves with what happens** rather than on a clock:

- up with battles won, worlds taken and the aim advanced; down with battles lost, worlds lost and ships lost against ships had;
- up when the home is threatened (the defender's resolve), down when the home is safe and the aim is far;
- down when a leader falls (the succession is its own crisis); up when a leader at the front wins;
- **a tick nobody fights in costs both sides**: an armed peace in all but name, which should end in terms within a few ticks rather than stand for seventy thousand years; the unyielding tire of it too, only slower;
- the hate, the monster reckoning and the unyielding posture keep their terms.

### How a war ends

1. **One side is defeated**: its will is spent while the other's holds, or its home is taken, or it is gone. What it loses is what the winner's aim asks, capped by what the winner holds in reach: a world or two, tribute, vassalage, slavery or the end — `yield` and `capitulate` today, reading the aim instead of a fixed "the front and two more".
2. **Both lose the will**: peace, or exhaustion between peoples that cannot treat — as today.
3. **Terms are accepted**: a settlement before either side is broken. This is the third ending, and it is what makes most limited wars short.

### Terms

A side that sues offers what the other's aim wants, as the other's aim is believed, from what it has to give:

- **worlds** — the contested one, or more;
- **vassalage** — the loser's own compromise, offered so as not to be enslaved (below);
- **resources** — tribute for a term, which exists (`TermPeace` against `TermFlow`);
- **artifacts** — a wielded remain or a carried rarity handed over (a mobile source changes holder, as a taker carries one off today);
- **peace alone** on the lines as they stand. The winner can offer too — "we keep what we took" — once its aim is met. The other side's council answers from its own appraisal: it accepts when pressing on is believed to cost more than the terms are worth, or its will is low. The terms are contracts (`contract.go` already makes a peace bought with tribute, `TermPeace` against `TermFlow`); a ceded world is a new term. Accepted terms are a fact, with a truce whose length grows with what was given up.

A side that is afraid (fear high, the balance believed against it) sues early and holds while it waits: point 2 of the brief. A side that believes it is winning presses and refuses terms short of its aim: point 1. A side that believes it is winning and is wrong presses into the battles that correct it: point 3.

### Allies fight

A war joined by pact is a war like any other: **every ally holds its own war council**, and the ally's aim is **defence** of its principal — met when the principal's war ends well, abandoned (a betrayal, as now) when the ally sues alone.

An ally's council reads the alliance as well as the enemy:

- **The other allies**, as it knows them: who else is in the war and pressing, how strong it believes each is, how far each is from the enemy. The balance it appraises is the coalition's against the enemy's, not its own alone — a small member of a strong alliance can believe it is winning.
- **What the alliance is worth**: the pact's age and its members, the renown and infamy the record already keeps (`renown`, `infamy`), the betrayals between them (`betrayed`), and the enemy's menace to itself. Suing alone is a separate peace, which the record counts a betrayal and every member weighs; leaving the alliance for no good reason is the most expensive thing an ally can do, and the council should price it so.
- **Its own stake**: whether the enemy can reach its worlds, and what the war has cost it.

So an ally stays in a war it is losing on its own account when the alliance believes it is winning and is worth keeping, and leaves one it is winning when the alliance is falling apart around it. The cascade already exists: a declaration calls the target's defensive allies, whose joining calls their enemies' allies in turn. With fleets in it, that is a world war (pattern 3); and it already ends the way world wars end — when the principals make peace, the joined wars end with them (`warEnded`'s `pact_peace`).

The sharpening measures how far the cascade reaches today once the joined wars have fleets, before anything is tuned: 53 pacts in seed 5, most of two or three members, one of seventeen.

### Old enemies

What makes the same two peoples fight again is what they remember: the grudge (`resent`, decaying 0.5% a thousand years), the claim an heir holds on its sibling's worlds, the count of wars fought (`Fought`), the world that changed hands last time. A pair with two wars behind it and a grudge standing is a **rivalry**: each side's bar against the other is lower, and the aim **escalates** with each war — the first over a border, the next for redress, the third for the whole (the Punic pattern). The truce (5 to 15 kyr, longer after heavier terms) spaces them out; decay of the grudge ends the rivalry when nothing renews it.

Step 19's loop — two pacifist heirs refighting an empty war five hundred times — is the failure this must not bring back. A war with a council has fleets in it, which is what the loop lacked, and the gate counts loop pairs (more than twenty wars) at zero.

### Cold war

Two peoples that menace each other (`mind.Menaces`, `mind.Threatens`) and whose councils each read the other as too strong to strike — the `Watch` verdict, which exists — are rivals at peace. Three things make that a cold war rather than a quiet border:

- **The arms race.** Shipbuilding gains a deterrence term: build toward the believed strength of the strongest menacing neighbour, times fear (`mind.WantShips` gains an input; the docks already build to the want).
- **Watching.** Pickets and scouts on each other exist (`explore.go`); a rival is watched harder.
- **The incident.** A limited war over a world at the border, which both sides settle quickly because each believes the total war it could become is one it would lose. Fleets meeting in the dark (`intercept.go`) are incidents already.

It ends when a belief changes: a scout finds the rival hollow (`empires.md`: "the war comes when somebody finally scouts"), a pact, or one side's decline.

### Clients, and proxy wars

A vassal is an actor on a leash: its council sits, and may strike free peoples and **other masters' vassals** within reach, never its own master, its master's allies or its master's other vassals. It can call on its master as on an ally, and the master's council answers from its own reading. A war on another master's vassal is not a war on that master unless the master's council chooses to join it — and between two powers in a cold war, each reading the other as too strong to fight directly, it chooses to back its client with relief and ships instead. That is a proxy war (pattern 6), and nothing in it is scripted.

A slave stays inert: no council, no fleet. **Where the clients come from**:

- **Defeat is slavery**, usually, as now: a people broken in war is taken and kept.
- **Vassalage is a compromise the losing side offers** — a term of a suit, as above — and the winner's council takes it when it is good enough: a client that fights at its side and pays is worth more than a slave to a realm with rivals, less to one that only wants the worlds.
- **The strong offer it without a war.** A large realm's council, looking at a small neighbour that would clearly lose a war to it, can offer vassalage outright (earth and water): the small people's council accepts when it believes the war would be lost, and refuses at its peril — the refusal is a reason the large realm can declare on. Most clients should come this way: the large realms ringed by the small peoples that chose not to fight them.

Seed 5 had 34 vassals and 61 slaves standing at the present, and 89 wars ending in enslavement against one in vassalage, so the offer without a war is the change that makes proxy wars possible at all.

**What vassalage is: protection, for a price.** Today it is abstract — the master's levels gain half what a slave gives, the vassal loses a little of its own and half its reach, nothing is paid and nothing is owed, and since a held people cannot be declared on there is nothing to protect it from. Once clients act and can be struck, the bond has to carry both sides of a bargain:

- **The vassal pays, and it is meant to hurt.** A standing tribute: a share of the vassal's **whole income**, not of its spare — bearable in good years, heavy in lean ones — sent to its patron each tick. It is a use in the vassal's flows like any other (in the `Word` category, which the flows keep for contracts), so the vassal's direction decides where it stands: **paid first**, and in a lean tick what goes unfed is the vassal's own works, fleets or road; or **paid short**, and the vassal keeps its own going at the risk of its patron's anger. That choice is the vassal's council's, reading its fear of the patron, the patron's believed strength and nearness, and how hard the times are. A patron paid short takes offence by the share missing — a grudge, and a war council summoned — and may answer with a demand, a punitive war, or by withdrawing its protection; a vassal that keeps paying short is a vassal in revolt in all but name, and the revolt reads it. The tribute the patron receives is income to it, which replaces the flat levels a master gains from a vassal today (a slave's labour stays as it is).
- **How much: the rate is the bond's own, and it moves.** It is meant to be felt, not to cripple: bounded (on the order of 5% to 35% of income, most bonds near a tenth to a seventh — the sharpening sets the numbers against the flows).
  - *Set at the bond*, from the patron's character (greed, a conqueror's stance, how different it finds the vassal push it up), the circumstances (highest on a surrender at the end of a war, lower for an offer accepted without one, lowest for a people that sought a patron itself), the balance as the vassal believes it (the more hopeless its position, the more can be asked), and noise.
  - *Part of the terms*: when vassalage is offered, in a suit or without a war, the rate comes with it, and the small people's council can refuse a rate it could not bear and fight instead. The council decides, as everywhere else.
  - *Reviewed as the bond goes on*, drifting toward a target the patron's character and its regard for the vassal set: **lightened** for a vassal that answered its calls, fought its enemies, or paid in full through lean years; **raised** for one that paid short, traded with its enemies, refused a call, or grew strong enough to worry it. Each change is a chronicle event, so a bond reads as a history.
  - *The patron's own interest is the guard*: a patron that is not a fool does not kill the goose, and lightens the tribute of a vassal in real distress (its fields unfed) rather than lose it; a mean or foolish one squeezes, and the revolt that follows is earned. A patron in decline is asked for less, or not asked.
- **The patron protects.** An attack on a vassal is an attack on the patron's honour. The patron's council is summoned as by a pact's call, and heard more surely than a pact's: it **joins the war**, or at least **sends ships** to stand with the vassal (relief, as an ally's today), or stays out. Whatever it does, the attacker has wronged it — a grudge against the attacker, a personal offence, which the patron's councils read afterwards.
- **Staying out costs the patron.** It is a betrayal of the vassal in the record's sense (`betray`), its renown falls with everyone who weighs it, and a vassal abandoned watches its patron the way a slave watches a declining master: the revolt reads it.
- **The guarantee deters.** Any council weighing an attack on a vassal appraises the patron with it — the patron's believed strength, times how likely it is believed to come, from its record of coming. So **a small people does not dare strike a great power's client**; a rival great power may, when it reads the patron as unwilling or as too weak to face it directly, and that is where proxy wars come from; a coalition may.
- **Accepting it is a calculation.** The small people offered vassalage weighs the war it believes it would lose, and also the protection against its other neighbours against the tribute's cost: a people menaced from two sides takes a patron more readily than one with only the patron to fear.

So vassalage is a bargain both sides can read and both can break: the vassal by revolting when the patron fails it, the patron by abandoning the vassal and paying for it in renown and grudge.

### Momentum, and the machinery of a conquest wave

A people that keeps winning grows bolder, and the rules should let it:

- **A campaign that takes its target asks the council, not the road home.** If the next enemy world is beyond the fleet's hop, it is a decision to press on from the new base rather than an automatic return.
- **Appetite grows with eating.** Worlds taken recently lower the council's bar for the next war and raise the will for the current one: a term the ordinary council and the war council both read, decaying over tens of thousands of years.
- **A conquered world is a base.** The next campaign musters there, not at home.
- **The leader at the front** (step 9) adds its levels where the fighting is, and dies soonest.

So a winning realm keeps going until it is stopped, spent, or its leader falls and the succession breaks it. That is the conquest wave's machinery; **what starts a wave** (the marginal periphery, the unification, the windfall, the hollow target believed hollow) stays in `empires.md`, which then only has to raise a leader in a wave's state and let this proposal's rules carry it.

## The gate: the patterns as measurements

Every pattern is read from the record by a measurement that exists before anything changes (stage 0), on **twenty seeds**, per seed as well as in total — a run's war numbers have a spread wider than their middle, and `specs/notes/war-halving.md` shows a single seed can be a factor of twenty. The thresholds are set at sharpening against stage 0's baseline; the shape of each is:

| pattern | measured as | wanted |
|---|---|---|
| 1 short and long | the length in ticks of wars with a battle; battles per tick of war by length | both a quarter or more of fought wars at three ticks or under and a quarter or more at twelve and over; the long ones fought throughout, not once: the median long war has battles in a third of its ticks or more |
| 2 old enemies | pairs with three wars or more, each fought, separated by peace | some in most seeds; loop pairs (more than twenty wars) none |
| 3 world wars | systems of wars overlapping in time and sharing a people, with fleets on three fronts or more | at least one of eight peoples or more in most seeds |
| 4 border disputes | wars with a world as their aim between peoples of five worlds or more, over within three ticks, a world or two changing hands | the most common kind of war between large realms |
| 5 cold wars | rival pairs at peace for fifty thousand years or more, each building toward the other, with an incident between them | some in most seeds |
| 6 proxy wars | wars between vassals of different masters, a master's ships in them, the masters not at war with each other | some over the batch |
| (vassalage) | attacks on vassals by who attacks (a smaller people, a rival power), and what the patron did (joined, sent ships, stayed out); the tribute's rate by how the bond began and by the patron's character, how it moved over the bond, the ticks it was paid short, and what the patron did about it | a smaller people's attack on a great power's client rare; the patron coming in most attacks, staying out a minority that the record counts against it; rates spread and ordered as the character and the circumstances say, and moving both ways over a bond's life; a vassal's own fields unfed for the tribute a small minority of its ticks; tribute paid in full in most ticks, short mostly in lean ones, and some vassals paying short often enough to anger their patrons |
| 7 conquest waves | a people taking worlds from three peoples or more within fifty thousand years | some over the batch, and more once `empires.md` starts them |

And the health of it, printed beside: campaigns per war by side (the side declared on striking back in a third of fought wars or more), battles per war, worlds taken per war, how wars end, first wars per distinct pair, wars per thousand people-ticks, the decline index's curve, the length of the ages, nothing capped — and the batch's run time, since many more fleets are in flight.

## Stages

**Stage 0: the instrument.** The seven measurements and the health table, in `techstats` and in a `TestWarShape` (`WAR=1`, twenty seeds), reading the record only. No behaviour change: `TestSameHistory` and the legends digest do not move. It prints today's baseline, which is what the thresholds are set against.

**Stage 1: the war council, will, aims and terms.** Press, hold, sue for each side of each war; will that follows the war; the aim column and `War.Aim`; terms as contracts, the ceded world, vassalage and the artifact as new terms; the offer of vassalage without a war; every war under a council however it began; momentum. Patterns 1, 4 and 7's machinery.
- *Shifts the histories*, everywhere.
- *Testable*: a side that believes it is winning presses and one that fears it sues, on fixed inputs to the mind; a defender strikes back; a limited war ends when its world is taken and the terms are accepted; will rises with a battle won and falls with one lost; the side declared on can win; an inherited war gets campaigns.

**Stage 2: allies and old enemies.** The ally's aim, its own council, the cascade read and tuned; the rivalry, its lower bar and its escalating aim. Patterns 2 and 3.

**Stage 3: cold war and clients.** The deterrence term in the want, the rival watched, the incident as a limited war; the vassal's council on its leash; vassalage as a bargain — the vassal's standing tribute on its whole income at a rate set by the patron's character, the circumstances of the bond and noise, part of the terms, and reviewed with the patron's regard; its place in the vassal's direction and the patron's anger at a tribute paid short; the patron's guarantee (summoned, joining or sending ships, the attacker's offence, the betrayal of staying out), the guarantee read by every council weighing an attack on a client, and the small people's weighing of protection against tribute. Patterns 5 and 6.

Stages 0 and 1 are plan step 10; stages 2 and 3 are plan step 11. Each regenerates the batch once. Stage 1 is the large one and is worth landing and reading alone, since everything after it is read against the wars it makes.

## Implementation notes

- **The mind**: `internal/mind/war.go` — `WarCouncil(WarInput) WarVerdict` (press, hold, sue, with a reason for `-ai`), `AnswerTerms`, the aim's escalation and momentum, all tuned in `mind.Tuning.War`. The history gathers the inputs and executes, as `council.go` does for the first strike.
- **Where it runs**: in the people's civ step beside the council, over the people's own wars, so its draws are the people's stream's (the streams rule). A summons from the war phase sets a flag, as `Summoned` does now.
- **The aim**: a column `aim` on `causes.json`'s war causes; `War.Aim`, with the escalation reading `Nth`; exported and in `FORMAT.md`.
- **Terms**: `Term` gains a ceded world, vassalage and an artifact (a mobile source or a wielded remain changing hands); the terms are a `Contract` and its facts; accepted terms are an `FPeace` of a new way, or a fact of their own (the sharpening decides), with the terms in its parameters.
- **Events**: a war council's verdict is a chronicle event when it changes (a side turning from hold to press, from press to sue); an offer refused and an offer accepted are facts, so the tellings hold them.
- **Watch the loop and the cascade**: loop pairs and cascade seeds are printed per seed; a change that brings either back is a failure of the stage, whatever the totals say.
- **Run time**: more fleets in flight is more cost in the expeditions phase and the sightings; the batch time is printed, and a second optimisation pass may be wanted after step 11.

## Settled in review (2026-09-23)

- **The time scale.** A war stays one war; a long one is a string of campaigning seasons a tick each, and a tick nobody fights in costs both sides, so staring ends in terms. The gate reads battles per tick of war.
- **Terms.** Worlds, vassalage, resources and artifacts can all be offered, and peace alone.
- **Allies.** Every ally's council decides for itself, reading the other allies as it knows them and what the alliance is worth; breaking it is priced as the betrayal it is.
- **Vassals and slaves.** Defeat is slavery, usually. Vassalage is the compromise a losing side offers, and what a large realm offers a small neighbour that would clearly lose a war to it.
- **Vassalage is protection for a price.** The vassal pays a standing tribute of a share of its whole income, which in lean times it must put before its own needs or risk its patron's anger; an attack on it is an offence to the patron, who joins the war, sends ships, or stays out and pays for it in renown and in the vassal's loyalty; and the guarantee deters, so a small people does not dare strike a great power's client.

## Open questions

- **Whether a vassal's leash reaches its master's rivals only when the master consents**, or whenever the vassal's own council reads the target as fair game.
- **Momentum's strength before `empires.md`**: enough to make a winning realm keep going, not so much that one conqueror takes every seed before the cycle is tuned.
- **What a war council costs.** One per side per war per tick on the cadence; seed 8's cascade had 1551 wars.
- **Whether the muster should declare at the launch** instead of at its start: today a people announces a war and then spends three ticks gathering, which is the warning a real mobilisation gives, but it also makes the side declared on wait three ticks for a war it could already be fighting.
