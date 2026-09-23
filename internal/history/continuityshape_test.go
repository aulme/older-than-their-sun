package history

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"
)

// TestContinuityShape is the measurement behind step 8's gate
// (specs/plan.md; specs/proposals/continuity.md): over a batch of whole
// ages it reads every people that can set, every hundred thousand years,
// for its continuity and its stiffness, and it pairs every dark age with
// the continuity its people had the tick before. It prints the two ends
// of the failure — the facings of Ossification per million years of
// life by continuity band, and a dark age's depth by the same bands —
// and what separates continuity from stiffness: their correlation over
// the samples, and how full the two corners that differ are. It draws
// nothing, so the ages it walks are the ages an unsampled run walks.
//
//	CONTINUITY=1 go test ./internal/history -run TestContinuityShape -v -timeout 4h
//
// CONTINUITY_STARS (400) and CONTINUITY_SEEDS (10) set the batch.
func TestContinuityShape(t *testing.T) {
	if os.Getenv("CONTINUITY") == "" {
		t.Skip("a measurement, not a gate: set CONTINUITY=1 to run it")
	}
	stars, seeds := 400, 10
	if s := os.Getenv("CONTINUITY_STARS"); s != "" {
		stars, _ = strconv.Atoi(s)
	}
	if s := os.Getenv("CONTINUITY_SEEDS"); s != "" {
		seeds, _ = strconv.Atoi(s)
	}
	out := make([]contShape, seeds)
	var wg sync.WaitGroup
	for i := range seeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = measureContinuity(uint64(i+1), stars)
		}()
	}
	wg.Wait()
	var all contShape
	for i, s := range out {
		t.Logf("seed %d: %s", i+1, s.line())
		all.add(s)
	}
	for _, l := range all.report() {
		t.Log(l)
	}
}

// lossBands are the bands the batch is read in, on the loss (one less
// the continuity) and not the continuity, since continuity bunches under
// one for everyone with medicine and letters and the loss is where
// peoples differ: a machine's is a third of a mayfly's.
var lossBands = []float64{0, 0.03, 0.045, 0.07, 0.15, 0.4, 1.01}

func lossBand(loss float64) int {
	for i := len(lossBands) - 2; i >= 0; i-- {
		if loss >= lossBands[i] {
			return i
		}
	}
	return 0
}

const nBands = 6

// contShape is what one or many ages gave. The exposure and the facings
// are split by era, early (0 to 2) and late (3 and 4), since the early
// are young and loose and would carry any correlation with age; the dark
// ages are split by stiffness, under one and over, since stiffness is the
// other term in a dark age's depth.
type contShape struct {
	samples       int
	cs, ss, ls    []float64 // continuity, stiffness and loss at each sample
	kyr           [2][nBands]float64
	faced         [2][nBands]int
	darks         [2][nBands]int
	depthSum      [2][nBands]float64
	past          float64 // continuity's term in the setting, summed over people-ticks
	pastN         int
	golden, blind int    // high continuity and loose ways; low continuity and set ways
	heavens       int    // still in heaven at the present: a heaven that fades or is taken ends under another cause
	uploads       [3]int // overcome, scarred, declined
	peoples       int
}

func (a *contShape) add(b contShape) {
	a.samples += b.samples
	a.cs = append(a.cs, b.cs...)
	a.ss = append(a.ss, b.ss...)
	a.ls = append(a.ls, b.ls...)
	for g := range 2 {
		for i := range nBands {
			a.kyr[g][i] += b.kyr[g][i]
			a.faced[g][i] += b.faced[g][i]
			a.darks[g][i] += b.darks[g][i]
			a.depthSum[g][i] += b.depthSum[g][i]
		}
	}
	a.past += b.past
	a.pastN += b.pastN
	a.golden += b.golden
	a.blind += b.blind
	a.heavens += b.heavens
	for i := range a.uploads {
		a.uploads[i] += b.uploads[i]
	}
	a.peoples += b.peoples
}

func (s contShape) line() string {
	return fmt.Sprintf("%d peoples, %d samples, r %.2f, heavens %d, uploads %v", s.peoples, s.samples, pearson(s.cs, s.ss), s.heavens, s.uploads)
}

func quantile(x []float64, q float64) float64 {
	if len(x) == 0 {
		return math.NaN()
	}
	y := append([]float64(nil), x...)
	sort.Float64s(y)
	return y[int(q*float64(len(y)-1))]
}

func (s contShape) report() []string {
	logs := make([]float64, len(s.ls))
	for i, l := range s.ls {
		logs[i] = math.Log(max(l, 1e-4))
	}
	out := []string{
		fmt.Sprintf("%d peoples, %d samples; continuity against stiffness r = %.2f, the log of the loss against stiffness r = %.2f", s.peoples, s.samples, pearson(s.cs, s.ss), pearson(logs, s.ss)),
		fmt.Sprintf("the loss: quartiles %.3f %.3f %.3f, tenth %.3f, ninetieth %.3f; continuity's term in the setting, mean over people-ticks %.3f",
			quantile(s.ls, 0.25), quantile(s.ls, 0.5), quantile(s.ls, 0.75), quantile(s.ls, 0.1), quantile(s.ls, 0.9), s.past/float64(max(s.pastN, 1))),
		fmt.Sprintf("corners: continuity over 0.95 with stiffness under 0.5 %d (%.1f%%); continuity under 0.9 with stiffness over 1.5 %d (%.1f%%)",
			s.golden, 100*float64(s.golden)/float64(max(s.samples, 1)), s.blind, 100*float64(s.blind)/float64(max(s.samples, 1))),
		fmt.Sprintf("the upload: overcome %d, scarred %d, declined into heaven %d; still in heaven at the present %d", s.uploads[0], s.uploads[1], s.uploads[2], s.heavens),
		"loss band     | eras 0-2: Myr  facings/Myr | eras 3-4: Myr  facings/Myr | stiff<1: dark ages  depth | stiff>=1: dark ages  depth",
	}
	for i := range nBands {
		out = append(out, fmt.Sprintf("%.3f-%.3f   | %13.0f  %11.3f | %13.0f  %11.3f | %18d  %5.3f | %19d  %5.3f",
			lossBands[i], min(lossBands[i+1], 1),
			s.kyr[0][i]/1000, 1000*float64(s.faced[0][i])/max(s.kyr[0][i], 1),
			s.kyr[1][i]/1000, 1000*float64(s.faced[1][i])/max(s.kyr[1][i], 1),
			s.darks[0][i], s.depthSum[0][i]/float64(max(s.darks[0][i], 1)),
			s.darks[1][i], s.depthSum[1][i]/float64(max(s.darks[1][i], 1))))
	}
	return out
}

// measureContinuity walks one age.
func measureContinuity(seed uint64, stars int) contShape {
	cfg := DefaultConfig()
	cfg.Stars = stars
	var s contShape
	type before struct {
		loss, stiff float64
		faced, band int
		late        bool
	}
	last := map[int]before{} // each people as it stood at the end of the tick before
	seen := 0
	n := 0
	cfg.Sample = func(w *World) {
		// the dark ages of this tick, read against the tick before
		for ; seen < len(w.Events); seen++ {
			e := w.Events[seen]
			if e.Kind != FDarkAge || e.Subject < 0 {
				continue
			}
			b, ok := last[e.Subject]
			if !ok {
				continue
			}
			d, _ := e.P["depth"].(float64)
			g := 0
			if b.stiff >= 1 {
				g = 1
			}
			s.darks[g][b.band]++
			s.depthSum[g][b.band] += d
		}
		for _, c := range w.Civs {
			if !c.Active() || !c.stiffens() {
				continue
			}
			k := w.continuity(c)
			loss := 1 - k
			// the facings of the tick just gone, against the band it was in
			if b, ok := last[c.ID]; ok {
				g := 0
				if b.late {
					g = 1
				}
				s.kyr[g][b.band] += w.dt
				s.faced[g][b.band] += c.Tally.OssFaced - b.faced
			}
			last[c.ID] = before{loss: loss, stiff: c.Stiff, faced: c.Tally.OssFaced, band: lossBand(loss), late: c.Era >= 3}
			s.past += w.contStiff(c)
			s.pastN++
			if n%100 == 0 {
				s.samples++
				s.cs = append(s.cs, k)
				s.ss = append(s.ss, c.Stiff)
				s.ls = append(s.ls, loss)
				if k > 0.95 && c.Stiff < 0.5 {
					s.golden++
				}
				if k < 0.9 && c.Stiff > 1.5 {
					s.blind++
				}
			}
		}
		n++
	}
	w := Generate(seed, cfg)
	s.peoples = len(w.Civs)
	for _, c := range w.Civs {
		if c.Cause == "heaven" {
			s.heavens++
		}
	}
	for _, e := range w.Events {
		if e.Kind == KFaced && e.P["filter"] == "upload" {
			switch e.P["outcome"] {
			case "overcome":
				s.uploads[0]++
			case "scarred":
				s.uploads[1]++
			case "declined":
				s.uploads[2]++
			}
		}
	}
	return s
}

func pearson(x, y []float64) float64 {
	if len(x) < 2 {
		return math.NaN()
	}
	var mx, my float64
	for i := range x {
		mx += x[i]
		my += y[i]
	}
	mx /= float64(len(x))
	my /= float64(len(y))
	var sxy, sxx, syy float64
	for i := range x {
		sxy += (x[i] - mx) * (y[i] - my)
		sxx += (x[i] - mx) * (x[i] - mx)
		syy += (y[i] - my) * (y[i] - my)
	}
	return sxy / math.Sqrt(sxx*syy)
}
