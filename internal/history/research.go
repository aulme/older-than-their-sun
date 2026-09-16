package history

import "worldgen/internal/tech"

// research is a pursuit: a civilisation picks one node it can reach and
// banks points toward it until the price is paid. The pick is drawn from
// weights that are species tilt times current focus times momentum in the
// domains it already knows, so a people specialises and no one climbs the
// whole tree. Deep nodes cost far more than shallow ones, and miracles cost
// more still and are a conscious choice: the leap.
func (w *World) research(c *Civ) {
	rate := 0.12 * c.Species.Rate() * (1 + 0.08*c.Soc) * (1 + 0.03*float64(len(c.Systems)))
	rate *= c.rateMul(w)
	c.Progress += rate * w.dt
	for c.Active() {
		if c.Pursuit != "" && !w.canPursue(c, tech.Get(c.Pursuit)) {
			c.Pursuit = ""
		}
		if c.Pursuit == "" {
			c.Pursuit = w.choose(c)
			if c.Pursuit == "" {
				c.Progress = 0
				return
			}
			if n := tech.Get(c.Pursuit); n.Miracle {
				w.log("The %s turn everything they have toward %s. It will take ages, and it may not come.", c.Name, n.Name)
			}
		}
		n := tech.Get(c.Pursuit)
		if c.Progress < n.Price() {
			return
		}
		c.Progress -= n.Price()
		c.Pursuit = ""
		if n.Chance > 0 && w.R.Float64() > n.Chance {
			if n.Miracle {
				w.log("The %s come close to %s and fall short. The work of ages goes for nothing.", c.Name, n.Name)
			}
			continue
		}
		w.learn(c, n, true)
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
	if len(c.held()) > 0 {
		m *= 1.5 // the miracle pulls everything else along
		if c.surging(w.Now) {
			m *= 1.5
		}
	}
	if c.miracle("ansible") {
		m *= 1.5 // every mind in one room
	}
	if n := tech.Get(c.Pursuit); n != nil {
		switch n.Domain {
		case tech.Exotic:
			m *= w.Law.ExoticMul() // dead stars nearby to learn from
		case tech.Industry:
			m *= w.Law.IndustryMul() // metals are ore
		}
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

// canPursue says whether a node is open to a civilisation now.
func (w *World) canPursue(c *Civ, n *tech.Node) bool {
	if n == nil || c.Known[n.Key] || (c.Locked[n.Domain] && n.Key != c.Species.World.Unlock) {
		return false
	}
	if n.Patience > 0 && float64(w.Now-c.Born)/1000 < n.Patience {
		return false
	}
	if n.Miracle && c.miracle(n.Key) {
		return false // already theirs by birth or by a find
	}
	for _, p := range n.Prereqs {
		if !c.Known[p] {
			return false
		}
	}
	return true
}

// choose picks the next pursuit.
func (w *World) choose(c *Civ) string {
	depth := map[string]int{}
	for k := range c.Known {
		if n := tech.Get(k); n.Era >= 1 {
			depth[n.Domain]++
		}
	}
	var avail []*tech.Node
	var weights []float64
	total := 0.0
	for _, n := range tech.Nodes {
		if !w.canPursue(c, n) {
			continue
		}
		f := c.Focus[n.Domain]
		if f == 0 {
			f = 1
		}
		wt := n.Weight * c.Species.DomainMul(n.Domain) * f * (1 + 0.25*float64(depth[n.Domain]))
		if c.Locked[n.Domain] {
			wt = n.Weight
		}
		if n.Miracle {
			wt = n.Weight * c.leapWeight(n.Key)
		}
		avail = append(avail, n)
		weights = append(weights, wt)
		total += wt
	}
	if len(avail) == 0 || total <= 0 {
		return ""
	}
	x := w.R.Float64() * total
	for i, a := range avail {
		x -= weights[i]
		if x < 0 {
			return a.Key
		}
	}
	return avail[len(avail)-1].Key
}

// learn adds a node, applies its side effects and fires its filter.
func (w *World) learn(c *Civ, n *tech.Node, fire bool) {
	c.Known[n.Key] = true
	if _, ok := c.Learned[n.Key]; !ok {
		c.Learned[n.Key] = w.Now
	}
	if c.Pursuit == n.Key {
		c.Pursuit = ""
	}
	if n.Miracle {
		how := "leap"
		if w.finding {
			how = "found"
		}
		w.gain(c, n.Key, how)
	}
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
