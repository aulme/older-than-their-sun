package history

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"
)

// TestDeclineShape samples the galaxy through a whole age and writes one
// row per sample, so the candidate variables of the decline index
// (specs/proposals/decline.md, "The index") can be read against each
// other over the same runs rather than argued about. It measures; it
// gates nothing, and the sampling draws nothing, so the ages it walks
// are the ages an unsampled run walks.
//
//	DECLINE=/some/dir go test ./internal/history -run TestDeclineShape -v -timeout 4h
//
// One file per seed, `decline-<seed>.csv`, with a header naming the
// galaxy's fixed counts.
func TestDeclineShape(t *testing.T) {
	dir := os.Getenv("DECLINE")
	if dir == "" {
		t.Skip("a measurement, not a gate: set DECLINE=<dir> to run it")
	}
	stars := 400
	if s := os.Getenv("DECLINE_STARS"); s != "" {
		stars, _ = strconv.Atoi(s)
	}
	seeds := []uint64{1, 2, 3, 4, 5, 6, 7, 8}
	if s := os.Getenv("DECLINE_SEEDS"); s != "" {
		seeds = nil
		n, _ := strconv.Atoi(s)
		for i := 1; i <= n; i++ {
			seeds = append(seeds, uint64(i))
		}
	}
	var wg sync.WaitGroup
	for _, seed := range seeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sampleAge(t, dir, seed, stars)
		}()
	}
	wg.Wait()
}

// sampleAge runs one age with a sampler attached and writes its rows.
func sampleAge(t *testing.T, dir string, seed uint64, stars int) {
	cfg := DefaultConfig()
	cfg.Stars = stars
	var rows []string
	every := 100 // ticks: a sample every hundred thousand years
	n := 0
	var was []int // the owners at the last sample, for the churn
	cfg.Sample = func(w *World) {
		if n%every == 0 {
			rows = append(rows, declineRow(w, &was))
		}
		n++
	}
	w := Generate(seed, cfg)
	hab := 0
	for i := range w.G.Stars {
		if w.G.Stars[i].Hab > 0 {
			hab++
		}
	}
	var b []byte
	b = fmt.Appendf(b, "# seed %d stars %d habitable %d present %d waning %d capped %v peoples %d\n",
		seed, len(w.G.Stars), hab, w.Present, w.Waning, w.Capped, len(w.Civs))
	b = append(b, "year,held,habheld,active,rising,remnant,dead,ossified,asleep,ruled,born,medstiff,reach,fertility,hazard,thin,changed,medage,oldheld,big,bigage,wars\n"...)
	for _, r := range rows {
		b = append(b, r...)
		b = append(b, '\n')
	}
	// the present itself, after the loop has left it
	b = append(b, declineRow(w, &was)...)
	b = append(b, '\n')
	if err := os.WriteFile(fmt.Sprintf("%s/decline-%d.csv", dir, seed), b, 0o644); err != nil {
		t.Error(err)
	}
	t.Logf("seed %d: present %.1f Myr, waning %.1f Myr, %d peoples, %d rows",
		seed, float64(w.Present)/1e6, float64(w.Waning)/1e6, len(w.Civs), len(rows))
}

// declineRow is one sample: everything the index might be built from.
func declineRow(w *World, was *[]int) string {
	held, habheld := 0, 0
	for i, o := range w.Owner {
		if o < 0 || !w.Civs[o].Active() {
			continue
		}
		held++
		if w.G.Stars[i].Hab > 0 {
			habheld++
		}
	}
	var active, rising, remnant, dead, ossified, asleep, ruled, born int
	var stiffs []float64
	reached := make([]bool, len(w.G.Stars))
	for _, c := range w.Civs {
		if c.Born <= w.Now {
			born++
		}
		switch {
		case c.Stage == Dead:
			dead++
		case !c.Active():
			remnant++
		default:
			active++
			stiffs = append(stiffs, c.Stiff)
			if c.Ossified {
				ossified++
			}
			if c.Asleep {
				asleep++
			}
			if c.Master >= 0 {
				ruled++
			}
			if !c.Ossified && c.Stiff < 1 {
				rising++
			}
			if c.Reach >= 1 {
				for s := range w.G.Stars {
					if !reached[s] && w.G.Dist(c.Home, s) <= c.Reach {
						reached[s] = true
					}
				}
			}
		}
	}
	reach := 0
	for _, r := range reached {
		if r {
			reach++
		}
	}
	med := 0.0
	if len(stiffs) > 0 {
		sort.Float64s(stiffs)
		med = stiffs[len(stiffs)/2]
	}
	// the churn: stars that changed hands since the last sample, which
	// says whether a flat occupancy is a still map or a boiling one
	changed := 0
	if len(*was) == len(w.Owner) {
		for i, o := range w.Owner {
			if o != (*was)[i] {
				changed++
			}
		}
	}
	*was = append((*was)[:0], w.Owner...)
	// the age of what stands, and how much of the map the old hold
	var ages []float64
	oldheld, big, bigage := 0, 0, 0.0
	for _, c := range w.Civs {
		if !c.Active() {
			continue
		}
		ages = append(ages, float64(w.Now-c.Born)/1e6)
		n := 0
		for _, s := range c.Systems {
			if w.G.Stars[s].Hab > 0 {
				n++
			}
		}
		if w.Now-c.Born > 5_000_000 {
			oldheld += n
		}
		if len(c.Systems) > big {
			big, bigage = len(c.Systems), float64(w.Now-c.Born)/1e6
		}
	}
	medage := 0.0
	if len(ages) > 0 {
		sort.Float64s(ages)
		medage = ages[len(ages)/2]
	}
	wars := 0
	for _, war := range w.Wars {
		if !war.Over {
			wars++
		}
	}
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.3f,%d,%.4f,%.3f,%.3f,%d,%.2f,%d,%d,%.2f,%d",
		w.Now, held, habheld, active, rising, remnant, dead, ossified, asleep, ruled, born,
		med, reach, w.fertility(), w.Hazard, w.Thin,
		changed, medage, oldheld, big, bigage, wars)
}
