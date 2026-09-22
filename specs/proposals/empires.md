# The cycle of empires: consolidation, shattering, and the force decline needs

**Status:** Draft (2026-09-22). Depends on `continuity.md` and `leaders.md`, and carries **`decline.md`'s stage 3** — the force — which moved here because the lever decline specified and the separatism specified here are the same check with two outcomes.
**Last updated:** 2026-09-22

## Problem

The step-6 sampling of eight ages at 400 stars says the galaxy consolidates once and then stops. Every seed produces a large holder — 59, 61, 109 worlds. **One seed in eight sees it fall**, and after that fall nobody re-consolidates: the largest holder goes 6, 8, 11, 12 worlds over the next twenty-four million years. Meanwhile 89% to 100% of what is held at the present is held by peoples over five million years old; running wars in the last third of the age are nought or one in six seeds of eight; and the hazard, the one ambient pressure that could reach a big empire, is already pinned at its 2.5 ceiling in four seeds.

So there is no cycle. There is a rise, a settling, and a very long wait.

The arc wanted is: many small warring states, consolidating into warlords and federations, consolidating again into a few empires, which come apart — two or three times over, each peak lower than the last, inside a declining envelope.

**Two structural facts about that arc.** Fragmentation and decline are different variables: a shattering redistributes worlds among heirs and does not move the held share at all. And re-consolidation needs war; a galaxy at peace cannot re-assemble its pieces. Both have to be built, and they are tuned against different measures.

## Scope

**In.** Separatism and the quiet letting-go; coalitions against a hegemon; war degrading what it is fought over; plagues reading connectivity; the conquest wave; later-striking filters; the concentration measures; and decline's stage 3, whose gate this proposal inherits.

**Out.** Leaders themselves and succession (`leaders.md`). Lifespan, continuity and the archive (`continuity.md`). The galaxy's geography (`galaxy-in-motion.md`).

## Design

### The shared quantities

Mechanisms cohere when they push on a few shared numbers rather than each carrying a private trigger and a private effect — the hazard already works this way and needs no coordination. Everything below expresses itself through four:

- **Overreach** — how far a realm extends beyond what its seat can reach and feed. New.
- **Spentness** — how degraded a world or source is. New; this is what makes each peak lower than the last.
- **What is remembered against the holder** — grudge and the monster reckoning. Exists.
- **Continuity** — `continuity.md`. The wall and the hazard remain as they are.

### Separatism, and the quiet going

One check, two outcomes, on a realm that is set in its ways, has been still a long time, and holds provinces the centre has stopped paying for — `Shed`, `DormantSince` and `ShedSince` already track exactly that, so **the trigger is economic and the consequence is military.**

- The province has enough to stand on: it **declares**, becomes a people, and is at war with its parent from the moment it does. Contiguous and peripheral, so the successor is a viable polity that can go on to fight its neighbours — which is what makes a shattering the start of the next cycle rather than the end of the story.
- It has not: it is **let go**, quietly unowned, ruins left. This is `decline.md`'s lever 1, and the sad end of the arc rather than a separate mechanism.

Neither may renew. `loseSystem` ends with `w.renew(c, 0.05) // a loss is something new`, which takes stiffness off and resets `Still` — the two conditions this check reads — so the lever would switch itself off after one use. The exception carries a comment saying why.

A world let go is **spent** for a long while: settlement and loss already balance at 2 to 31 stars a million years around a total that does not fall, so a freed world would simply be refilled. Measure it at zero before setting it.

### Coalitions

The council's fear term scales with the target's strength, so **the largest empire in the galaxy is the safest thing in it** — which is why the late map never moves and why nothing ever attacks a weakening hegemon. Where a people is at war with everyone and remembered as a monster by everyone, fear should **invert**: the threat drives peoples together instead of apart.

The ingredients exist — the monster reckoning is galaxy-wide and shared, `Pacts` exist, and the slight already makes third parties take sides in other peoples' wars. What is missing is the council reading "this is everyone's problem" as a reason to join rather than to stay away. The crazed immortal is the forcing case; ordinary hegemons get it too.

Also to be checked before anything is built: whether the failure is opportunity or will. Reach coverage at the present runs from 0.20 to 0.85, so in some seeds the survivors cannot cross to each other; and `Intel`/`Scouted` may be stale, so an ossified empire still *looks* strong. If it is the second, the mechanism is better than a fix: reputation outlives strength, and the war comes when somebody finally scouts.

The revolt path from inside already exists — `Declines` and `Seen`, "declines suffered, watched by slaves for revolt". Check whether it ever fires before building anything new.

### War spends what it is fought over

A world or source bombarded loses a share of its yield, recovering over tens of millions of years — restored by the next age, shrinking through this one. This is what makes the cycles decay instead of running forever, and it gives the present a texture: the richest worlds are the ones nobody fought over.

Two cautions. It is a positive feedback into decline, so it is tuned *with* the index and not before. And it makes war self-limiting, which could suppress the late wars the arc needs.

### Plagues and connectivity

There is no population in the model, so density has to be a proxy, and the honest one is **connectivity**: plagues spread by contact and trade, and a hegemon is the polity with the densest internal trade graph, so a plague in a large realm reaches all of it. `Structures` and `Works` are per-star and give built-up-ness where a capital is wanted.

It comes with a trade-off already in the data: `quarantine` is an iron scar, so surviving by closing up buys ossification.

### The conquest wave

The historical pattern is consistent enough to be a trigger set: a **marginal** people, holding little; a **latent advantage** that fragmentation had made unusable — which is why nomads and riders should be likeliest, though Macedon was neither; a **unification**, and the best available is that *the horde arises from the winner of a civil war*, which is literally both Temüjin and Philip and which closes the cycle on itself; a **windfall**; and a **rich, ossified target in reach that is believed to be hollow** — belief, so it runs through `Intel`.

The prime target is a *pair of exhausted great powers*, which is the Rashidun case and which ties the wave to the other mechanisms rather than leaving it standing apart.

It is a leader (`leaders.md`) in a state: dials hard to aggression and risk, expansion cheap, upkeep low because it lives on what it takes. **Low continuity is its fragility** — it cannot hold what it seizes, its provinces are alien and remember what was done to them, and it has no institution to inherit. That is the cost that low stiffness has never had.

And the coupling that is free: **terror makes the conquest cheap and the realm ungovernable.** Crimes accumulate into the monster reckoning, which cuts trade and raises everyone's will to fight. The method writes its own collapse into every telling in the galaxy.

### Later filters

Every ambient pressure in the model is ossification, so every empire dies the same way. What is missing is a filter whose trigger is **size and age** rather than tech — overextension, a hegemon's reckoning, a succession crisis of the realm rather than of a person. Whether the elders do anything to large empires today is unknown and measurable: facings by cause, against the size of the people facing them.

### Decline's stage 3

`decline.md`'s force lands here, and its gate is this proposal's: **the present holds under two thirds of the height in every seed and under half in the median; no seed has its height in the last third of its age; the still span is under a third of the age in the median; ages stay inside 20 to 80 Myr; nothing is capped.** The fertility floor kept at step 7 comes out when no seed in twenty falls through to it.

The hazard's decline term needs its **ceiling** raised with it or it is inert in half the runs.

## Implementation notes

- The concentration measures come first and are cheap: largest holder's share, top-three share, independent polities (free peoples, with vassals and slaves under their master), wars running, held over reach. The step-6 sampler (`Config.Sample`, `declineshape_test.go`) already records most of them. **That curve is what this proposal is a statement about and nobody has seen it yet.**
- Several of the plan's tuning candidates are war-shaped and should be folded into this proposal's tuning pass rather than done separately, since each otherwise costs a regeneration of its own.
- The war measures the plan owes — first wars per distinct pair, wars per thousand people-ticks, on twenty seeds — print here, before and after.

## Open questions

- **Does the arc happen at all today?** The concentration curve answers it, and how long one cycle takes — the one seed that completed one did it in about eight million years, which would fit three or four into an age.
- **Are peoples too long-lived for the cycle?** The median standing people is 14 to 53 Myr, one to several cycles, so the same peoples would be warlords twice over. That may be wanted or may mean the cycles must be faster.
- **Opportunity or will**, for the wars that do not happen.
- **Federations as a distinct stage**: consolidation by pact and vassalage rather than conquest, absorbing long-held vassals into one realm — a second road to empire with a different collapse, coming apart along its old seams.
- **How much of the force is separatism** and how much is the quiet going, and whether `SpentKyr` is needed at all.
- **Whether deathlessness becomes the master cause.** The Long Silence, Heaven and the tyrant's civil wars all hang off it; the non-immortality routes want tuning in the same pass so it is one cause among several.
