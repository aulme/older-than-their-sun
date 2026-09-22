package history

import (
	"math"
	"math/rand/v2"
	"os"
	"testing"

	"worldgen/internal/species"
)

// TestWearCadence asks whether putting a telling through the wearing
// every n ticks, with each tale's rate compounded over the gap, wears
// it as a telling put through every tick.
//
// It cannot be asked of whole runs. Changing the cadence changes what
// is drawn and when, so every run is a different world, and the worlds
// are wildly unlike each other: over twenty-four seeds at 120 stars the
// wars of a run ran from 2 to 3100, a standard deviation of two and a
// half times the mean, and two halves of the same twenty-four differed
// by three hundred percent. Neither the per-run numbers nor the tales
// pooled out of them can see a five percent shift through that.
//
// So the wearing is asked on its own: one people, a telling built to
// order, and nothing running but the wearing. Then the tales are
// independent and the answer is the process itself, which is what the
// compounding is a claim about.
//
//	WEAR=1 go test ./internal/history -run TestWearCadence -v
func TestWearCadence(t *testing.T) {
	if os.Getenv("WEAR") == "" {
		t.Skip("a measurement, not a gate: set WEAR=1 to run it")
	}
	const reps, tales, ticks = 200, 400, 600
	t.Logf("%d runs of %d tales over %d ticks each: %d tales in all per cadence", reps, tales, ticks, reps*tales)
	t.Logf("a share of p over N has a standard error of sqrt(p(1-p)/N); at p=0.2 and N=%d that is %.4f, or %.2f%% of it",
		reps*tales, math.Sqrt(0.2*0.8/float64(reps*tales)), 100*math.Sqrt(0.2*0.8/float64(reps*tales))/0.2)
	t.Logf("%6s %10s %10s %10s %10s %12s", "every", "exact", "worn", "myth", "forgot", "max dev")
	var base [4]float64
	for _, every := range []int{1, 2, 4, 8, 16, 32, 64, 128} {
		var got [4]float64
		for rep := range reps {
			w := newTestWorld(t, uint64(1000+rep), 30)
			w.Cfg.WearEvery = every
			c := spawnAt(w, 0, species.GenerateWith(w.R, 1, "lush", species.Biological, 0))
			fillTelling(w, c, tales, rep)
			for range ticks {
				w.Ticks++
				w.Now += w.Cfg.Step
				w.wear(c)
			}
			for _, tl := range c.Lore {
				switch {
				case tl.Forgot:
					got[3]++
				default:
					got[tl.Wear]++
				}
			}
		}
		n := float64(reps * tales)
		for i := range got {
			got[i] /= n
		}
		if every == 1 {
			base = got
			t.Logf("%6d %10.4f %10.4f %10.4f %10.4f %12s", every, got[0], got[1], got[2], got[3], "-")
			continue
		}
		// only the shares with enough behind them to mean anything: a
		// share under a twentieth is a handful of tales and says more
		// about the draw than the cadence
		dev := 0.0
		for i := range got {
			if base[i] >= 0.05 {
				if d := abs(got[i]/base[i] - 1); d > dev {
					dev = d
				}
			}
		}
		flag := ""
		if dev > 0.05 {
			flag = "  OVER 5%"
		}
		t.Logf("%6d %10.4f %10.4f %10.4f %10.4f %11.1f%%%s", every, got[0], got[1], got[2], got[3], 100*dev, flag)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// fillTelling gives a people a telling of n tales spread over the ages
// behind it, of assorted weights, so the wearing is asked of the spread
// of rates it really sees and not of one.
func fillTelling(w *World, c *Civ, n, rep int) {
	r := rand.New(rand.NewPCG(uint64(rep), 0x5eed))
	kinds := []Kind{FWar, FBetrayal, FMet, FTrade, FTaken, FPlague, FSettle, FEnslaved}
	for i := range n {
		k := kinds[i%len(kinds)]
		age := Year(r.Int64N(30_000_000)) // up to thirty million years behind
		e := &Event{ID: len(w.Events), Year: w.Now - age, Kind: k, Subject: c.ID + 1 + i%3, Object: c.ID, Star: -1, Plague: -1}
		w.Events = append(w.Events, e)
		c.Lore = append(c.Lore, &Tale{Fact: e.ID, Learned: w.Now - age, Source: Witnessed, From: -1, Blamed: -1})
	}
}
