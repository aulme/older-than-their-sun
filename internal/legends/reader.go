package legends

import (
	"worldgen/internal/record"
	"worldgen/internal/species"
)

// Reader is a loaded run for the batch reports: the same view the
// legends render from, with the few words the reports print exported.
type Reader struct{ v *view }

// Open loads a run for reading.
func Open(r *record.Run) *Reader { return &Reader{load(r)} }

// Text resolves every name token in a text.
func (r *Reader) Text(s string) string { return r.v.names.Text(s) }

// Voice is how a people names: none, transcribed or translated.
func (r *Reader) Voice(id int) record.Voice { return r.v.st.Voices[id] }

// Species is a people's blood as a species.
func (r *Reader) Species(c *record.Civ) *species.Species { return r.v.species(c) }

// CauseText is why a people fell or ended, with the names resolved.
func (r *Reader) CauseText(c *record.Civ) string { return r.Text(r.v.causeText(c)) }

// IntoText is what a transformed people became.
func (r *Reader) IntoText(c *record.Civ) string { return r.Text(r.v.intoText(c)) }

// OriginText is how a people or a blood came to be.
func (r *Reader) OriginText(o species.Making) string { return r.Text(r.v.originText(o)) }

// RecordText is one entry of a people's record in words.
func (r *Reader) RecordText(rec record.Record) string { return r.Text(r.v.recordText(rec)) }

// WarCause is why a war was declared; WarResult how it ended.
func (r *Reader) WarCause(w *record.War) string  { return r.Text(r.v.warCause(w)) }
func (r *Reader) WarResult(w *record.War) string { return r.v.warResult(w) }

// Active says whether a people is still rising; Living whether it lives.
func Active(c *record.Civ) bool { return active(c) }
func Living(c *record.Civ) bool { return living(c) }

// SortOf is a fact's own moral sort.
func SortOf(e *record.Event) string { return sortOf(e) }
