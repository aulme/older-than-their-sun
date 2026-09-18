package mind

import (
	"fmt"

	"worldgen/internal/flow"
)

// Building: what a people raises when it raises something. It builds what
// fixes its deepest want; with nothing wanting, the largest source in
// reach that nothing harnesses yet, or the work that lifts it most.

// Site is one structure at one star a people could build.
type Site struct {
	Key    string
	Star   int
	Yield  flow.Income // what it would give there, per tick
	Upkeep flow.Income // what it would cost
	Levels float64     // the levels it lifts, all three summed
}

// BuildInput is what the choice is made from.
type BuildInput struct {
	Sites []Site
	Want  flow.Income // what more by kind would run everything
	Spare flow.Income // what is left this tick after the fed uses
}

// BuildChoice is the decision.
type BuildChoice struct {
	Pick   int     // index into Sites, or -1 for nothing
	Fixes  float64 // how much of the want the pick meets
	Worth  float64 // its yield and levels, for the choice with nothing wanting
	Wanted bool    // the want decided it
}

// Why says the choice in a line.
func (b BuildChoice) Why() string {
	switch {
	case b.Pick < 0:
		return "nothing: the spare covers no structure twice over"
	case b.Wanted:
		return fmt.Sprintf("what meets %.1f of the want", b.Fixes)
	}
	return fmt.Sprintf("the largest thing in reach, worth %.1f", b.Worth)
}

// Build chooses a site. Only a site whose upkeep the spare covers Cover
// times over is affordable. Among those, the one that meets most of the
// want; with nothing wanting, or nothing meeting it, the one worth most:
// its yield plus its levels weighted. Ties go to the earlier site.
func Build(in BuildInput, t *Tuning) BuildChoice {
	p := &t.Build
	out := BuildChoice{Pick: -1}
	for i, s := range in.Sites {
		if !in.Spare.Covers(s.Upkeep.Scale(p.Cover)) {
			continue
		}
		fixes := 0.0
		for k := range s.Yield {
			fixes += min(s.Yield[k], in.Want[k])
		}
		worth := s.Yield.Total() + p.LevelWeight*s.Levels
		switch {
		case out.Pick < 0:
		case fixes > out.Fixes:
		case fixes == out.Fixes && !out.Wanted && worth > out.Worth:
		default:
			continue
		}
		out.Pick, out.Fixes, out.Worth, out.Wanted = i, fixes, worth, fixes > 0
	}
	return out
}
