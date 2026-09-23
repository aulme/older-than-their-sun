package history

import (
	"math"
	"os"
	"slices"
	"testing"
	"time"

	"worldgen/internal/species"
)

// TestLoreRulesAreWhole holds the one constraint the kept dials put on
// the rules: every coefficient must be a multiple of 0.005, since the
// running total is integers of a four-hundredth of a point and a tale's
// wear multiplies by halves. A rule that breaks it would be silently
// rounded, and the kept number would stop being the walked one.
func TestLoreRulesAreWhole(t *testing.T) {
	t.Parallel()
	for i, d := range loreRules {
		for j, x := range [8]float64{d.Aggression, d.Risk, d.Greed, d.Fear, d.Loyalty, d.Hunger, d.Patience, d.Hate} {
			n := x * (dialDenom / 2)
			if math.Abs(n-math.Round(n)) > 1e-9 {
				t.Errorf("rule %d dial %d is %v, which is not a whole four-hundredth of a point", i, j, x)
			}
		}
		for wear := int8(0); wear <= 2; wear++ {
			u := unitsOf(d, wear)
			k := 1 + 0.5*float64(wear)
			for j, x := range [8]float64{d.Aggression, d.Risk, d.Greed, d.Fear, d.Loyalty, d.Hunger, d.Patience, d.Hate} {
				if got, want := float64(u[j])/dialDenom, x*k; math.Abs(got-want) > 1e-12 {
					t.Errorf("rule %d dial %d at wear %d: kept %v, meant %v", i, j, wear, got, want)
				}
			}
		}
	}
}

// walked is the kept summaries computed the long way, without touching
// the people. It shares the rules with the kept totals on purpose: what
// it holds is the bookkeeping — that every tale learned, worn, forgotten
// or restored moved the total the right way — not that the rules
// themselves say what they used to. That was checked against the walk
// they replaced, on a whole run's dossier; see specs/plan.md step 7.
func walked(w *World, c *Civ) (dialUnits, int) {
	var u dialUnits
	n := 0
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Events[t.Fact]
		u.add(taleUnits(c, f, t), 1)
		if ownGrief(c, f) {
			n++
		}
	}
	return u, n
}

// walkedSick is the kept sickness tales computed the long way: the
// tales the suspicion pass would find if it walked the telling, in the
// telling's order, which is the order it read them in before the list
// was kept.
func walkedSick(w *World, c *Civ) []*Tale {
	var out []*Tale
	for _, t := range c.Lore {
		if sickTale(c, w.Events[t.Fact]) {
			out = append(out, t)
		}
	}
	return out
}

// TestSummariesAreKept is the gate for summaries.go: at every tick of a
// whole age, every people whose kept summaries claim to be good holds
// exactly what a walk of its telling gives. The totals are integers, so
// the check is exact and not a tolerance — which is the point of keeping
// them as integers. A tale learned, worn, forgotten, pruned, restored
// off a wall, inherited or carried through a sundering all pass under
// it, and a people that changes its mind has to have read its telling
// again before the tick is out.
func TestSummariesAreKept(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Stars = 200
	checked, tales := 0, 0
	cfg.Sample = func(w *World) {
		for _, c := range w.Civs {
			u, n := walked(w, c)
			// the griefs are right at every moment: their sort is the
			// fact's own, and no change of judgment moves it
			if n != c.experience {
				t.Fatalf("at %d, %s keeps %d griefs and its telling gives %d", w.Now, c.Tok(), c.experience, n)
			}
			// the sickness tales are right at every moment too, and in
			// the telling's own order: what the suspicion pass reads
			// must be what a walk would have handed it
			if sick := walkedSick(w, c); !slices.Equal(sick, c.sickLore) {
				t.Fatalf("at %d, %s keeps %d sickness tales and its telling gives %d", w.Now, c.Tok(), len(c.sickLore), len(sick))
			}
			if !c.loreKept {
				continue // a whole telling moved, or a mind: the next read takes the dials again
			}
			if u != c.loreUnits {
				t.Fatalf("at %d, %s keeps dials %v and its telling gives %v", w.Now, c.Tok(), c.loreUnits, u)
			}
			checked++
			tales += len(c.Lore)
		}
	}
	Generate(11, cfg)
	if checked < 1000 || tales < 100_000 {
		t.Fatalf("checked %d peoples over %d tales: too little to hold anything", checked, tales)
	}
}

// TestJudgmentRereadsTheTelling: the dials' rules are read through a
// people's morality, so a people that changes its mind is not left with
// the dials of the mind it had. Nothing else in a telling moves when a
// morality does, which is why the summaries are kept at all.
func TestJudgmentRereadsTheTelling(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Stars = 200
	cfg.Until = 6_000_000
	w := Generate(3, cfg)
	moved := 0
	for _, c := range w.Civs {
		if len(c.Lore) < 20 {
			continue
		}
		before := w.loreDials(c)
		for _, m := range []Morality{{Kind: Individual}, {Kind: Herd}, {Kind: Amoral}, {Kind: Fixation, Object: Conquest}} {
			if m == c.Morality {
				continue
			}
			c.think(m)
			u, _ := walked(w, c)
			if got := w.loreDials(c); got != u.dials() {
				t.Fatalf("%s changed its mind and kept the old dials: %v against %v", c.Tok(), got, u.dials())
			}
			if w.loreDials(c) != before {
				moved++
			}
		}
	}
	if moved == 0 {
		t.Fatal("no people's dials moved with its judgment: the test is checking nothing")
	}
}

// TestSummaryCost is what keeping the summaries buys, asked the way step
// 6 asked the wearing's: a run with the summaries kept is a different
// history from a run without, so the two runs' times say nothing. It is
// asked instead on one telling in one process — the walk against the
// kept read, on the same tales.
//
// The walk here is one walk for both summaries. The code it replaced
// made two, one in setDials and one in reckon, so the saving is the
// larger of the two figures below, not the smaller.
//
//	SUMMARY=1 go test ./internal/history -run TestSummaryCost -v
func TestSummaryCost(t *testing.T) {
	if os.Getenv("SUMMARY") == "" {
		t.Skip("a measurement, not a gate: set SUMMARY=1 to run it")
	}
	const reps, ticks = 20, 2000
	t.Logf("%d tellings, %d reads of each", reps, ticks)
	t.Logf("%8s %12s %12s %10s", "tales", "walked ns", "kept ns", "saved")
	for _, tales := range []int{50, 100, 200, 400} {
		var walk, kept time.Duration
		for rep := range reps {
			cfg := DefaultConfig()
			cfg.Stars = 30
			w := newWorld(uint64(2000+rep), cfg)
			w.dt, w.Now = float64(cfg.Step)/1000, cfg.Dawn
			c := spawnAt(w, 0, species.GenerateWith(w.R, 1, "lush", species.Biological, 0))
			fillTelling(w, c, tales, rep)
			w.resum(c)
			start := time.Now()
			for range ticks {
				walked(w, c)
			}
			walk += time.Since(start)
			start = time.Now()
			for range ticks {
				w.loreDials(c)
				_ = c.experience
			}
			kept += time.Since(start)
		}
		a := float64(walk.Nanoseconds()) / float64(reps*ticks)
		b := float64(kept.Nanoseconds()) / float64(reps*ticks)
		t.Logf("%8d %12.0f %12.0f %9.1f%%", tales, a, b, 100*(1-b/a))
	}
}
