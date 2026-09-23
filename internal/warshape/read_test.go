package warshape

import (
	"testing"

	"worldgen/internal/record"
)

// run is a small record: four peoples, the first two large, at war
// twice; the third made the second's vassal by a war; a third war joined
// to the first by the people they share.
func run() *record.Run {
	d := &record.Dossier{Seed: 1, Dawn: 0, Present: 200_000, Step: 1000}
	st := &record.State{}
	for i := range 4 {
		st.Civs = append(st.Civs, &record.Civ{ID: i})
	}
	var ch []*record.Event
	ev := func(y record.Year, k record.Kind, s, o int, p record.P) {
		ch = append(ch, &record.Event{ID: len(ch), Year: y, Kind: k, Subject: s, Object: o, Star: -1, P: p})
	}
	for c, n := range []int{6, 7, 1, 1} {
		for range n {
			ev(0, record.KWorldHeld, c, -1, nil)
		}
	}
	st.Wars = []*record.War{
		{ID: 0, Sides: [2]int{0, 1}, Began: 10_000, Ended: 12_000, Over: true, Nth: 1, Taken: [2]int{1, 0}, Result: "peace"},
		{ID: 1, Sides: [2]int{1, 2}, Began: 11_000, Ended: 30_000, Over: true, Nth: 1, Result: "vassal"},
		{ID: 2, Sides: [2]int{0, 1}, Began: 100_000, Ended: 101_000, Over: true, Nth: 2},
	}
	ev(30_000, record.KWarOver, 1, 2, record.P{"result": "vassal"})
	ev(30_000, record.KMaster, 2, -1, record.P{"master": 1, "vassal": true})
	st.Battles = []record.Battle{{Year: 10_500, Attacker: 0, Defender: 1}, {Year: 20_000, Attacker: 1, Defender: 2}, {Year: 150_000, Attacker: 0, Defender: 1}}
	st.Fleets = []*record.Fleet{{Owner: 1, Target: 0, Kind: "campaign", Launched: 10_200}}
	return &record.Run{Dossier: d, State: st, Chronicle: ch}
}

func TestRead(t *testing.T) {
	s := Read(run())
	w0, w1 := s.Wars[0], s.Wars[1]
	if w0.Battles != 1 || w0.Ticks != 2 || w0.Campaigns != [2]int{0, 1} || w0.Worlds != [2]int{6, 7} {
		t.Errorf("first war read as %+v", w0)
	}
	if w1.Battles != 1 || w1.Ticks != 19 || w1.BattleTicks != 1 {
		t.Errorf("second war read as %+v", w1)
	}
	if s.Stray != 1 {
		t.Errorf("%d battles in no war, want the one at 150 kyr", s.Stray)
	}
	if len(s.Systems) != 2 || s.Systems[0].Peak != 3 || s.Systems[0].Fronts != 2 {
		t.Errorf("systems %+v, want the first two wars joined by the second people", s.Systems)
	}
	if len(s.Bonds) != 1 || s.Bonds[0].Origin != "war" || !s.Bonds[0].Vassal || s.Bonds[0].Years != 170_000 {
		t.Errorf("bonds %+v, want one vassal by war held to the present", s.Bonds)
	}
	// The two large peoples are at peace from 12 kyr to the end, but for a
	// short war at 100 kyr and a battle at 150: a cold war.
	if s.Cold != 1 {
		t.Errorf("%d cold wars, want 1", s.Cold)
	}
}

func TestAtIsBefore(t *testing.T) {
	h := []span{{10, 1}, {20, 2}}
	for _, c := range []struct {
		y    record.Year
		want int
	}{{5, 0}, {10, 0}, {11, 1}, {20, 1}, {21, 2}} {
		if got := at(h, c.y, 0); got != c.want {
			t.Errorf("at %d: %d, want %d", c.y, got, c.want)
		}
	}
}
