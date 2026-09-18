package history

import (
	"sort"

	"worldgen/internal/species"
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
		era = max(era, n.Era)
		if !c.working(k) {
			// dormant: no levels, no reach; the envelope holds a while
			if !c.starved(k, w.Now) {
				env += n.Env
			}
			continue
		}
		mil, sur, soc = mil+n.Mil, sur+n.Sur, soc+n.Soc
		reach = max(reach, n.Reach)
		if n.Speed > 0 {
			speed = min(speed, n.Speed)
		}
		env += n.Env
	}
	working := map[string]int{}
	for _, wk := range c.Works {
		if !wk.Dark {
			working[wk.Key]++
		}
	}
	for _, key := range structureKeys {
		cnt := working[key]
		if cnt == 0 {
			continue
		}
		s := tech.Structures[key]
		m := 1 + 0.25*float64(min(cnt-1, 2))
		mil, sur, soc = mil+s.Mil*m, sur+s.Sur*m, soc+s.Soc*m
	}
	// rarities: what is had gives its levels, and a beacon its reach; a
	// wielded artifact is one of them
	rm, rs, rc, rr := w.levelsFromRarities(c)
	mil, sur, soc, reach = mil+rm, sur+rs, soc+rc, reach+rr
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
	if n := len(c.Systems); n > 4 && !ansible && !c.Species.Is(species.Hive) {
		soc -= min(3, 0.15*float64(n-4))
	}
	// dominion: the held feed the master's armies, works and confidence, and
	// lose more than the master gains; a vassal gives half and loses half
	slaves, vassals := 0, 0
	for _, o := range w.Civs {
		if o.Living() && o.Master == c.ID {
			if o.Vassal {
				vassals++
			} else {
				slaves++
			}
		}
	}
	held := min(4, float64(slaves)+0.5*float64(vassals))
	mil, sur, soc = mil+0.4*held, sur+0.2*held, soc+0.3*held
	if !c.Free() {
		if c.Vassal {
			mil, sur, soc = mil-0.3, sur-0.1, soc-0.2
		} else {
			mil, sur, soc = mil-0.8, sur-0.4, soc-0.6
		}
	}
	p := c.Species.Profile()
	env += p.Env
	if c.Species.Sub == species.Parasite {
		soc += min(2, 0.5*float64(c.Hosts+slaves)) // a parasite is as rich as its hosts
	}
	if c.Species.Is(species.Planetary) && c.Known["grafting"] {
		reach /= p.Reach // a world's reach is a third until it learns to graft
	}
	soc += c.Morale
	c.Quality = clamp(mil, 0, 10)
	if c.Aloft {
		mil = w.ships(c) // the fleets are the people
	}
	mil -= c.Away // what is out with the fleets
	c.Mil, c.Sur, c.Soc = clamp(mil, 0, 10), clamp(sur, 0, 10), clamp(soc, 0, 10)
	w.setDials(c)
	reach *= c.Species.ReachMul()
	if !c.Free() {
		if c.Vassal {
			reach *= 0.5
		} else {
			reach = 0
		}
	}
	c.Reach, c.Speed, c.Era = reach, speed, era
	if c.Starfaring == 0 && reach >= 1 {
		c.Starfaring = w.Now
	}
	c.Envelope = env
	if c.Sur >= 5 {
		c.Envelope++
	}
	if c.Sur >= 8 {
		c.Envelope++
	}
	if c.Aloft && c.Stage == Emergent {
		c.Stage = Interstellar
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
