# War: how peoples fight, and how a war ends

**Status:** Stages 0, 1 and 2 built (2026-09-24); stage 3 built and parked on the branch `war-stage3-wip` with its gate short, and the tuning of all four gathered for a follow-up in `specs/notes/war-tuning-followup.md`. Placed in `specs/plan.md` as steps 10 and 11, after `leaders.md` (a leader at the front is one of the things a war council reads) and before `empires.md`, whose conquest wave and coalitions ride on the machinery here and whose tuning pass is read against the wars this leaves. Read by `art.md` (a war is most of the circumstances worth making anything about) and `miracles.md`.
**Last updated:** 2026-09-24

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

Every pattern is read from the record by a measurement that exists before anything changes (stage 0, `internal/warshape`, printed by `TestWarShape` and by `techstats`), on **twenty seeds** at 400 stars, per seed as well as in total — a run's war numbers have a spread wider than their middle, and `specs/notes/war-halving.md` shows a single seed can be a factor of twenty. The baseline is stage 0's, at `6e2af9f`'s history; the thresholds are set against it. "Most seeds" is 14 of 20.

| pattern | measured as | baseline (stage 0) | wanted |
|---|---|---|---|
| 1 short and long | fought wars (a battle in them) by length in ticks; the share of a long war's ticks carried: a battle, or a campaign in flight (settled 2026-09-24) | 28% of wars fought; of those 26% at three ticks or under, 42% at twelve and over; the median long war fought in 6% of its ticks; short under a quarter in 8 seeds, long in 2 | 60% of wars fought or more; a quarter or more short and a quarter or more long, in total and in most seeds; the median long war carried in a third of its ticks or more |
| 2 old enemies | pairs with three fought wars or more and twenty wars or fewer, the wars separated by peace (a pair has one war open at a time) | 111 pairs, in 18 seeds; 25 loop pairs (more than twenty wars), the most 117 | some in most seeds, a hundred or more over the batch (the count is stage 2's to raise: stage 1's settled grudges lower it, accepted 2026-09-24); loop pairs none |
| 3 world wars | systems of wars joined by a shared people while both were open, read at their widest moment: peoples at war at once, fronts with a battle | 13 of eight peoples or more on three fought fronts or more, in 6 seeds; the widest 22 peoples on 5 fronts | at least one in most seeds |
| 4 border disputes | wars between realms of five worlds or more when it began, over within three ticks, one or two worlds taken; the aim once stage 1 records it | 2180 such wars: 61% short and unfought, 33% long, border disputes 5% | the most common kind of war between large realms, a third of them or more |
| 5 cold wars | pairs that have fought, both large, at peace with each other fifty thousand years or more with an incident in the stretch (a short war, or a battle with no war); each building toward the other, once stage 3 records the rival watched | 78 pairs, in 4 seeds, all without the build-up; 429 quiet pairs | some in most seeds, with the build-up |
| 6 proxy wars | wars between vassals of different masters, a master's ships in them (a battle of its own against the other side, or a fleet sent), the masters not at war with each other | 8 wars between vassals of different masters, none with a master's ships | five or more over the batch, in three seeds or more |
| (vassalage) | bonds by how they began (war, meeting, birth or uplift, a rider, other) and how long they lasted; attacks on vassals by who attacks (smaller than the patron, a rival power, the patron) and what the patron did (joined, sent ships, stayed out); the tribute's rate by how the bond began and by the patron's character, how it moved over the bond, the ticks it was paid short, and what the patron did about it, once stage 3 records them | vassals: 47% by meeting, 9% by war; slaves: 18% by a rider, 9% by war; both held a median of 14 to 16 kyr; 74 attacks on vassals, 46% of them by a smaller people, the patron staying out in 87% | a smaller people's attack on a great power's client rare, under a fifth of the attacks; the patron coming in most attacks, staying out a minority that the record counts against it; bonds by war more common than today, and lasting; rates spread and ordered as the character and the circumstances say, and moving both ways over a bond's life; a vassal's own fields unfed for the tribute a small minority of its ticks; tribute paid in full in most ticks, short mostly in lean ones, and some vassals paying short often enough to anger their patrons |
| 7 conquest waves | a people taking worlds in war from three peoples or more within fifty thousand years | 3 peoples, in 1 seed, from 4, 4 and 3 | five or more over the batch, in three seeds or more, and more once `empires.md` starts them |

And the health of it, printed beside, with its baseline:

| | baseline (stage 0) | wanted |
|---|---|---|
| campaigns a war | the declarer 0.43, the side declared on 0.00; 60% of wars with no campaign | the side declared on striking back in a third of fought wars or more (today 1%) |
| battles, worlds taken a war | 0.57 battles (the upper quartile 1), 0.43 worlds | read, not gated: more of both is the point |
| how wars end | tribute 19%, peace 18%, exhaustion 16%, a side's fall 12%, peace by the pact 12%, vassal 9%, enslaved 8%, capitulation 3%, truce 2% | a war that ends in a truce with nothing settled a minority; defeat and terms both common |
| first wars per distinct pair; wars per thousand people-ticks | 0.007; 0.25 | read, not gated (`specs/notes/war-halving.md`) |
| the age | 38 / 48 / 56 Myr; the decline index at the present 0.46 / 0.56 / 0.60, at the waning 0.40; none capped | ages inside 20 to 80 Myr, nothing capped |
| run time | 1h30m for the twenty, seed 8 the longest at 11 minutes with the others running | within twice the baseline |

## Stages

**Stage 0: the instrument.** The seven measurements and the health table, in `techstats` and in a `TestWarShape` (`WAR=1`, twenty seeds), reading the record only. No behaviour change: `TestSameHistory` and the legends digest do not move. It prints today's baseline, which is what the thresholds are set against. *Done*: see "As built (stage 0)".

**Stage 1: the war council, will, aims and terms.** Press, hold, sue for each side of each war; will that follows the war; the aim column and `War.Aim`; terms as contracts, the ceded world, vassalage and the artifact as new terms; the offer of vassalage without a war; every war under a council however it began; momentum. Patterns 1, 4 and 7's machinery.
- *Shifts the histories*, everywhere.
- *Testable*: a side that believes it is winning presses and one that fears it sues, on fixed inputs to the mind; a defender strikes back; a limited war ends when its world is taken and the terms are accepted; will rises with a battle won and falls with one lost; the side declared on can win; an inherited war gets campaigns.
- *Done*: see "As built (stage 1)"; three of the gate's rows are left to the tuning at the end.

**Stage 2: allies and old enemies.** The ally's aim, its own council, the cascade read and tuned; the rivalry, its lower bar and its escalating aim. Patterns 2 and 3.
- *Done*: see "As built (stage 2)"; world wars in most seeds is not reached, and a question for the user.

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

## As built (stage 0)

`internal/warshape` reads a `record.Run` and nothing else; `TestWarShape` (`WAR=1`; `WAR_STARS`, `WAR_SEEDS`, `WAR_FROM`) exports each world it generates and reads it, and `techstats` prints the same lines as "The shape of war" in `report.md`. The choices made in building it:

- **Joining the record.** A battle and a campaign fleet carry no war; they are joined to the war between their two peoples open at their year, since a pair has one war open at a time. Every battle of the baseline joined one. A fleet is a campaign by its kind and counted for the side that launched it.
- **Worlds and masters through time** are folded from the silent `world_held`, `world_lost` and `master` events, read as they stood before the year asked about, so a war's sides are sized as they were when it began.
- **A system of wars** is the wars joined by a shared people while both were open; it is read at its widest moment (the start of one of its wars), since a chain of overlaps can run for millions of years and its size over all time says nothing.
- **A bond's origin** is what else the record holds in the year it began: a war between the two ending, a meeting, the held people's birth or uplift, a rider taking it. What the baseline showed: most vassals are made at a meeting, most slaves outside war are a rider's hosts, and bonds are short (a median of 14 kyr for a vassal, 16 for a slave) — the protection stage 3 builds has to last to mean anything.
- **What the record cannot yet say** is printed as such: a war's aim (stage 1, `War.Aim`; until then a border dispute is read by its size, length and the worlds taken), the build-up of a cold war (stage 3's rival watched), and the tribute's rate, its movement and the ticks paid short (stage 3). Each stage that adds one exports it and extends the reader, so the row reads the thing itself.
- **What the baseline says the stages must fix**, beyond the patterns: only 28% of wars see a battle, the side declared on never strikes back (1% of fought wars), and a long war is fought in 6% of its ticks — the stare the review found, now counted.

## As built (stage 1)

The war council, will, aims and terms, as `internal/mind/war.go` (the judgments) and `internal/history/warcouncil.go` (the gathering and the executing), every number in `mind.Tuning.War`. The choices made on the way:

- **The aim** is a column on `causes.json`'s war causes, raised to ending for the hating and to submission for a conqueror, and `War.Aim` is exported (`FORMAT.md`). A conqueror whose first fleet sails for a border world and not the home fights for that world. The side declared on holds, and presses to retake what it lost (`aimTarget` reads the worlds it lost that the enemy still holds).
- **The council** sits on the ordinary cadence (`Cadence` 0.5 a tick) and whenever the war summons it: a battle, a world taken or lost, a leader fallen, the side declared on at the declaration. It presses at its bar (declarer 0.5, side declared on 0.55 less 0.1 once it has lost more than it took, an ally 0.55, the hating 0.3; the conqueror, the vengeful and the unyielding 0.1 bolder), holds, or sues: with its aim met (peace on the lines), its will under 0.15, afraid below 0.1 + 0.2 × fear odds once the war has begun, or three ticks with nobody fighting and no fleet it would send (nine while its docks build the fleet it wants). The unyielding and the hating never sue. A campaign out is the council waiting.
- **Strength counts the ships out** against the enemy (`outAgainst`), so a people that sends most of its fleet does not read itself beaten the tick it sails.
- **Terms**: the side that sues offers the first of what the other's aim wants that it has — worlds (never the home, never to a horde), tribute, a mobile artifact, vassalage (not from the unyielding, a conqueror or what cannot be held) — or peace on the lines. The other answers from the terms' worth to its aim against its odds of taking the rest, its greed and a conqueror's hunger, its will, and whether it would press at all. Accepted terms are a fact (`settled`), refused ones an event (`terms_refused`), a truce three thousand years longer per world given up, and claims on what was settled renounced. **Who won**: peace on the lines offered with the offerer's aim met is the offerer's; anything else offered is the war bought off, and the side that took it won.
- **Will** moves with the fighting: a battle won or lost, ships lost against ships had (`ShipLoss`), a leader winning at the front or falling, worlds taken and lost. A tick nobody fights costs 0.1, growing with the ticks since the last battle (`IdleRamp` 5) and doubled after fifty thousand years; the unyielding pay half, the home threatened halves it, a total aim that cannot reach the home pays half again; a fleet in flight stops it. Hunts keep the old clock.
- **How a war ends**: a side's will spent is a yield to what the other's aim asks (a world or two, tribute, the home for a total aim), or peace with the other named winner when it has nothing the other could take; a home taken, however it ends, is the taker's war; a people that yields to the same power three times bends as its vassal; a war with a side that has passed under a master ends `held`; and no war is declared on or by a held people, but a hunt, whose hunter cannot know whose it is.
- **After**: the side ahead (the winner, or with nobody yielding the side with its aim met and more taken than lost) has its grudge settled; the others resent as before. The loser, and a declarer that gained nothing, grows **wary** of the other: 0.1 on its bars per war, up to 0.8, fading at 0.2% a thousand years; the wariness can lift a bar past certainty, so a people beaten often enough does not go to war with that enemy again however sure it is.
- **Vassalage without a war** (`yoke`): a conqueror or opportunist of five worlds or more and three times the other's, about to strike at odds of 0.85 or better, offers first; the small people bends below 0.3 + 0.3 × fear (0.2 more for the submissive), and a refusal is the cause `defiance`.
- **Momentum**: each world taken adds a point of appetite, halving every twenty thousand years; it takes 0.03 a point off the bars (0.15 at most) and adds 0.1 a point to the will a war starts with. A conquered world with no next target in a hop keeps the fleet as its guard, and the council sends the next campaign from there. The muster declares its war when the fleet sails, not when it is called.
- **Allies** keep the old answer until stage 2, with one damper: an ally goes to war in its own name only with ships to send and not wary of the enemy to the limit (`arms`); otherwise it answers as an ally with no front does. A seed's cascade of 77 wars in a tick between shipless allies came from that.
- **Loops cut with the scenarios and the watchers** (`specs/notes/war-stage1-handover.md` has the list): a hunt on another's vassal ended at once and declared again (160 in a seed); a spared home told as a peace with no winner; the winner's grudge renewed at the war's end; a vengeful people bought off with its grudge left standing; a rider trying a people that had barred it; hunts on the same hole from the same evidence (a closed hunt spends it, `Civ.HuntEnded`; a pacifist hunts nothing that does not reach its home; a hunter three times worsted lets the hole be, `Kinds.HuntWary`).
- **The tools**: `cmd/scenario` and `internal/history/scenario.go` build a small world from a spec and tell a run of it; the specs in `internal/history/testdata/scenarios` are regression tests on five seeds. `internal/history/watch.go` runs detectors inside `Generate` (a loop pair, a people passed between masters, an endless war, a jump in wars) and dumps each catch with the reasons that led there (`Config.Reason`, every decision's reason without the `-ai` log); `TestWarShape` takes `WAR_WATCH=dir` and `WAR_TUNE=G.F=v,...`, and `worldgen` takes `-watch`. The watch also tallies what each tick of a long fought war was.

The gate on twenty seeds at 400 stars, against stage 0:

| | stage 0 | stage 1 | wanted |
|---|---|---|---|
| fought | 28% of all wars | 64% of fleet wars (57% of all) | 60% ✓ |
| short / long of fought, in total | 26% / 42% | 35% / 28% | a quarter each ✓ |
| short / long, in most seeds | 8 / 2 seeds | 14 / 8 of 20 | 14 each: short ✓, long ✗ |
| a long war's ticks carried (battles alone) | (6%) | 34% (18%) | a third ✓ |
| old enemies; loop pairs | 111 pairs, 18 seeds; 25 | 127 pairs, 17 seeds; 0 (the most wars of a pair 17) | ✓ |
| world wars | 13, 6 seeds | 27, 6 seeds | stage 2's |
| border disputes | 5% | 25% (long 60%) | a third, the most common ✗ |
| conquest waves | 3, 1 seed | 53, 5 seeds | 5+, 3+ seeds ✓ |
| side declared on strikes back | 1% | 16% | a third ✗ |
| how wars end | truce 2% | terms 31%, a side's fall 21%, enslaved 14%, peace 10%, exhaustion 6%, held 6%, truce 1% | truce a minority, defeat and terms common ✓ |
| the age | 38 / 48 / 56 Myr | 33 / 42 / 47 Myr, none capped | ✓ |
| run time | 1h30m | 1h37m | within 2× ✓ |

What the ticks of a long fought war were: fought 23%, a fleet in flight 20%, a side waiting on its docks for the fleet it wants about a quarter, a side with its aim met offering the lines and refused about a tenth.

## As built (stage 2)

Allies and old enemies, as `internal/mind/war.go` (`Bound`, `Escalate`, the rival's bar in `mind.Bar`) and `internal/history` (`warcouncil.go`, `pact.go`, `appraise.go`), every number in `mind.Tuning.War` and `mind.Tuning.Bar`. The choices made on the way:

- **The ally's council** is the war council every war already had, for the aim defence. What it adds is what the alliance is worth to the ally (`mind.Bound`, 0 to 1): 0.3 to begin with, 0.2 more for a pact a hundred thousand years old, 0.05 a member past two up to three, 0.3 when the enemy menaces the ally itself, 0.1 a point of the principal's renown; 0.5 less when the principal has broken faith with it. At its fullest it takes 80% off the ally's reasons to sue alone (the will, the fear, the idle ticks) and 1 off any terms offered it alone, since peace alone is a separate peace. An ally's war counts as fought while its principal's is, so the ally does not tire of a war its principal is fighting.
- **An ally too small to carry the war** to the enemy (its council holding for the odds, the ships or the reach) sends ships to stand at its principal's home, sized as a call's relief and never leaving its own home unsafe. A battle records the allies whose relief stood in the defender's sky (`Battle.Relief`, in `FORMAT.md`), and the battle counts for the ally's own war, in the history and in the instrument.
- **The rivalry** is two peoples with two wars fought between them and a grudge standing on either side (`rival`). A posture that wants war strikes its old enemy at a bar 0.1 lower; one that wants no other war (the defensive, the confederate, the unyielding) wants its old enemy at 0.6. Not the submissive, not a pacifist.
- **The escalating aim** (`mind.Escalate`): the second war between two with a grudge standing is for redress at least; the third and after is for the whole only when the declarer holds twice the other's worlds — Rome was the larger by the third Punic war — and for redress again between rivals of a size. The whole at the third war regardless ended the rivalries it was meant to carry: over twenty seeds, 104 old-enemy pairs with it and 188 without, and 5569 wars against 7194.
- **Redress asks back what was lost** (`War.Asks`): the worlds the declarer lost to the other in their last war, one to three, is what meets the aim and what terms must give. Met at one world whatever was lost, the rivals' wars were over as they began (fought redress wars 58% short, 8% long); at two for every redress a third of the wars went, and the old enemies with them.
- **Tried and taken back**: an ally that comes with relief going to war in its own name. It filled the systems of wars with belligerents that fought nothing (joined wars fought in 19%) and gave fewer world wars, not more.
- **Loops cut** (each a test that fails without it): a master hunting its own anti-memetic slave for the settlers it lost at the slave's world, and enslaving it again twenty times (a loss at the hands of a people the victim holds is no hole in its ledger); the same hole hunted every hundred and fifty thousand years because the wariness that stopped it faded as fast as each failed hunt added to it (failed hunts are remembered, `Civ.HuntsFailed`); a realm buying the same neighbour off twenty times, a world at a time (a war bought off is a yield, and a people bought off twice offers itself as a vassal); three peoples sending settlers for a million years to a world an unseen people held (a ship lost without trace makes the star one ships do not come back from, `dread`); a muster outliving the war it was called in and opening the next (a muster knows its war, and stands down when it ends); and a submissive heir striking its sibling for a claimed world twenty-nine times, each war given up in two ticks, because the prize of the claim brought the bar down under a wariness capped at 0.8 (past five wars come off worst, fading, a people does not go to war with that enemy again at all: `War.WaryStop`). The watch's reasons now keep the council's verdict on a people it might strike, which is what that last one needed to be read.
- **The tools grew**: `TestWarShape` prints the gate with a verdict per row (`warshape.Gate`), keeps it (`WAR_SAVE`) and prints an earlier one beside it (`WAR_BASE`), takes a list of seeds (`WAR_LIST`), and writes every seed's record (`WAR_OUT`); `cmd/warshape` reads those records again in seconds — the report and the gate, the widest systems of wars war by war (`-systems`), the pairs past a count (`-pairs`), a people's chronicle over a span (`-events`), and how many peoples send fleets (`-census`). The report gains the allies (pacts, joined wars, separate peaces, relief) and the rivalries (wars by their place between the pair, and their aims).

The gate on twenty seeds at 400 stars, against stage 1 on the same seeds; and, since one batch of twenty swings by more than the differences read here, both stages again on seeds 21 to 40 (stage 2's on the code before the last two changes, redress's ask and the lost fleet's worlds, with two seeds lost to the bug the second fixed):

| | stage 1 (1–20) | stage 2 (1–20) | stage 1 (21–40) | stage 2 (21–40, 18 seeds) | wanted |
|---|---|---|---|---|---|
| fought, of fleet wars | 64% | 60% (59.8) | 62% | 62% | 60% |
| short / long of fought | 35% / 28% | 42% / 20% | 38% / 26% | 44% / 17% | a quarter each |
| short / long, in most seeds | 14 / 8 | 15 / 6 | 14 / 16 | 15 / 7 | 14 each |
| a long war's ticks carried | 34% | 40% | 41% | 44% | a third ✓ |
| old enemies; loop pairs | 127, 17 seeds; 0 | 122, 18 seeds; 0 | 66, 15 seeds; 3 | 100, 13 seeds; 0 | 100+, most seeds; none |
| world wars | 27, 6 seeds | 32, 6 seeds | 11, 5 seeds | 26, 6 seeds | most seeds |
| border disputes | 25% | 23% | 30% | 25% | a third |
| conquest waves | 53, 5 seeds | 35, 6 seeds | 1, 1 seed | 26, 7 seeds | 5+, 3+ seeds |
| side declared on strikes back | 16% | 13% | 10% | 10% | a third |
| the age | 28–70 Myr | 20–67 Myr | 25–66 Myr | 19–50 Myr | 20 to 80 |

Allies (seeds 1–20): 865 wars joined by pact, 36% fought, the ally sending a campaign in 19%; they ended in its principal's peace in 44%; 141 separate peaces; 5491 relief fleets. Rivalries: second wars fought 74%, third and later 82%.

**What moved, and what did not.** Old enemies rose and the loops went, on both samples; world wars rose on the second. **Long wars fell** on both, 28% to 20% and 26% to 17% of fought wars: more wars between old enemies, and those for redress or a world short. It is the one row stage 2 made worse, and goes to the tuning below.

**World wars in most seeds is not reached, and cannot be by the cascade.** In fourteen of the twenty seeds between 12 and 34 peoples ever send a campaign in the whole age, and no more than 10 to 27 send fleets of any kind in their busiest million years (`cmd/warshape -census`); a system of eight peoples at war at once on three fought fronts would need a third or more of them in one conflict. The world wars come in the crowded seeds: 5 of the 6 with 38 or more fleet senders in a million years, and one other. Whether the row should be read over the seeds that have the peoples for one, or left to `empires.md`, which is what fills the quiet seeds, is the user's to decide.

## As built (stage 3)

Cold war and clients, as `internal/mind/client.go` (the judgments, every number in `mind.Tuning.Client`), `internal/history/client.go` (the leash, the call, the guarantee) and `internal/history/tribute.go` (the bond). The choices made on the way:

- **The rival watched**: each council names its rival, the neighbour it most fears in reach (`watchRival`, the `threat` the pacts already read), and a change of rival is a note in the record (`rival`). The docks build toward a share of the rival's fleet as believed, 0.3 a point of fear up to half of it (`mind.Deterrence`, a term in `WantShips`), and a picket is kept against the rival whatever the fear. The instrument reads a cold war's build-up as each of the two naming the other its rival in the stretch; fleets meeting in the dark with no war between the two are incidents, as the proposal has it.
- **The leash** (`mayWar`): free peoples and vassals may be at war, a slave with nobody; a vassal never with its master, its master's allies or its master's other vassals; a patron with its own client only over a tribute unpaid; a hunt always. A vassal sits its councils and makes no pacts. What a vassal takes in war goes to the patron at the top of its chain (`overlord`): a client's conquest is its patron's.
- **Another's client is struck for its worlds**: a war on it is for a world, not its submission, and its home is no target (`protected`), so a client with nothing but its home is not struck at all. Without that, a home taken and spared was a free win and two powers took one client off each other a hundred and thirty-five times.
- **The call**: a client struck calls its patron, who always hears it; the patron's council (`mind.AnswerClient`) joins the war with the will and the odds, sends ships to stand with the client with the will and not the odds — a power it would not face directly, which is how a proxy war is fought — or stays out, a betrayal (`abandoned`) that puts the question of revolt to the client. The attack is an offence to the patron whatever it does.
- **The guarantee**: an attacker reckons a share of the patron's fleet as believed, by the patron's record of coming (calls answered of calls heard, half a call answered before the first), as ships standing over the client's world (`mind.Guarantee`, `ComeRate`). In the scenarios a jackal that strikes a weak people alone in every seed does not strike the same people as a great power's client.
- **The tribute** (`tribute.go`): the bond sets a rate on the vassal's whole income from the patron's greed, a conqueror's stance, how different it finds the vassal, how the bond came about (surrender, offer, sought) and how hopeless the vassal's position was, with noise, between 5% and 35% (`mind.TributeRate`); it is a use in the vassal's flows, fed after the fields when the vassal pays first and where the direction puts the word when it risks paying short (`mind.PaysFirst`: loyalty and awe against lean times); paid, it is income to the patron and replaces the levels a vassal gave its master; paid short, it angers the patron by the tick, until the patron makes war on its client for it (the cause `unpaid`, once a truce, the war spending the anger). The bond is reviewed every ten thousand years, the rate moving toward its target, lightened for a vassal that paid in full or is in distress under a patron no fool, raised for one that paid short or grew strong (`mind.ReviewRate`). The bond and each review are notes in the record (`bond`, `bond_rate`), and the ticks owed and paid short are in the tally.
- **Loops cut**: a patron punishing its client every two thousand years (the truce, and the war spending the anger); a client taken back and forth between powers (the client struck for worlds only); allies joined against an ally's war fighting on alone and again at every call (a war's end ends the wars joined to it, whether it was a principal's or an ally's); and old enemies back every hundred thousand years as the fading wariness let them under the line (wars come off worst are remembered, `Civ.Worsted`, and five of them end it).

STAGE3_GATE

## Settled in review (2026-09-23)

- **The time scale.** A war stays one war; a long one is a string of campaigning seasons a tick each, and a tick nobody fights in costs both sides, so staring ends in terms. The gate reads battles per tick of war.
- **Terms.** Worlds, vassalage, resources and artifacts can all be offered, and peace alone.
- **Allies.** Every ally's council decides for itself, reading the other allies as it knows them and what the alliance is worth; breaking it is priced as the betrayal it is.
- **Vassals and slaves.** Defeat is slavery, usually. Vassalage is the compromise a losing side offers, and what a large realm offers a small neighbour that would clearly lose a war to it.
- **Vassalage is protection for a price.** The vassal pays a standing tribute of a share of its whole income, which in lean times it must put before its own needs or risk its patron's anger; an attack on it is an offence to the patron, who joins the war, sends ships, or stays out and pays for it in renown and in the vassal's loyalty; and the guarantee deters, so a small people does not dare strike a great power's client.

## Settled at stage 1 (2026-09-24)

- **A fleet on its way is the war being fought.** A long war's ticks are read as carried when a battle is fought in them or a campaign of either side is in flight at the other; a flight of centuries at sail is a campaigning season as much as the battle at its end. The instrument prints both shares, battles alone and carried.
- **Old enemies fall at stage 1, and that is accepted for now.** The war council settles the winner's grudge (the winner of a spared home, of a yield, of terms bought off), so fewer pairs come back to the same quarrel; the rivalry that brings them back is stage 2's, and the row is read there.

## Open questions

- **Whether a vassal's leash reaches its master's rivals only when the master consents**, or whenever the vassal's own council reads the target as fair game.
- **Momentum's strength before `empires.md`**: enough to make a winning realm keep going, not so much that one conqueror takes every seed before the cycle is tuned.
- ~~**What a war council costs.**~~ Answered at stage 1: the twenty seeds ran in 1h37m against stage 0's 1h30m; the cost that mattered was the linear war lookup, now indexed by pair.
- ~~**Whether the muster should declare at the launch**~~ Answered at stage 1: it does; a muster that stands down has started nothing.

## Tuning left for later

The research for this tuning — the variants tried and their numbers, the stage 3 batches, the loops, the new definition of a world war (more than half the galaxy's worlds held by peoples at war in one system), and how to read the gate on forty seeds — is `specs/notes/war-tuning-followup.md`.

Stage 1 was landed with three of its rows short, to be tuned with the scenarios and the watch's tally rather than knob by knob: at twelve seeds two batches of near-identical rules differ by more than one number moves the result (two knob batches were tried and dropped; see the handover note), so each is to be worked through its mechanism and read on twenty seeds.

- **Long wars in most seeds** (8 of 20; wanted 14): long wars are a quarter of fought wars in total but concentrated in the seeds with many peoples.
- **Border disputes** (25% of the wars between large realms; wanted a third and the most common kind): a border war between two large realms settles by a ceded world at tick four or five, just past the row's three, because a side waits three idle ticks before it sues and sits on the cadence; and 60% of large realms' wars run long.
- **The side declared on striking back** (16% of fought wars; wanted a third): it reads its odds against the declarer, which picked the fight, and seldom clears its bar.
- **Old enemies and the long war's battles** were settled at stage 1 (above) and are to be improved later too: the rivalry at stage 2, and battles in a long war's ticks beyond the fleets in flight.
- **Long wars, again** (stage 2): a fifth of fought wars on both samples, down from over a quarter. The wars stage 2 adds are old enemies' for redress or a world, and short; the long ones are the total aims' and the allies'.
- **Read the gate on forty seeds.** Two samples of twenty differ by more than most of the rows move (old enemies 127 and 66 under the same rules, conquest waves 53 and 1): a row should be read on seeds 1 to 40 before a change is kept or dropped for it.
