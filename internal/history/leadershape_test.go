package history

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"

	"worldgen/internal/mind"
)

// TestLeaderShape is the measurement behind step 9's gate (specs/plan.md;
// specs/proposals/leaders.md): over a batch of whole ages it reads every
// leader that rose — its occasion, form, stance against its people's own,
// temperament and end — and every succession faced, by the continuity
// its realm had when it was faced. The gate is that the three outcomes
// are all there and that a lost leader costs what continuity says: the
// less a realm keeps of its past, the more often the succession fails.
//
//	LEADERS=1 go test ./internal/history -run TestLeaderShape -v -timeout 4h
//
// LEADERS_STARS (400) and LEADERS_SEEDS (10) set the batch.
func TestLeaderShape(t *testing.T) {
	if os.Getenv("LEADERS") == "" {
		t.Skip("a measurement, not a gate: set LEADERS=1 to run it")
	}
	stars, seeds := 400, 10
	if s := os.Getenv("LEADERS_STARS"); s != "" {
		stars, _ = strconv.Atoi(s)
	}
	if s := os.Getenv("LEADERS_SEEDS"); s != "" {
		seeds, _ = strconv.Atoi(s)
	}
	out := make([]*World, seeds)
	var wg sync.WaitGroup
	for i := range seeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := DefaultConfig()
			cfg.Stars = stars
			out[i] = Generate(uint64(i+1), cfg)
		}()
	}
	wg.Wait()
	var all []*Leader
	for i, w := range out {
		rose, deathless, mad, ends := LeadersOf(w)
		t.Logf("seed %d: %d peoples, %d leaders (%d deathless, %d mad), ends %v; %d wars, %d battles, %d meetings", i+1, len(w.Civs), rose, deathless, mad, ends, len(w.Wars), len(w.Battles), len(w.Meetings))
		all = append(all, w.Leaders...)
	}
	for _, l := range leaderReport(all) {
		t.Log(l)
	}
}

// leaderReport is the batch read: counts by occasion, form, place and
// end; how many were turned against their own bent; and the succession
// by the continuity band it was faced in.
func leaderReport(ls []*Leader) []string {
	var out []string
	count := func(key func(*Leader) string) string {
		m := map[string]int{}
		for _, l := range ls {
			m[key(l)]++
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		s := ""
		for _, k := range keys {
			s += fmt.Sprintf(" %s %d (%.0f%%)", k, m[k], 100*float64(m[k])/float64(len(ls)))
		}
		return s
	}
	out = append(out, fmt.Sprintf("%d leaders", len(ls)))
	out = append(out, "by occasion:"+count(func(l *Leader) string { return l.Occasion }))
	out = append(out, "by form:"+count(func(l *Leader) string { return l.Form }))
	out = append(out, "by place:"+count(func(l *Leader) string {
		if l.Front {
			return "front"
		}
		return "capital"
	}))
	out = append(out, "runs rather than falls:"+count(func(l *Leader) string { return strconv.FormatBool(l.Flees) }))
	out = append(out, "by end:"+count(func(l *Leader) string { return l.End }))
	out = append(out, "turned (stance on the other side of war from the people's own):"+count(func(l *Leader) string {
		ro, rs := rungOf(l.Own), rungOf(l.Stance)
		switch {
		case l.Stance == l.Own:
			return "same"
		case mind.Hostile(l.Stance) != mind.Hostile(l.Own):
			return "inverted"
		case (rs-2)*(rs-2) > (ro-2)*(ro-2):
			return "amplified"
		}
		return "tempered"
	}))
	out = append(out, "front leaders by end:"+count(func(l *Leader) string {
		if !l.Front {
			return "(capital)"
		}
		return l.End
	}))
	battles, fell := 0, 0
	for _, l := range ls {
		battles += l.Battles
		if l.End == "fell_field" {
			fell++
		}
	}
	out = append(out, fmt.Sprintf("battles fought at a leader's side: %d; leaders fallen with the field: %d", battles, fell))
	out = append(out, "leaders' pushes: median "+fmt.Sprintf("%.2f", median(func() []float64 {
		var p []float64
		for _, l := range ls {
			p = append(p, l.Push)
		}
		return p
	}())))
	// reign by place and form
	reign := map[string][]float64{}
	for _, l := range ls {
		if l.End == "" || l.End == "with_people" || l.End == "deposed" {
			continue
		}
		k := "capital"
		if l.Front {
			k = "front"
		}
		reign[k] = append(reign[k], float64(l.Ended-l.Rose)/1000)
	}
	for _, k := range []string{"front", "capital"} {
		out = append(out, fmt.Sprintf("reign at the %s, ended by a loss: %d, median %.1f kyr", k, len(reign[k]), median(reign[k])))
	}
	// succession by continuity band
	bands := []float64{-1, 0, 1, 2}
	type row struct{ n, over, scar, dec int }
	rows := make([]row, len(bands)+1)
	faced := 0
	for _, l := range ls {
		if l.Faced == "" {
			continue
		}
		faced++
		b := sort.SearchFloat64s(bands, l.Doublings)
		r := &rows[b]
		r.n++
		switch l.Faced {
		case "overcome":
			r.over++
		case "scarred":
			r.scar++
		default:
			r.dec++
		}
	}
	out = append(out, fmt.Sprintf("successions faced: %d", faced))
	for i, r := range rows {
		lo, hi := math.Inf(-1), math.Inf(1)
		if i > 0 {
			lo = bands[i-1]
		}
		if i < len(bands) {
			hi = bands[i]
		}
		if r.n == 0 {
			out = append(out, fmt.Sprintf("  doublings %v to %v: none", lo, hi))
			continue
		}
		out = append(out, fmt.Sprintf("  doublings %v to %v: %d faced, overcome %.0f%%, scarred %.0f%%, declined %.0f%%", lo, hi, r.n,
			100*float64(r.over)/float64(r.n), 100*float64(r.scar)/float64(r.n), 100*float64(r.dec)/float64(r.n)))
	}
	return out
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	return s[len(s)/2]
}
