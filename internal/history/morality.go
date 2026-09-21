package history

import (
	"math/rand/v2"

	"worldgen/internal/flow"
	"worldgen/internal/species"
)

// Morality: what a people counts as wrong. Four kinds, rolled at birth from
// the people's nature and changed only by the drift events (a schism's
// branch, the Church, an uplift, a machine successor). Every place a people
// judges a fact reads sortFor, which overrides the fact's own sort by the
// judge's morality: an amoral people holds no monsters, a conquest people
// counts a taking a deed, a herd people barely counts an enslavement. A
// fixation also steers the direction and the trade.

// MoralKind is the shape of the good.
type MoralKind uint8

const (
	Amoral     MoralKind = iota // there is no word for wrong
	Individual                  // each one, in itself
	Herd                        // the many; the one is nothing
	Fixation                    // the Object, and nothing else
)

func (k MoralKind) String() string { return [...]string{"amoral", "individual", "herd", "fixation"}[k] }

// The objects of a fixation.
const (
	Conquest  = "conquest"
	Spawning  = "spawning"
	Knowing   = "knowing"
	OldThings = "old things"
	Holding   = "holding"
)

var objects = []string{Conquest, Spawning, Knowing, OldThings, Holding}

// Morality is a people's judgment.
type Morality struct {
	Kind   MoralKind
	Object string // the fixation's object; "" for the other kinds
}

// Word is the morality in a word or two, for the summary and the batch.
func (m Morality) Word() string {
	if m.Kind == Fixation {
		return "fixed on " + m.Object
	}
	return m.Kind.String()
}

// Portrait is the sentence the legends give it at birth.
func (m Morality) Portrait() string {
	switch m.Kind {
	case Amoral:
		return "They have no word for wrong."
	case Individual:
		return "They hold each life a good in itself."
	case Herd:
		return "For them the many are everything and the one is nothing."
	}
	switch m.Object {
	case Conquest:
		return "For them the only good is the taking of worlds."
	case Spawning:
		return "For them the only good is more of themselves."
	case Knowing:
		return "For them the only good is knowing."
	case OldThings:
		return "For them the only good is what the old ones left."
	}
	return "For them the only good is holding what they have."
}

// column is the morality's column in the override table.
func (m Morality) column() int {
	switch m.Kind {
	case Individual:
		return 0
	case Herd:
		return 1
	case Amoral:
		return 2
	}
	for i, o := range objects {
		if o == m.Object {
			return 3 + i
		}
	}
	return 7
}

// The roll. Weights start even-ish and each row of the tilt table that
// fits the species multiplies them; the fixation's object is rolled among
// the five with the tilts the same rows give it.

// moralHints is what the caller knows beyond the species: the Sight held
// from birth, a branch of a people that found much, a kind the roll is
// turned away from (a schism's branch), a multiplier on fixation and the
// object a fixation would be on (a machine successor).
type moralHints struct {
	Sight       bool
	FoundMuch   bool
	Against     MoralKind // 255 for none
	FixationMul float64   // 0 for none
	Object      string
}

var noHints = moralHints{Against: 255}

// moralTilt is one row of the table: when it fits, the kind weights are
// multiplied and the named objects tilted.
type moralTilt struct {
	when    func(sp *species.Species, h moralHints) bool
	kinds   [4]float64         // amoral, individual, herd, fixation; 1 leaves a weight alone
	objects map[string]float64 // fixation objects tilted
}

var moralBase = [4]float64{Amoral: 15, Individual: 30, Herd: 30, Fixation: 25}

var moralTilts = []moralTilt{
	{func(sp *species.Species, h moralHints) bool { return sp.Has("herd") || sp.Has("swarming") }, [4]float64{2, 0.04, 3, 1}, nil}, // the hive and the unconscious carry the same row in their profiles
	{func(sp *species.Species, h moralHints) bool { return sp.Has("individualist") || sp.Has("solitary") }, [4]float64{1, 3, 0.2, 1}, nil},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("caste") || sp.Has("collective") }, [4]float64{1, 1, 2, 1}, nil},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("eusocial") }, [4]float64{1, 1, 2, 1}, nil},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("conqueror") }, [4]float64{1, 1, 1, 3}, map[string]float64{Conquest: 3}},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("expansionist") || sp.Has("swarming") }, [4]float64{1, 1, 1, 3}, map[string]float64{Spawning: 3}},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("curious") || sp.Has("contemplative") }, [4]float64{1, 1, 1, 2}, map[string]float64{Knowing: 2}},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("curious") && h.Sight || h.FoundMuch }, [4]float64{1, 1, 1, 2}, map[string]float64{OldThings: 2}},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("pragmatic") }, [4]float64{1, 1, 1, 2}, map[string]float64{Holding: 2}},
	{func(sp *species.Species, h moralHints) bool { return sp.Has("pacifist") }, [4]float64{0, 2, 1, 1}, nil},
}

// rollMorality is the core: a morality from the species and the hints.
// The substrate and the shape tilt through the profile (a machine or a
// parasite toward the amoral and the fixed, a living world away from
// the individual and the herd); the unconscious are amoral, always.
func rollMorality(r *rand.Rand, sp *species.Species, h moralHints) Morality {
	p := sp.Profile()
	if p.Amoral {
		return Morality{Kind: Amoral}
	}
	kinds := moralBase
	for k := range kinds {
		kinds[k] *= p.Morals[k]
	}
	objs := map[string]float64{}
	for _, o := range objects {
		objs[o] = 1
	}
	for _, row := range moralTilts {
		if !row.when(sp, h) {
			continue
		}
		for k := range kinds {
			kinds[k] *= row.kinds[k]
		}
		for o, m := range row.objects {
			objs[o] *= m
		}
	}
	if h.FixationMul > 0 {
		kinds[Fixation] *= h.FixationMul
	}
	if h.Against < 4 {
		kinds[h.Against] *= 0.5 // the branch leans away from what it left
	}
	kind := MoralKind(draw(r, kinds[:]))
	if kind != Fixation {
		return Morality{Kind: kind}
	}
	if h.Object != "" {
		return Morality{Kind: Fixation, Object: h.Object}
	}
	ws := make([]float64, len(objects))
	for i, o := range objects {
		ws[i] = objs[o]
	}
	return Morality{Kind: Fixation, Object: objects[draw(r, ws)]}
}

// draw picks an index by weight; the last with weight when all are zero.
func draw(r *rand.Rand, ws []float64) int {
	total := 0.0
	for _, w := range ws {
		total += w
	}
	if total <= 0 {
		return len(ws) - 1
	}
	x := r.Float64() * total
	for i, w := range ws {
		x -= w
		if x < 0 {
			return i
		}
	}
	return len(ws) - 1
}

// The override table. A row is a fact kind; the columns are individual,
// herd, amoral, conquest, spawning, knowing, old things, holding. A cell
// of none is the fact held as a woe by its sufferer and as nothing by
// anyone else. Kinds not in the table keep the fact's own sort.

type judgment struct {
	Sort   Sort
	Weight float64
	None   bool
}

func deed(w float64) judgment  { return judgment{Sort: Deed, Weight: w} }
func crime(w float64) judgment { return judgment{Sort: Crime, Weight: w} }
func folly(w float64) judgment { return judgment{Sort: Folly, Weight: w} }

var none = judgment{None: true}

var moralTable = map[FactKind][8]judgment{
	FEnslaved:     {crime(4), crime(1), none, deed(2), crime(2), none, none, deed(1)},
	FBred:         {crime(4), none, none, none, deed(2), deed(1), none, none},
	FScoured:      {crime(5), crime(5), none, deed(2), crime(5), crime(3), crime(2), crime(3)},
	FHomeBroken:   {crime(5), crime(5), none, deed(2), crime(5), crime(3), crime(2), crime(3)},
	FBurned:       {crime(3), crime(3), none, deed(1), crime(4), crime(2), crime(3), crime(3)},
	FTaken:        {crime(2), crime(2), none, deed(2), crime(2), crime(1), crime(1), crime(2)},
	FWar:          {crime(2), crime(2), none, deed(2), crime(2), crime(1), crime(1), crime(2)},
	FBetrayal:     {crime(3), crime(4), none, folly(1), crime(2), crime(2), crime(2), crime(3)},
	FStripped:     {crime(3), crime(2), none, deed(1), crime(2), crime(1), crime(2), crime(4)},
	FManna:        {crime(2), none, none, none, folly(1), crime(1), none, none},
	FUnleashed:    {folly(4), folly(4), folly(2), folly(3), folly(4), crime(2), crime(4), folly(4)},
	FSealed:       {deed(2), deed(2), none, deed(1), deed(1), folly(2), folly(3), deed(2)},
	FMastered:     {deed(2), deed(2), deed(1), deed(1), deed(1), deed(4), deed(5), deed(3)},
	FFind:         {deed(2), deed(2), deed(1), deed(1), deed(1), deed(4), deed(5), deed(3)},
	FUplift:       {deed(3), deed(1), none, deed(1), deed(3), deed(3), deed(2), folly(1)},
	FFreed:        {deed(4), deed(1), none, folly(2), deed(2), deed(1), deed(1), folly(1)},
	FYield:        {deed(3), deed(2), deed(1), folly(3), deed(1), deed(2), deed(2), deed(2)},
	FSettle:       {deed(1), deed(2), deed(1), deed(1), deed(3), deed(1), deed(1), deed(2)},
	FStars:        {deed(1), deed(2), deed(1), deed(1), deed(3), deed(1), deed(1), deed(2)},
	FEmbargo:      {crime(1), crime(1), none, none, crime(2), none, none, crime(3)},
	FStrikeBought: {crime(1), crime(1), none, deed(1), crime(1), none, none, crime(1)},
	FBoughtOff:    {crime(2), crime(3), none, folly(1), crime(2), crime(1), crime(1), crime(2)},
	FSlight:       {crime(1), crime(1), none, none, crime(1), none, none, crime(3)},
	FPlagueGiven:  {crime(2), crime(2), none, none, crime(1), crime(1), none, crime(2)},
	FRefused:      {crime(1), crime(1), none, none, crime(1), none, none, crime(2)},
	FPoisoned:     {crime(4), crime(4), none, deed(1), crime(4), crime(3), crime(2), crime(4)},
}

// moralRows is the table as an array by fact kind, since sortFor runs
// per tale per people per tick.
var moralRows []struct {
	has bool
	row [8]judgment
}

func init() {
	for k, row := range moralTable {
		for int(k) >= len(moralRows) {
			moralRows = append(moralRows, struct {
				has bool
				row [8]judgment
			}{})
		}
		moralRows[k].has, moralRows[k].row = true, row
	}
}

// sortFor is a fact as one people judges it: the table's cell for its
// morality, else the fact's own sort. A cell of none is a woe of the
// fact's weight to the sufferer and nothing to anyone else. A woe is a woe
// under every morality: a morality decides what is done to others, not
// what hurts.
func sortFor(c *Civ, f *Fact) (Sort, float64) {
	if int(f.Kind) >= len(moralRows) || !moralRows[f.Kind].has {
		return f.sort(), f.weight()
	}
	j := moralRows[f.Kind].row[c.Morality.column()]
	if !j.None {
		return j.Sort, j.Weight
	}
	if f.Object == c.ID {
		return Woe, f.weight()
	}
	return Nothing, 0
}

// judges says whether a people's judgment of a fact differs from the
// fact's own sort.
func judges(c *Civ, f *Fact) bool {
	s, _ := sortFor(c, f)
	return s != f.sort()
}

// moralityDiff is what morality adds to how alien two peoples are to each
// other: a whole point between individual and herd, half to or from an
// amoral people, half between two fixations on different things.
func moralityDiff(a, b Morality) float64 {
	switch {
	case a.Kind == b.Kind && (a.Kind != Fixation || a.Object == b.Object):
		return 0
	case a.Kind == Amoral || b.Kind == Amoral:
		return 0.5
	case a.Kind == Fixation && b.Kind == Fixation:
		return 0.5
	case a.Kind == Fixation || b.Kind == Fixation:
		return 0.5
	}
	return 1 // individual against herd
}

// differs is how alien two peoples are: the blood, then the morality.
func (c *Civ) differs(e *Civ) float64 {
	d := difference(c.Species, e.Species) + moralityDiff(c.Morality, e.Morality) + driftGap(c, e)
	if e.Species.HasPower("mirror") {
		d = max(0, d-mirrorWis) // it answers in one's own voice; see eldritch.go
	}
	return d
}

// fixation is the category a fixation feeds first, if any: conquest arms,
// spawning and old things the road, knowing the mind, holding the works.
func (m Morality) fixation() (flow.Category, bool) {
	if m.Kind != Fixation {
		return 0, false
	}
	switch m.Object {
	case Conquest:
		return flow.Arms, true
	case Spawning, OldThings:
		return flow.Road, true
	case Knowing:
		return flow.Mind, true
	}
	return flow.Works, true
}

// fixed says whether a people is fixed on an object.
func (c *Civ) fixed(object string) bool {
	return c.Morality.Kind == Fixation && c.Morality.Object == object
}

// The drift. Each adapter sets the morality and says so.

// bornMorality rolls a people's morality at birth from its blood and the
// Sight, and logs the sentence.
func (w *World) bornMorality(c *Civ) {
	h := noHints
	h.Sight = c.Species.Miracle() == "foresight"
	c.Morality = rollMorality(w.R, c.Species, h)
	w.log("%s", c.Morality.Portrait())
}

// branchMorality is a schism's branch: it keeps the parent's judgment, or
// with a chance rolls again, leaning away from it; a branch of a people
// that found much leans to the old things.
func (w *World) branchMorality(nc, parent *Civ) {
	nc.Morality = parent.Morality
	if w.R.Float64() >= 0.3 {
		return
	}
	h := noHints
	h.Against = parent.Morality.Kind
	h.Sight = parent.miracle("foresight")
	h.FoundMuch = parent.Tally.FindSurvey+parent.Tally.FindSettle+parent.Tally.FindChance >= 3
	nc.Morality = rollMorality(w.R, nc.Species, h)
	if nc.Morality != parent.Morality {
		w.log("The %s have gone their own way in what they count as wrong. %s", nc.Name, nc.Morality.Portrait())
	}
}

// churchMorality is the Church winning: the people is turned to a
// fixation on knowing or on holding.
func (w *World) churchMorality(c *Civ) {
	m := Morality{Kind: Fixation, Object: Knowing}
	if w.R.Float64() < 0.5 {
		m.Object = Holding
	}
	if c.Morality == m {
		return
	}
	c.Morality = m
	w.log("The church of the %s teaches what is good, and it is one thing. %s", c.Name, m.Portrait())
}

// upliftMorality is an uplifted people taught its uplifter's judgment,
// half the time; else it keeps the one it was born with.
func (w *World) upliftMorality(nc, by *Civ) {
	if w.R.Float64() >= 0.5 || nc.Morality == by.Morality {
		return
	}
	nc.Morality = by.Morality
	w.log("The %s were taught what the %s call wrong. %s", nc.Name, by.Name, nc.Morality.Portrait())
}

// machineMorality is a machine successor: it leans hard to a fixation, on
// the thing its makers were doing when it outgrew them.
func (w *World) machineMorality(nc, makers *Civ) {
	h := noHints
	h.FixationMul = 3
	h.Object = w.doing(makers)
	nc.Morality = rollMorality(w.R, nc.Species, h)
	if nc.Morality.Kind == Fixation {
		w.log("What the %s hold good is what their makers were doing when they were outgrown. %s", nc.Name, nc.Morality.Portrait())
	} else {
		w.log("%s", nc.Morality.Portrait())
	}
}

// doing is what a people is about, as a fixation's object: at war,
// conquest; ships in flight, spawning; a remain lately found, old things;
// the works first in its direction, holding; else knowing.
func (w *World) doing(c *Civ) string {
	switch {
	case len(c.Wars) > 0:
		return Conquest
	case len(c.Voyages) > 0:
		return Spawning
	case len(c.Wielded) > 0:
		return OldThings
	case len(c.Order) > 0 && c.Order[0] == flow.Works:
		return Holding
	}
	return Knowing
}

// Judged counts, for the batch, the tales a people holds whose fact is a
// crime and its judgment is not, and the reverse.
func Judged(w *World, c *Civ) (excused, condemned int) {
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
		s, _ := sortFor(c, f)
		switch {
		case f.sort() == Crime && s != Crime && s != Woe:
			excused++
		case f.sort() != Crime && s == Crime:
			condemned++
		}
	}
	return
}

// SplitFacts counts, for the batch, the crimes known to two peoples or
// more, and of them the ones that are a crime to one and a deed to
// another.
func SplitFacts(w *World) (split, shared int) {
	holders := map[int][]int{}
	for _, c := range w.Civs {
		for _, t := range c.Lore {
			if !t.Forgot {
				holders[t.Fact] = append(holders[t.Fact], c.ID)
			}
		}
	}
	for _, f := range w.Facts {
		hs := holders[f.ID]
		if f.sort() != Crime || len(hs) < 2 {
			continue
		}
		shared++
		crime, deed := false, false
		for _, id := range hs {
			switch s, _ := sortFor(w.Civs[id], f); s {
			case Crime:
				crime = true
			case Deed:
				deed = true
			}
		}
		if crime && deed {
			split++
		}
	}
	return
}
