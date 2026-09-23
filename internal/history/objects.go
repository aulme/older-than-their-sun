package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/species"
)

// The Ember and the Manna: two miracles that make an object each, a mobile
// source with a swarm's worth of yield, a rolled form and the form's
// price, since a departure from physics has a literal consequence. Losing
// the object does not lose the miracle; the people makes another after a
// million years. Both are found as elder artifacts and taken by force like
// any mobile rarity. The Manna copies: a cutting given to a partner is an
// instance of its own, so it spreads by gift along the trade graph. A
// sentient Manna is a crime to whoever judges it so, may rise as a people,
// and any Manna may get loose.

const (
	emberYield      = 8.0       // energy from the Ember
	mannaYield      = 6.0       // organic matter from the Manna
	fungusMetal     = 1.0       // and metal from the fungus that eats stone
	pocketDim       = 2.0       // what a pocket star loses each time it is moved
	singularityFree = 0.3       // chance a captive singularity gets loose when its star is taken
	tapThrough      = 0.01      // per kyr, something comes through a tap into a hotter place
	tapGlare        = -0.5      // the glare's cost in survival
	holeGrow        = 0.01      // energy per kyr a hole in the vacuum grows by
	holeDoom        = 16.0      // at this much it is a doom for the star
	mannaSentient   = 0.2       // chance a Manna of any form but the last thinks
	mannaRise       = 0.001     // per kyr, a sentient Manna rises as a people
	mannaLoose      = 0.0005    // per kyr, a Manna gets out; doubled in the holder's dark age
	cuttingChance   = 0.1       // per kyr, a cutting goes to a partner in want
	remakeYears     = 1_000_000 // before a lost object is made again
	objectLeap      = 0.01      // the leap weight of an object's miracle against the six: they are rare
	objectShare     = 0.02      // the share of elder miracle artifacts that are an Ember or a Manna
	throughDiff     = 5.5       // what it takes to burn off what comes through
	looseRadius     = 2.0       // light years the Manna loose takes with it
	singularityBurn = 4.0       // light years a loosed singularity blasts
)

// objectKeys are the miracles that make an object, in a fixed order.
var objectKeys = []string{"ember", "manna"}

// objectForm is one form an object can take, and what it does: a hook per
// tick while held, one when it changes hands, one when its star is taken
// by force. A hook that is nil does nothing.
type objectForm struct {
	Key    string
	Weight float64 // from data/miracles.json, with the yield and the levels
	Yield  flow.Income
	Levels [3]float64
	Tick   func(w *World, c *Civ, s *Source)
	Moved  func(w *World, c *Civ, s *Source)
	Taken  func(w *World, c, e *Civ, s *Source, star int) bool // true when the object is gone
}

var emberForms = []objectForm{
	{Key: "pocket_star",
		Moved: func(w *World, c *Civ, s *Source) {
			s.Yield[flow.E] = max(0, s.Yield[flow.E]-pocketDim)
			if c != nil {
				w.event(KPocketDimmed, c, nil, -1, P{"source": s.ID, "gives": s.Yield[flow.E]})
			}
		}},
	{Key: "singularity",
		Taken: func(w *World, c, e *Civ, s *Source, star int) bool {
			if w.R.Float64() >= singularityFree {
				return false
			}
			w.lose(s, e, "loose")
			w.loose = append(w.loose, star)
			w.event(KSingularityLoose, nil, nil, star, P{"source": s.ID})
			return true
		}},
	{Key: "hot_tap",
		Tick: func(w *World, c *Civ, s *Source) {
			if s.Star >= 0 && w.chance(tapThrough) {
				w.through(c, s)
			}
		}},
	{Key: "vacuum_hole",
		Tick: func(w *World, c *Civ, s *Source) {
			s.Yield[flow.E] += holeGrow * w.dt
			if s.Yield[flow.E] >= holeDoom && s.Star >= 0 {
				w.doom(c, s)
			}
		}},
}

var mannaForms = []objectForm{
	{Key: "mould"},
	{Key: "mat"},
	{Key: "brood"},
	{Key: "kelp"},
	{Key: "fungus"},
}

// objectForms is the forms by miracle, filled at init so the hooks may
// refer to the walks that read it.
var objectForms = map[string][]objectForm{}

func init() {
	objectForms["ember"] = emberForms
	objectForms["manna"] = mannaForms
	for _, key := range objectKeys {
		m := tables.miracleByKey[key]
		if m == nil || len(m.Forms) != len(objectForms[key]) {
			panic("history: the forms of " + key + " are not the file's")
		}
		for i := range objectForms[key] {
			f, d := &objectForms[key][i], m.Forms[i]
			if f.Key != d.Key {
				panic("history: the form " + f.Key + " of " + key + " is not the file's " + d.Key)
			}
			f.Weight, f.Yield, f.Levels = d.Weight, d.Yield, d.Levels
		}
	}
}

// formOf is the form of an object, or nil.
func formOf(s *Source) *objectForm {
	for i := range objectForms[s.Key] {
		if objectForms[s.Key][i].Key == s.Form {
			return &objectForms[s.Key][i]
		}
	}
	return nil
}

// objects is the civ step: a people that holds a miracle of an object and
// no object makes one, at once or a million years after losing the last;
// then each object it holds does what its form does.
func (w *World) objects(c *Civ) {
	for _, key := range objectKeys {
		if !c.miracle(key) || (key == "manna" && c.Has("kinfed")) || w.holdsObject(c, key) {
			continue
		}
		if since, ok := c.Remade[key]; ok && w.Now-since < remakeYears {
			continue
		}
		how := "made"
		if c.Miracles[key] == "born" && c.Remade[key] == 0 {
			how = "born"
		}
		w.makeObject(c, key, nil, nil, how)
	}
	for _, id := range w.mobile {
		s := w.Sources[id]
		if s.Form == "" || s.Holder != c.ID {
			continue
		}
		if f := formOf(s); f != nil && f.Tick != nil {
			f.Tick(w, c, s)
		}
		if !c.Active() || s.Holder != c.ID {
			continue
		}
		if s.Key == "manna" {
			w.mannaStep(c, s)
		}
	}
}

// tickObjects is the phase after the fleets: what got loose in a taking
// this tick does its harm now, once the taking is done.
func (w *World) tickObjects() {
	for _, star := range w.loose {
		w.blast(star, singularityBurn, "singularity", nil, 1, -1)
	}
	w.loose = nil
}

// holdsObject says whether a people holds an object of a miracle.
func (w *World) holdsObject(c *Civ, key string) bool {
	for _, id := range w.mobile {
		if s := w.Sources[id]; s.Key == key && s.Holder == c.ID && s.Form != "" {
			return true
		}
	}
	return false
}

// makeObject makes the object of a miracle for a people: rolls its form
// and, for the Manna, whether it thinks, unless a parent is given, when
// it is a cutting of the parent's form. An elder artifact given as l is
// the object; else a legacy is written for it, so it can be found when it
// is left behind and taken when its star is. How says the line: made,
// born, found or cutting.
func (w *World) makeObject(c *Civ, key string, l *Legacy, parent *Source, how string) *Source {
	forms := objectForms[key]
	var f *objectForm
	sentient := false
	if parent != nil {
		f, sentient = formOf(parent), parent.Sentient
	} else {
		total := 0.0
		for i := range forms {
			total += forms[i].Weight
		}
		x := w.R.Float64() * total
		for i := range forms {
			x -= forms[i].Weight
			if x < 0 {
				f = &forms[i]
				break
			}
		}
		if f == nil {
			f = &forms[len(forms)-1]
		}
		if key == "manna" {
			sentient = w.R.Float64() < mannaSentient
		}
	}
	s := w.addSource(&Source{Key: key, Kind: MadeSource, Star: c.Home, Mobile: true, Rarity: true,
		Yield: f.Yield, Levels: f.Levels, Holder: c.ID, Carried: -1, Legacy: -1,
		Form: f.Key, Sentient: sentient, Maker: c.ID, Made: w.Now})
	if l == nil {
		l = &Legacy{ID: len(w.Legacies), Age: -1, Maker: c.ID, Kind: Artifact, Star: c.Home, Node: key, Portrait: f.Key,
			State: Wielded, People: -1, Finder: c.ID, Level: "miracle", Cond: Abandoned, Source: s.ID, Plague: -1}
		w.addLegacy(l)
	} else {
		l.Source = s.ID
		l.State, l.Finder, l.Level = Wielded, c.ID, "miracle"
		if l.Elder != nil {
			s.Kind, s.Maker = ElderSource, -1
		}
	}
	s.Legacy = l.ID
	held := false
	for _, x := range c.Wielded {
		if x == l {
			held = true
		}
	}
	if !held {
		c.Wielded = append(c.Wielded, l)
	}
	if c.Aloft {
		if best := w.greatestFleet(c); best != nil {
			s.Carried, s.Star = best.ID, -1
		}
	}
	switch how {
	case "cutting":
		// the giver's line says it
	default:
		w.event(KObjectMade, c, nil, -1, P{"object": key, "how": how, "form": f.Key, "source": s.ID})
	}
	if sentient && key == "manna" {
		w.fact(FManna, c, nil, c.Home).with(P{"way": "eat"})
	}
	w.recompute(c)
	return s
}

// wieldObject is a found object put to use: an elder's, which becomes an
// object now, or a made one, which changes hands.
func (w *World) wieldObject(c *Civ, l *Legacy) {
	if l.Source < 0 {
		w.makeObject(c, l.Node, l, nil, "found")
		return
	}
	s := w.Sources[l.Source]
	w.transfer(s, nil, c)
	s.Star, s.Carried = c.Home, -1
	w.event(KObjectMade, c, nil, -1, P{"object": l.Node, "how": "wield", "form": s.Form, "source": s.ID})
	if s.Sentient {
		w.fact(FManna, c, nil, c.Home).with(P{"way": "wield"})
	}
}

// unleashObject is an object's artifact opened by somebody who could not
// hold it: an Ember burns once and takes the star with it; a Manna gets
// out and eats. What it was is lost.
func (w *World) unleashObject(c *Civ, l *Legacy) {
	if l.Source >= 0 {
		w.lose(w.Sources[l.Source], nil, "loose")
	}
	if l.Node == "ember" {
		w.blast(l.Star, singularityBurn, "ember", nil, 1, l.ID)
		return
	}
	w.blast(l.Star, looseRadius, "manna", nil, 1, l.ID)
}

// moved is an object changing hands or ships: its form's price.
func (w *World) moved(c *Civ, s *Source) {
	if f := formOf(s); f != nil && f.Moved != nil {
		f.Moved(w, c, s)
	}
}

// taken is an object's star taken by force by c from e: its form's
// price; true when the object did not survive it.
func (w *World) taken(c, e *Civ, s *Source, star int) bool {
	if f := formOf(s); f != nil && f.Taken != nil {
		return f.Taken(w, c, e, s, star)
	}
	return false
}

// lose is an object ending: it leaves its holder, yields to nobody, and
// its remain is lost.
func (w *World) lose(s *Source, from *Civ, fate string) {
	w.transfer(s, from, nil)
	s.Star, s.Carried, s.Fate = -1, -1, fate
	if s.Legacy >= 0 {
		w.setState(w.Legacies[s.Legacy], Lost)
	}
}

// through is something coming through a tap into a hotter place: burned
// off, or the star's worlds are lost to it and the Ember lies there.
func (w *World) through(c *Civ, s *Source) {
	star := s.Star
	if c.Mil+w.R.NormFloat64()*1.5 >= throughDiff+c.traitDiff("find") {
		w.event(KThrough, c, nil, star, P{"burned": true, "source": s.ID})
		return
	}
	w.event(KThrough, c, nil, star, P{"burned": false, "source": s.ID})
	w.bury(s, c, star)
	s.Fate = "through"
	if contains(c.Systems, star) {
		w.loseSystem(c, star, "burned", because("ember_eaten"))
	}
}

// doom is a hole in the vacuum grown past holding: its star begins to go
// out. The home star fails as any failing sun does; another world is
// lost now. The tellings say who opened it.
func (w *World) doom(c *Civ, s *Source) {
	star := s.Star
	w.fact(FVacuumHole, c, nil, star).with(P{"source": s.ID})
	w.lose(s, c, "doom")
	w.G.Stars[star].Failing = true
	if star != c.Home && contains(c.Systems, star) {
		w.loseSystem(c, star, "burned", reason{})
	}
}

// mannaStep is what a Manna may do each tick: rise, if it thinks; get
// loose, twice as likely while its holder is in a dark age.
func (w *World) mannaStep(c *Civ, s *Source) {
	if s.Sentient && s.Star >= 0 && w.chance(mannaRise) {
		w.rise(c, s)
		return
	}
	p := mannaLoose
	if float64(w.Now-c.LastDark) < w.Cfg.Tuning.Appraise.DarkAge {
		p *= 2
	}
	if s.Star >= 0 && w.chance(p) {
		w.getLoose(c, s)
	}
}

// rise is a sentient Manna becoming a people at its star, by the uplift
// path: a new blood grown for the table, holding that world as a vassal
// of the eaters, whose yield from it ends at once. The eaters keep their
// seat: a Manna at the only world does not rise.
func (w *World) rise(c *Civ, s *Source) {
	star := s.Star
	if !contains(c.Systems, star) || len(c.Systems) < 2 {
		return
	}
	sp := species.Generate(w.R, w.G.Stars[star].Mult)
	sp.Add("table")
	sp.Made = species.MadeBy("table", c.ID)
	w.lose(s, c, "rose")
	w.loseSystem(c, star, "risen", reason{})
	nc := w.spawnCiv(star, sp, c.ID)
	w.setMaster(nc, nc.Master, true) // through the ledger, so the vassalage is in the chronicle
	nc.Seen = c.Declines
	for _, k := range knownOf(c) {
		if w.R.Float64() < 0.5 {
			w.know(nc, k)
		}
	}
	w.forget(nc, 0.3)
	w.recompute(nc)
	w.fact(FRise, nc, c, star).with(P{"traits": sp.TraitKeys()})
	w.inherit(nc, c, 1)
}

// getLoose is a Manna getting out: it yields to nobody, it eats what is
// near, and what is left of it at that star is a people that makes more
// of itself out of what it finds; its remain is lost, since it is the
// people now.
func (w *World) getLoose(c *Civ, s *Source) {
	star := s.Star
	w.lose(s, c, "loose")
	w.fact(FLoose, c, nil, star)
	w.blast(star, looseRadius, "manna_grown", c, 1, -1)
	if nc := w.replicatorAt(star, species.Biological, species.MadeBy("manna_loose", c.ID), false); nc != nil {
		nc.Species.Parent = c.Species
		w.event(KNamedItself, nc, nil, star, P{"way": "growth", "traits": nc.Species.TraitKeys()})
	}
}

// cutting is a Manna given: a partner in want of organic matter and
// holding none gets an instance of the giver's form, and the giver keeps
// its own.
func (w *World) cutting(a, b *Civ) {
	if b.OwnWant[flow.O] <= 0 || b.Has("kinfed") {
		return
	}
	for _, id := range w.mobile {
		s := w.Sources[id]
		if s.Key != "manna" || s.Holder != a.ID || s.Form == "" {
			continue
		}
		if w.holdsObject(b, "manna") || !w.chance(cuttingChance) {
			return
		}
		s.Given++
		w.event(KCutting, a, b, -1, P{"source": s.ID})
		w.makeObject(b, "manna", nil, s, "cutting")
		return
	}
}
