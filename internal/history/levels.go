package history

import (
	"sort"

	"worldgen/internal/tech"
)

// structureKeys is a fixed order for summing structures, so float sums do not depend on map order.
var structureKeys = func() []string {
	var out []string
	for k := range tech.Structures {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}()

// recompute derives the three levels, reach, speed, envelope and era from
// species, known tech, structures, wielded artifacts, scars and morale.
func (w *World) recompute(c *Civ) {
	mil, sur, soc := c.Species.Base()
	reach, speed, env, era := 0.0, 100.0, 0, 0
	ansible := c.miracle("ansible")
	for _, k := range knownOf(c) {
		n := tech.Get(k)
		mil, sur, soc = mil+n.Mil, sur+n.Sur, soc+n.Soc
		reach = max(reach, n.Reach)
		if n.Speed > 0 {
			speed = min(speed, n.Speed)
		}
		env += n.Env
		era = max(era, n.Era)
	}
	for _, key := range structureKeys {
		cnt := c.Structures[key]
		if cnt == 0 {
			continue
		}
		s := tech.Structures[key]
		m := 1 + 0.25*float64(min(cnt-1, 2))
		mil, sur, soc = mil+s.Mil*m, sur+s.Sur*m, soc+s.Soc*m
	}
	for _, l := range c.Wielded {
		switch l.Level {
		case "mil":
			mil += 2
		case "sur":
			sur += 1.5
		case "soc":
			soc += 1.5
		case "reach":
			reach += 15
		case "all":
			mil, sur, soc = mil+1, sur+1, soc+1
		}
	}
	// miracles: the dominant fact about whoever holds one
	if ansible {
		soc += 3
		reach *= 1.5
	}
	if c.miracle("directed_evolution") {
		sur += 3
		env += 2
		// they do not build ships; they breed bodies that cross the dark on their own
		reach = max(reach, 12)
		speed = min(speed, 100)
	}
	if c.miracle("ftl") {
		reach = max(reach, 45)
		speed = min(speed, 0.3)
		mil += 1
	}
	if c.miracle("unmaking") {
		mil += 4
	}
	if c.miracle("chorus") {
		soc += 4
	}
	if c.miracle("foresight") {
		mil, sur, soc = mil+1, sur+1, soc+1
	}
	if c.Boons[BoonAligned] {
		soc += 0.5
	}
	if c.Boons[BoonUnity] {
		soc += 0.5
	}
	if c.Boons[BoonCommunion] {
		mil, sur, soc = mil+1, sur+1, soc+1
	}
	if c.Scars[ScarCentralism] {
		soc += 1
	}
	if c.Scars[ScarOssified] {
		soc -= 0.5
	}
	if c.Scars[ScarLeftBehind] {
		mil, soc = mil-1, soc-1
	}
	// distance: colonies drift unless something holds them together
	if n := len(c.Systems); n > 4 && !ansible && !c.Has("hive") {
		soc -= min(3, 0.15*float64(n-4))
	}
	// slaves feed the master's armies
	slaves := 0
	for _, o := range w.Civs {
		if o.Living() && o.Master == c.ID && !o.Vassal {
			slaves++
		}
	}
	mil += min(2, 0.5*float64(slaves))
	soc += c.Morale
	c.Mil, c.Sur, c.Soc = clamp(mil, 0, 10), clamp(sur, 0, 10), clamp(soc, 0, 10)
	reach *= c.Species.ReachMul()
	if !c.Free() {
		if c.Vassal {
			reach *= 0.5
		} else {
			reach = 0
		}
	}
	c.Reach, c.Speed, c.Era = reach, speed, era
	c.Envelope = env
	if c.Sur >= 5 {
		c.Envelope++
	}
	if c.Sur >= 8 {
		c.Envelope++
	}
	if c.Stage == Emergent && c.Reach >= 10 {
		c.Stage = Interstellar
	}
}

// avg of the named levels.
func (c *Civ) level(names ...string) float64 {
	t := 0.0
	for _, n := range names {
		switch n {
		case "mil":
			t += c.Mil
		case "sur":
			t += c.Sur
		case "soc":
			t += c.Soc
		}
	}
	return t / float64(len(names))
}
