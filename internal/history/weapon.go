package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/plague"
	"worldgen/internal/tech"
)

// Engineered plagues: a people that knows the craft may make a plague
// instead of waiting for one. The same object, with its shape chosen
// within the band the craft allows and its maker immune above the crudest
// rung. The council makes one when it has a reason (mind.MakeWeapon), the
// programme is a use under arms, and each tick the road to the target is
// open the maker tries: a plague of the body sneaked into goods, one of
// the mind into a message, one roll. Taking or being caught is the same
// crime, above every bar. Every craft node is a filter on discovery, and
// a held weapon leaks.

// Weapon is a plague made and held.
type Weapon struct {
	Plague int
	Target int // the people it was made for; -1 once that people is gone or barred
	Node   string
	Made   Year
}

// ScarVial is the horror of the vial: the craft above is closed and what
// is held is never used.
const ScarVial = "a horror of the vial"

func init() {
	def(&Filter{
		Key: "containment", Name: "the Vial", Levels: []string{"sur"}, Diff: 4, Repeat: true, Domain: "biology",
		Adjust: func(w *World, c *Civ) (levels []string, diff float64, domain string) {
			n := w.craftLearned(c)
			if n == nil {
				return nil, 0, ""
			}
			if n.Weapon.Memetic {
				return []string{"soc"}, float64(n.Weapon.Rung), "society"
			}
			return []string{"sur"}, float64(n.Weapon.Rung), "biology"
		},
		Overcome: func(w *World, c *Civ) {}, // most keep it in the vial; the legends note the ones that did not
		Scar: func(w *World, c *Civ) {
			if !c.Scars[ScarVial] {
				c.Scars[ScarVial] = true
				w.log("Something gets out of the vial among the %s and is burned with the building. They keep the craft and never use it, and go no further with it.", c.Name)
			}
		},
		Decline: func(w *World, c *Civ) {
			if n := w.craftLearned(c); n != nil {
				w.breakout(c, n, nil)
			}
		},
	})
}

// craftLearned is the plague-making node a people learned last, which is
// the one the vial is about.
func (w *World) craftLearned(c *Civ) *tech.Node {
	var best *tech.Node
	for _, n := range tech.Crafts {
		if c.Known[n.Key] && (best == nil || c.Learned[n.Key] >= c.Learned[best.Key]) {
			best = n
		}
	}
	return best
}

// craft is the best working plague-making node a people has of a kind,
// or nil; none at all under the vial's scar.
func (w *World) craft(c *Civ, memetic bool) *tech.Node {
	if c.Scars[ScarVial] {
		return nil
	}
	var best *tech.Node
	for _, n := range tech.Crafts {
		if n.Weapon.Memetic == memetic && c.Known[n.Key] && c.working(n.Key) && (best == nil || n.Weapon.Rung > best.Weapon.Rung) {
			best = n
		}
	}
	return best
}

// armPlagues is the council's pass over the peoples it could make a plague
// for: one held weapon per craft, made for the first with a reason and a
// road; a held one is turned on a new target when its own is gone, and
// burned when nobody gives a reason any more.
func (w *World) armPlagues(c *Civ) {
	if !c.Active() || !c.Free() {
		return
	}
	for _, memetic := range []bool{false, true} {
		n := w.craft(c, memetic)
		if n == nil {
			continue
		}
		wp := c.Weapons[n.Key]
		if wp != nil && wp.Target >= 0 {
			continue
		}
		e, aim := w.plagueTarget(c, memetic, true)
		switch {
		case wp == nil && e != nil:
			w.makePlague(c, e, n, aim)
		case wp != nil && e != nil:
			wp.Target = e.ID
		case wp != nil:
			if any, _ := w.plagueTarget(c, memetic, false); any == nil {
				w.log("The %s burn %s, having nobody left to give it to.", c.Name, w.Plagues[wp.Plague].Name)
				delete(c.Weapons, n.Key)
			}
		}
	}
}

// plagueTarget is the first people the council has a reason to make a
// plague for, with the aim the reason sets; with road, only one it has a
// road to.
func (w *World) plagueTarget(c *Civ, memetic, road bool) (*Civ, plague.Aim) {
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if !e.Active() || e.Own >= 0 || w.allied(c, e) || c.Truce[eid] > w.Now || e.Master == c.ID {
			continue
		}
		if !w.bears(e, plagueKind(memetic)) || (road && !w.roadTo(c, e, memetic)) {
			continue
		}
		if v := w.weaponReason(c, e); v.Make {
			return e, v.Aim
		}
	}
	return nil, 0
}

func plagueKind(memetic bool) plague.Kind {
	if memetic {
		return plague.Memetic
	}
	return plague.Biological
}

// weaponReason is the council on one people as a target.
func (w *World) weaponReason(c, e *Civ) mind.Weapon {
	bar, wants, _ := w.bar(c, e)
	in := mind.WeaponInput{Posture: c.posture(), Hates: c.hates(e), AtWar: c.Wars[e.ID], Grudge: c.Grudge[e.ID], Vengeful: c.posture() == mind.Vengeful}
	if wants {
		in.Bar = bar
		if !c.Wars[e.ID] && e.Free() && mind.Hostile(c.posture()) {
			ap := w.appraise(c, e, -1)
			in.Worth = ap.Prize > 0 && ap.Acted < bar
		}
	}
	v := mind.MakeWeapon(in)
	w.explain(c, "a plague for the "+e.Name, v)
	return v
}

// makePlague is the making: the shape by the aim within the craft's band,
// tailored to the target's blood where the craft allows, named for its
// maker, held from the next tick.
func (w *World) makePlague(c, e *Civ, n *tech.Node, aim plague.Aim) *Plague {
	t := &w.Cfg.Tuning.Plague
	wd := n.Weapon
	con, let := plague.Shape(aim, wd.Band, t)
	pl := plague.Make(plagueKind(wd.Memetic), con, let, wd.Conscious && w.R.Float64() < t.Conscious)
	if wd.Tailored {
		pl.Band = e.Species.ID
	}
	pl.Name = w.weaponName(c, wd.Memetic)
	p := &Plague{Plague: pl, ID: len(w.Plagues), Born: w.Now, FirstHost: -1, Cause: "made", Maker: c.ID, Made: true, Rider: -1}
	w.Plagues = append(w.Plagues, p)
	c.Weapons[n.Key] = &Weapon{Plague: p.ID, Target: e.ID, Node: n.Key, Made: w.Now}
	if wd.Immune {
		c.Immune[p.ID] = true // a thing designed has a designed cure
	}
	if wd.Memetic {
		w.log("The %s shape an idea to break the %s, and call it %s among themselves.", c.Name, e.Name, p.Name)
	} else {
		w.log("The %s breed a sickness for the %s, and call it %s among themselves.", c.Name, e.Name, p.Name)
	}
	if pl.Conscious {
		w.log("They have made it to think.")
	}
	return p
}

// weaponName is what the galaxy will call a made plague: the maker's
// gift, or its lie.
func (w *World) weaponName(c *Civ, memetic bool) string {
	word := "Gift"
	if memetic {
		word = "Lie"
	}
	n := 1
	for _, p := range w.Plagues {
		if p.Maker == c.ID && (p.Kind == plague.Memetic) == memetic {
			n++
		}
	}
	if n == 1 {
		return "the " + c.Name + " " + word
	}
	return "the " + ordinal(n) + " " + c.Name + " " + word
}

// weaponUses is the programmes that keep the held plagues: a use under
// arms at the craft's price.
func (w *World) weaponUses(c *Civ) []flow.Use {
	var out []flow.Use
	for _, k := range sortedKeys(c.Weapons) {
		n := tech.Get(k)
		p := w.Plagues[c.Weapons[k].Plague]
		out = append(out, flow.Use{Key: "weapon:" + k, Name: "the programme that keeps " + p.Name, Cat: flow.Arms, Era: n.Era, Need: w.needOf(c, n)})
	}
	return out
}

// useWeapons is a maker's tick: the leak on every held weapon, and the
// attempt where the road to the target is open. Under the vial's scar
// nothing is tried.
func (w *World) useWeapons(c *Civ) {
	t := &w.Cfg.Tuning.Plague
	for _, k := range sortedKeys(c.Weapons) {
		if !c.Active() {
			return
		}
		wp := c.Weapons[k]
		n := tech.Get(k)
		p := w.Plagues[wp.Plague]
		if wp.Made == w.Now {
			continue // a programme takes a tick
		}
		f := w.factors(c, p.Kind)
		if w.chance(plague.LeakChance(plague.Dirt(p.Kind, f, t), c.Shed["weapon:"+k], t)) {
			c.Tally.Leaks++
			w.log("The programme that keeps %s is not kept well enough.", p.Name)
			w.breakout(c, n, p)
			continue
		}
		if c.Scars[ScarVial] || wp.Target < 0 {
			continue
		}
		e := w.Civs[wp.Target]
		if !e.Active() || w.allied(c, e) {
			wp.Target = -1
			continue
		}
		road := "poison"
		if p.Kind == plague.Memetic {
			road = "whisper"
		}
		if !w.roadTo(c, e, p.Kind == plague.Memetic) || (p.Kind == plague.Biological && c.From[e.ID].Total()+e.From[c.ID].Total() == 0) {
			continue // nothing moving to hide it in
		}
		if w.attempt(c, e, p, road) {
			delete(c.Weapons, k) // what is loose is no longer held
		}
	}
}

// roadTo says whether a maker has a road to a target a plague could ride:
// a trade link for one of the body, a message the target reads for one of
// the mind.
func (w *World) roadTo(c, e *Civ, memetic bool) bool {
	if memetic {
		return e.Fathomed[c.ID] && !e.Closed[c.ID] && !e.Barred[c.ID] && w.hear(c, e)
	}
	return c.Trade[e.ID]
}

// kin says whether a people is of the blood a plague was tailored to:
// the species itself, or one branched from it.
func kin(c *Civ, band int) bool {
	if band < 0 {
		return true
	}
	for sp := c.Species; sp != nil; sp = sp.Parent {
		if sp.ID == band {
			return true
		}
	}
	return false
}

// attempt is one try to put a plague in a people by stealth: one roll at
// the contagion by the target's hygiene; caught either way is the crime.
// Returns whether it took.
func (w *World) attempt(c, e *Civ, p *Plague, roadKey string) bool {
	t := &w.Cfg.Tuning.Plague
	if !w.catchable(e, p) {
		return false
	}
	c.Tally.Attempts++
	n, _ := w.rungs(e, p.Kind)
	if w.chance(plague.CatchChance(p.Contagion, 1, plague.Hygiene(n, t), natureMul(e, p.Kind))) {
		w.infect(e, p, c, roadKey)
		w.poisoned(c, e, p, true)
		return true
	}
	censor := p.Kind == plague.Memetic && e.Known["censorship"] && e.working("censorship")
	if w.chance(plague.DetectChance(n, censor, t)) {
		c.Tally.Detected++
		w.poisoned(c, e, p, false)
	}
	return false
}

// poisoned is the crime of an attempt that took or was caught: the fact,
// the target's doors shut to the maker for good, the betrayal, and a war
// cause above every bar, with the target's allies called.
func (w *World) poisoned(c, e *Civ, p *Plague, took bool) {
	if took {
		c.Tally.Poisoned++
		p.Poisonings++
	}
	e.Barred[c.ID] = true
	w.factOf(FPoisoned, c, e, e.Home, p.Name)
	if took {
		w.log("The %s find %s was hidden in what the %s sent them, and made for them.", e.Name, p.Name, c.Name)
	} else {
		w.log("The %s catch the %s trying to hide %s in what they sent. They take nothing from them again.", e.Name, c.Name, p.Name)
	}
	if w.allied(c, e) {
		w.breakPacts(c, e)
	}
	w.betray(c, e, "poisoned the "+e.Name, 1)
	e.Grudge[c.ID] = max(e.Grudge[c.ID], 3)
	if e.Free() && !e.Wars[c.ID] && !e.Has("pacifist") {
		w.declare(e, c, "the poisoning")
	}
}

// breakout is the thing getting out: at discovery a plague of the craft's
// band with no shape chosen, from a held programme the plague it kept.
// The maker is host 0, and immune only if the craft grants it, in which
// case it carries what it cannot catch. A plague that thinks and gets out
// on its maker is a story told once per galaxy.
func (w *World) breakout(c *Civ, n *tech.Node, p *Plague) {
	t := &w.Cfg.Tuning.Plague
	wd := n.Weapon
	if p == nil {
		pl := plague.Loose(w.R, plagueKind(wd.Memetic), wd.Band, c.Name, t)
		p = &Plague{Plague: pl, ID: len(w.Plagues), Born: w.Now, FirstHost: -1, Cause: "breakout", Maker: c.ID, Rider: -1}
		w.Plagues = append(w.Plagues, p)
	} else {
		delete(c.Weapons, n.Key)
		p.Cause = "breakout"
	}
	c.Tally.Breakouts++
	if p.Conscious {
		if w.wokeOnMaker {
			p.Conscious = false
		} else {
			w.wokeOnMaker = true
		}
	}
	if !w.bears(c, p.Kind) {
		w.log("It gets out, and finds nothing in the %s to be in.", c.Name)
		return
	}
	inf := w.infect(c, p, nil, "loose")
	w.factOf(FHorrorMade, c, nil, c.Home, p.Name)
	if wd.Immune {
		inf.Carrier = true
		c.Immune[p.ID] = true
		w.log("It gets out. %s does nothing to the %s, who made it; it goes with everything they send.", upper(p.Name), c.Name)
		return
	}
	w.log("It gets out. %s is loose among the %s, who made it.", upper(p.Name), c.Name)
}
