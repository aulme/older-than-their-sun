package history

import "worldgen/internal/tech"

// research rolls discoveries for the tick. The domain is drawn from weights
// that are species tilt times current focus, then an available node in that
// domain is taken. Discovering a node can fire its filter.
func (w *World) research(c *Civ) {
	rate := 0.12 * c.Species.Rate() * (1 + 0.08*c.Soc) * (1 + 0.03*float64(len(c.Systems))) / (1 + 0.6*float64(c.Era))
	rate *= c.rateMul(w)
	n := min(w.count(rate), 8)
	for i := 0; i < n && c.Active(); i++ {
		w.discover(c)
	}
}

func (c *Civ) rateMul(w *World) float64 {
	m := 1.0
	if c.Scars[ScarNoMachines] {
		m *= 0.8
	}
	if c.Scars[ScarOssified] {
		m *= 0.7
	}
	if c.Scars[ScarLeftBehind] {
		m *= 0.1
	}
	if c.Boons[BoonAligned] {
		m *= 1.3
	}
	if c.Structures["dyson"] > 0 {
		m *= tech.Structures["dyson"].Rate
	}
	for p := range c.Trade {
		if w.Civs[p].Living() {
			m *= 1.15
			break
		}
	}
	if !c.Free() {
		if c.Vassal {
			m *= 0.7
		} else {
			m *= 0.3
		}
	}
	return m
}

func (w *World) discover(c *Civ) {
	var avail []*tech.Node
	var weights []float64
	total := 0.0
	for _, n := range tech.Nodes {
		if c.Known[n.Key] || (c.Locked[n.Domain] && n.Key != c.Species.World.Unlock) {
			continue
		}
		if n.Patience > 0 && float64(w.Now-c.Born)/1000 < n.Patience {
			continue
		}
		ok := true
		for _, p := range n.Prereqs {
			if !c.Known[p] {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		f := c.Focus[n.Domain]
		if f == 0 {
			f = 1
		}
		wt := n.Weight * c.Species.DomainMul(n.Domain) * f
		if c.Locked[n.Domain] {
			wt = n.Weight
		}
		avail = append(avail, n)
		weights = append(weights, wt)
		total += wt
	}
	if len(avail) == 0 {
		return
	}
	x := w.R.Float64() * total
	var n *tech.Node
	for i, a := range avail {
		x -= weights[i]
		if x < 0 {
			n = a
			break
		}
	}
	if n == nil {
		n = avail[len(avail)-1]
	}
	if n.Chance > 0 && w.R.Float64() > n.Chance {
		return
	}
	w.learn(c, n, true)
}

// learn adds a node, applies its side effects and fires its filter.
func (w *World) learn(c *Civ, n *tech.Node, fire bool) {
	c.Known[n.Key] = true
	for d, v := range n.Focus {
		if c.Focus[d] == 0 {
			c.Focus[d] = 1
		}
		c.Focus[d] *= v
	}
	if n.Key == c.Species.World.Unlock {
		for _, d := range c.Species.World.Locked {
			delete(c.Locked, d)
		}
	}
	was := c.Stage
	w.recompute(c)
	if n.Milestone && n.Text != "" {
		w.log(n.Text, c.Name)
	}
	if n.Key == "deep_time" {
		c.KnowsCycle = true
		w.log("The %s find their place in the turn: the age dawned %.0f million years ago, the galaxy is %s as fertile as it was then, and the next dawn is %.0f million years away. They will not see it.",
			c.Name, float64(w.Now-w.Cycle.Surges[len(w.Cycle.Surges)-1])/1e6, percent(w.fertility()), float64(w.NextSurge()-w.Now)/1e6)
	}
	if was == Emergent && c.Stage == Interstellar && !n.Milestone {
		w.log("The %s reach the stars.", c.Name)
	}
	if fire && n.Filter != "" {
		w.face(c, n.Filter, 0)
	}
}

// focus tilts research toward a domain for a while, in answer to a filter.
func (c *Civ) focus(domain string, by float64) {
	if c.Focus[domain] == 0 {
		c.Focus[domain] = 1
	}
	c.Focus[domain] *= by
}
