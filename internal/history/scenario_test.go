package history

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"worldgen/internal/mind"
	"worldgen/internal/plague"
)

// The scenario tests: each is a spec in testdata/scenarios, the moment a
// batch went wrong at built by hand (scenario.go), run through the real
// tick on several seeds with assertions on what came of it. On a failure
// the narrative is printed, so the failure explains itself; SCENARIO_V=1
// prints it on a pass too. `go run ./cmd/scenario <spec>` tells the same
// run.

// scenario builds the named spec on a seed, telling it into a buffer.
func scenario(t *testing.T, name string, seed uint64) (*Run, *bytes.Buffer) {
	t.Helper()
	s, err := LoadScenario(filepath.Join("testdata", "scenarios", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	s.Seed = seed
	var out bytes.Buffer
	r, err := s.Build(&out)
	if err != nil {
		t.Fatal(err)
	}
	return r, &out
}

// told prints a run's narrative, the last lines of it if long, when the
// test failed or SCENARIO_V is set.
func told(t *testing.T, out *bytes.Buffer) {
	t.Helper()
	if !t.Failed() && os.Getenv("SCENARIO_V") == "" {
		return
	}
	lines := strings.Split(out.String(), "\n")
	if len(lines) > 400 && os.Getenv("SCENARIO_V") == "" {
		lines = append(lines[:40], append([]string{"  …"}, lines[len(lines)-360:]...)...)
	}
	t.Log("\n" + strings.Join(lines, "\n"))
}

// seeds is the seeds a scenario test runs on: a behaviour that holds on
// one draw of the streams and not another is not a rule.
var seeds = []uint64{1, 2, 3, 4, 5}

// eachSeed runs a spec on every seed as a parallel subtest, for its
// ticks, and hands the run to check.
func eachSeed(t *testing.T, name string, check func(t *testing.T, r *Run)) {
	for _, seed := range seeds {
		t.Run(sprintf("seed%d", seed), func(t *testing.T) {
			t.Parallel()
			r, out := scenario(t, name, seed)
			defer told(t, out)
			r.Ticks(r.S.Ticks)
			check(t, r)
		})
	}
}

// fewWars fails when a pair has fought more wars than it should have in
// the run: the loop check every loop scenario makes.
func fewWars(t *testing.T, r *Run, a, b string, most int) {
	t.Helper()
	if n := len(r.Wars(a, b)); n > most {
		t.Errorf("%s and %s fought %d wars in %d ticks, want %d at most", a, b, n, r.S.Ticks, most)
	}
}

// The loop cases the war council's first batches found (step 10, stage
// 1; specs/notes/war-stage1-handover.md). Each is a pair that fought the
// same war over and over, and each scenario asserts the pair fights a
// handful of wars in three hundred thousand years at most, and the rule
// that broke the loop.

// TestScenarioPacifistClaims: two pacifist heirs with claims on each
// other's worlds do not go to war over them at all — the step-19 loop,
// five hundred empty wars; a pacifist is a pacifist before its claim.
func TestScenarioPacifistClaims(t *testing.T) {
	eachSeed(t, "pacifist-claims", func(t *testing.T, r *Run) {
		fewWars(t, r, "Elder", "Younger", 0)
	})
}

// TestScenarioClaims: heirs with claims on each other fight a few wars,
// and a claim settled by terms is given up by the side that took them.
func TestScenarioClaims(t *testing.T) {
	eachSeed(t, "claims", func(t *testing.T, r *Run) {
		fewWars(t, r, "Elder", "Younger", 6)
		for _, wr := range r.Wars("Elder", "Younger") {
			if wr.Result != "terms" {
				continue
			}
			for _, f := range r.W.Events {
				if f.Kind != FSettled || f.P["war"] != wr.ID {
					continue
				}
				l, v := r.W.Civs[f.Subject], r.W.Civs[f.Object]
				if v.Active() && l.Active() && claimsOn(v, l) > 0 {
					t.Errorf("war %d settled by terms, and %s still claims %d of %s's worlds", wr.ID, r.Name(v.ID), claimsOn(v, l), r.Name(l.ID))
				}
			}
		}
	})
}

// TestScenarioVengefulPacifist: a vengeful people with a grudge against a
// pacifist neighbour of its own size takes its redress or gives up, and
// so do the heirs it breaks into: the loop was the two capitulating and
// settling with nothing yielded, the grudge standing and renewed by each
// war, and the heirs carrying it on (seventeen wars in seed 5 without
// the winner's grudge cleared and the yields counted).
func TestScenarioVengefulPacifist(t *testing.T) {
	eachSeed(t, "vengeful-pacifist", func(t *testing.T, r *Run) {
		meek := r.Civ("Meek").ID
		n := 0
		for _, wr := range r.W.Wars {
			if wr.Sides[0] == meek || wr.Sides[1] == meek {
				n++
			}
		}
		if n > 8 {
			t.Errorf("the pacifist was fought %d times in %d ticks", n, r.S.Ticks)
		}
	})
}

// TestScenarioConqueror: a conqueror far the stronger wins its war on a
// pacifist, which is not given a truce with the fleet at its worlds; the
// strength behind it counts its ships out (TestStrengthCountsShipsOut).
func TestScenarioConqueror(t *testing.T) {
	eachSeed(t, "conqueror", func(t *testing.T, r *Run) {
		wars := r.Wars("Khan", "Meek")
		if len(wars) == 0 {
			t.Fatal("no war")
		}
		for _, wr := range wars {
			switch {
			case !wr.Over:
			case wr.Result == "truce":
				t.Errorf("war %d ended in a truce", wr.ID)
			case wr.Winner == r.Civ("Meek").ID:
				t.Errorf("war %d won by the pacifist: %s", wr.ID, wr.Result)
			}
		}
		fewWars(t, r, "Khan", "Meek", 3)
	})
}

// TestScenarioWaking: a living world wakes on the people inside its
// neighbourhood as often as its demands are refused, and the first
// waking is its war; the ones after are blows, not a new war each.
func TestScenarioWaking(t *testing.T) {
	eachSeed(t, "waking", func(t *testing.T, r *Run) {
		fewWars(t, r, "Gaia", "Settlers", 1)
	})
}

// TestScenarioHater: a xenophobe beside a machine people fights it to the
// end or not at all; the loop was a hater taking tribute at every truce.
func TestScenarioHater(t *testing.T) {
	eachSeed(t, "hater", func(t *testing.T, r *Run) {
		fewWars(t, r, "Purist", "Machine", 4)
		tribute := 0
		for _, wr := range r.Wars("Purist", "Machine") {
			if wr.Result == "tribute" {
				tribute++
			}
		}
		if tribute > 1 {
			t.Errorf("the hater took tribute %d times", tribute)
		}
	})
}

// TestScenarioRiders: two riders of two peoples that trade do not pass
// one host between them: one rider at a time wears a people.
func TestScenarioRiders(t *testing.T) {
	for _, seed := range seeds {
		t.Run(sprintf("seed%d", seed), func(t *testing.T) {
			t.Parallel()
			r, out := scenario(t, "riders", seed)
			defer told(t, out)
			w, host, other := r.W, r.Civ("Host"), r.Civ("Other")
			var riders []*Civ
			for _, c := range []*Civ{host, other} {
				p := w.newPlague(plague.Biological, c, "clean")
				p.Contagion, p.Lethality, p.Conscious = 0.9, 0.3, true
				w.infect(c, p, nil, "born")
				riders = append(riders, w.wake(p, c))
			}
			r.Named("Rider", riders[0])
			r.Named("Rival", riders[1])
			w.infect(host, w.Plagues[riders[1].Own], nil, "born") // the rival's plague in the host too
			r.Ticks(r.S.Ticks)
			changes := 0
			for _, f := range w.Events {
				if (f.Kind == FEnslaved || f.Kind == FVassal) && f.Object == host.ID {
					changes++
				}
			}
			if changes > 3 {
				t.Errorf("the host changed hands %d times in %d ticks", changes, r.S.Ticks)
			}
		})
	}
}

// TestScenarioHunt: a hunt declared on a region of doerless losses runs
// on its own clock and ends; the loop was hunts re-declared at once,
// under a drain that never ran down.
func TestScenarioHunt(t *testing.T) {
	for _, seed := range seeds {
		t.Run(sprintf("seed%d", seed), func(t *testing.T) {
			t.Parallel()
			r, out := scenario(t, "hunt", seed)
			defer told(t, out)
			c, x := r.Civ("Hunter"), r.Civ("Unseen")
			losses(r.W, c, x, x.Home, 3)
			r.W.deduce(c)
			if c.Tally.Hunts == 0 {
				t.Skip("no hunt on this field")
			}
			r.Ticks(r.S.Ticks)
			fewWars(t, r, "Hunter", "Unseen", 4)
		})
	}
}

// TestScenarioPacifistSpares: a pacifist people under an unyielding
// leader takes its pacifist neighbour's home and spares it; that is a war
// won, and the winner's grudge is settled by it. The loop was the spared
// home told as a peace with no winner, the grudge renewed, and the old
// quarrel taken up again every eighteen ticks (seed 3 of the first
// watched batch; eighteen wars in seed 1 of this spec).
func TestScenarioPacifistSpares(t *testing.T) {
	eachSeed(t, "pacifist-spares", func(t *testing.T, r *Run) {
		fewWars(t, r, "Warden", "Meek", 3)
	})
}

// TestScenarioHuntVassal: a hunt on a region where the anti-memetic
// people is another's vassal runs as a hunt does, on its own clock. The
// loop was the war ended at once as held, since a side was under a
// master, and the hunt declared again the next tick from the same
// ledger: 160 hunts of one pair in seed 12 of the first watched batch.
func TestScenarioHuntVassal(t *testing.T) {
	for _, seed := range seeds {
		t.Run(sprintf("seed%d", seed), func(t *testing.T) {
			t.Parallel()
			r, out := scenario(t, "hunt-vassal", seed)
			defer told(t, out)
			c, x := r.Civ("Hunter"), r.Civ("Unseen")
			losses(r.W, c, x, x.Home, 3)
			r.W.deduce(c)
			if c.Tally.Hunts == 0 {
				t.Skip("no hunt on this field")
			}
			r.Ticks(r.S.Ticks)
			fewWars(t, r, "Hunter", "Unseen", 4)
		})
	}
}

// TestScenarioUnyieldingStare: two unyielding peoples at war out of each
// other's reach, neither fighting, tire of it: the unyielding pay a share
// of the idle cost, not none, so the stare ends within a few dozen ticks.
func TestScenarioUnyieldingStare(t *testing.T) {
	eachSeed(t, "unyielding-stare", func(t *testing.T, r *Run) {
		for _, wr := range r.Wars("Stone", "Iron") {
			if !wr.Over || wr.Ended-wr.Began > 40_000 {
				t.Errorf("war %d stared for %d years (over %v)", wr.ID, max(wr.Ended, r.W.Now)-wr.Began, wr.Over)
			}
		}
	})
}

// TestScenarioPactCascade: two defence pacts face each other and the
// principals go to war. An ally with no ships to send does not go to war
// in its own name when called, so the cascade does not fill with empty
// wars between allies that fight nothing (seed 11 of the third watched
// batch: 77 wars in a tick, pairs of shipless unyielding allies at war
// twenty times, never a battle). Allies that come to have ships may join
// later; the principals' rivalry stays a handful of wars.
func TestScenarioPactCascade(t *testing.T) {
	for _, seed := range seeds {
		t.Run(sprintf("seed%d", seed), func(t *testing.T) {
			t.Parallel()
			r, out := scenario(t, "pact-cascade", seed)
			defer told(t, out)
			r.Ticks(2) // the calls arrive
			for _, pair := range [][2]string{{"Ward", "Rival"}, {"Ward", "Kin"}, {"Kin", "Crown"}} {
				fewWars(t, r, pair[0], pair[1], 0)
			}
			r.Ticks(r.S.Ticks - 1)
			fewWars(t, r, "Crown", "Rival", 8)
		})
	}
}

// Stage 2 (step 11): allies and old enemies.

// TestScenarioRivals: two defensive peoples with two wars behind them
// and a grudge standing go to war again, though neither would strike
// anyone else (not without the rivalry: no seed has a war then); the
// third war between rivals of a size is for redress, not the whole
// (TestEscalate has the larger declarer's); and the rivalry is a handful
// of wars, not a loop.
func TestScenarioRivals(t *testing.T) {
	var warred atomic.Int32
	t.Cleanup(func() {
		if n := warred.Load(); n < 2 {
			t.Errorf("old enemies went to war again in %d of %d seeds, want 2 or more", n, len(seeds))
		}
	})
	eachSeed(t, "rivals", func(t *testing.T, r *Run) {
		ws := r.Wars("Carthage", "Rome")
		if len(ws) > 0 {
			warred.Add(1)
		}
		fewWars(t, r, "Carthage", "Rome", 6)
		for _, wr := range ws {
			if wr.Nth == 3 && wr.Aim != mind.AimRedress { // the third, declared between equals; a later one may follow a war that made one the larger
				t.Errorf("war %d, the %s between old enemies of a size, is for %s, want redress", wr.ID, ordinal(wr.Nth), wr.Aim)
			}
		}
	})
}

// TestScenarioAllies: a conqueror declares on a people with two allies
// in defence pacts, one near the conqueror and one near the principal.
// Each ally with a front goes to war in its own name, for defence; one
// too small to carry the war to the enemy sends ships to stand with its
// principal; and its war ends with its principal's, in its own defeat,
// or when its will is spent in a separate peace the record counts
// against it.
func TestScenarioAllies(t *testing.T) {
	var joined, stood atomic.Int32
	t.Cleanup(func() {
		if n := joined.Load(); n < 3 {
			t.Errorf("allies joined the war in %d of %d seeds, want 3 or more", n, len(seeds))
		}
		if n := stood.Load(); n < 3 {
			t.Errorf("an ally too small to press stood with its principal in %d of %d seeds, want 3 or more", n, len(seeds))
		}
	})
	eachSeed(t, "allies", func(t *testing.T, r *Run) {
		for _, x := range r.W.Expeditions {
			if x.Kind == Relief && x.Target == r.Civ("Crown").ID && (x.Owner == r.Civ("Ward").ID || x.Owner == r.Civ("Kin").ID) {
				stood.Add(1)
				break
			}
		}
		for _, ally := range []string{"Ward", "Kin"} {
			for _, wr := range r.Wars(ally, "Khan") {
				if wr.Principal >= 0 {
					joined.Add(1)
					if wr.Aim != mind.AimDefence {
						t.Errorf("%s's war %d, joined by pact, is for %s", ally, wr.ID, wr.Aim)
					}
				}
			}
		}
	})
}
