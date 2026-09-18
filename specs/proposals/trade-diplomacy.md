# Trade diplomacy: the slight of a war on a partner's partner

**Status:** Draft (raised 2026-09-18, after plan step 6; to be built in the AI layer after step 12, see `specs/plan.md` step 12b)
**Last updated:** 2026-09-18

Assumes [one-tick](one-tick.md) and [resources-and-trade](resources-and-trade.md): every rate below is per thousand years, and trade is the per-tick sending of surplus between partners.

## Problem

Step 6 fed the roads and the war count rose by about 210 per ten worlds (fed roads mean more reach, more fronts, more worlds taken). The sim tamed it with restrictions on the economy (the launch spare capped at the own-income surplus, goods feeding uses only, the trade caps), and the decision was taken to accept the rise as the economy's real behaviour and put the restraint where it belongs: in the mind. Today a war costs the attacker only what the target itself sent it (the appraisal's `Loss`). Nobody else minds. A people can make war on a partner of its own partner, of its ally, of the strongest people it knows, and none of them thinks worse of it, though every one of them loses goods when the target's spare goes to fleets and its roads are cut. There is nothing that makes a trading circle hang together.

## Scope

**In.** A slight: a war on a people's partner is a wrong done to that people, sized by the trade between them. The slight as a grudge, read by everything that reads grudges (the bar, the war will, the enemy of the day, the pact answer, regard). The council's appraisal reading the slights a war would cause in the peoples whose opinion the attacker cares about, against the prize. Soft blocs as what follows. Lifting the step-6 restrictions once this holds the count. A `techstats` reading of blocs.

**Out.** Formal alliances of trade (a pact of trade is a later contract term); sanctions as a decided act (an embargo is already a refusal that lasts, and a slighted people refuses by its grudge); the player's view of a bloc.

## Design

### The slight

When `a` declares war on `b`, every people `p` that trades with `b` (`p.Trade[b]`, `p != a`) takes a slight from `a`:

```
slight(p, b) = SlightWeight × (p.From[b].Total() + b.From[p].Total()) / max(p.Income.Total(), 1)
```

what the pair moves each tick, as a share of what `p` takes in; `SlightWeight` 2, so a partner that supplies a tenth of `p`'s income is a slight of 0.2. Added to `p.Grudge[a]`, capped at `SlightCap` 1 per war. A grudge is what a wrong is in this sim: the bar drops for the wronged, the war will rises, the enemy of the day changes, a pact offered by `a` is answered colder, `a`'s goods are refused above the grudge bar (0.3, so two slights of a fifth are an embargo of `a` by `p`). It decays as every grudge does (ossification, step 15, gives grudges their decay; until then it wears as today).

The slight repeats: each tick the war runs, `p` takes a further `slight × SlightTick` (0.1) while `b` is still sending it less than it did the tick before the war, so a long war on a supplier is a long wrong. A world taken from `b` that held a source `p` drew on (a shared grant or reach, `rarities`' `via`) is a slight of its own, `SlightSource` 0.3.

The fact: `FSlight` (Crime 1, "{S} made war on the partners of {O}") written once per war per slighted people whose slight passed `SlightFact` 0.2, so the tellings carry it and the monster reckoning counts it; under morality's table it is a crime of one to individual and herd, nothing to amoral, three to holding.

### The appraisal

Before a war, the council asks what the slights would cost. For each people `p` that trades with the target and that the attacker *cares about*, the attacker sums the slight `p` would take. It cares about `p` when any of: `a.Trade[p]` (a partner; weight 1), `allied(a, p)` (a pact; weight 1), `a.Intel[p].Mil ≥ a.Mil` and `p` in reach of `a`'s holdings (a strong neighbour; weight 0.5), `p` is `a`'s master or vassal (weight 1). A people it hates, or that is a monster to it, weighs 0.

```
Offence = Σ_p care(p) × slight(p, target) × (1 + a.From[p].Total() / max(a.Income.Total(), 1))
```

the second factor being what `a` itself stands to lose when `p` refuses it. `mind.Appraise` takes `Offence` beside `Loss` and `Prize`; the bar is `Bar − PrizeWeight × (Prize − Loss − OffenceWeight × Offence)`, `OffenceWeight` 5 (so an offence of 0.2 costs as much bar as a prize of 1 gains), clamped as now. A fearful attacker (`Fear > FearBar`) doubles the strong-neighbour weight; a conqueror halves the whole. `Why()` says "would slight the X (0.3) and the Y (0.1)".

### What follows

- **Blocs.** Peoples that trade with each other, and with each other's partners, become expensive to attack from inside and cheap from outside: a war on a bloc member by an outsider slights the whole bloc, and the bloc's grudges make it refuse the outsider's goods and answer its pacts cold, so the outsider is pushed to trade elsewhere, which is another bloc. Blocs are not declared; they are the transitive closure of `Trade` as seen through the slights, and `techstats` reads them as the connected components of the partner graph at each people's zenith, with the wars inside a component against the wars across.
- **The restrictions lifted.** With the offence in the bar, the step-6 caps are the wrong tool: the launch spare cap (`direct`'s second `flow.Direct`) and the uses-only rule for goods go, and the trade caps go up to `CapBase` 0.5, `CapFast` 0.75. Each is one batch, compared against this proposal's own baseline, and the war count is read as the bloc report reads it: wars across blocs may rise, wars inside must fall.
- **A wronged partner's war.** A slighted people with a grudge past the grudge bar and a strong bar already treats the attacker as a grudge target; nothing new. What is new is that it got there without being attacked itself.

### Numbers

All in `mind.Tuning.Slight`: `Weight` 2, `Cap` 1, `Tick` 0.1, `Source` 0.3, `Fact` 0.2, `OffenceWeight` 5, `StrongWeight` 0.5. First guesses; the batch sets them.

## Implementation notes

One stage, in the mind and `war.go`: `mind.Slight(SlightInput)` (a pure size), `mind.Appraise` gains `Offence`, `history/slight.go` (the adapter at `declare`, the per-tick step in `wartime`, the source slight in `loseSystem`, `FSlight`, the telling line), `council.go` builds the offence input, `techstats` the bloc report. Then the lifting, one batch each. Placed after contracts (step 12) so that a bought guard and a sold sighting are in the graph before blocs are read, and so that the term registry can carry a `bloc` term later if a declared bloc is ever wanted.

## Open questions

- Should the slight read the partner's *dependence* (`p.Dependent[b]`) rather than the flow, so that a war on a supplier a people cannot do without is the deep wrong and a war on a partner it barely uses is nothing? Proposed: the flow, with dependence as a multiplier of 2.
- Does a people slight itself by attacking its own partner's partner when it is the stronger side and expects to take the supplier's worlds and the trade with them? Proposed: no; the prize already counts the worlds.
- Whether blocs should be visible to the player as named things, or stay a reading. Proposed: a reading, until the player layer wants a name.
