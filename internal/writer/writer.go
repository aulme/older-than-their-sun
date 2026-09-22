// Package writer turns a simulated history into a run directory: the
// records the history exports, the names the names pass coins over
// them, and the code revision, written as record.Write lays them out.
// It is the one place the simulation, the names pass and the record
// meet; everything downstream reads the directory.
package writer

import (
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"

	"worldgen/internal/history"
	"worldgen/internal/names"
	"worldgen/internal/record"
)

// Run is a world as records: the export, with the names table and the
// voices filled in and the code revision set. The run holds only what
// JSON holds, so a reader of the run in memory sees exactly what a
// reader of the files would.
func Run(w *history.World, at string) *record.Run {
	r := w.Export(at)
	book := names.Of(w)
	r.State.Names = book.All()
	if r.State.Names == nil {
		r.State.Names = []record.Name{}
	}
	for _, c := range w.Civs {
		r.State.Voices[c.ID] = book.Voice(c)
	}
	r.Dossier.Counts.Names = len(r.State.Names)
	r.Dossier.Code = revision()
	return record.Roundtrip(r)
}

// Write writes a world to a directory.
func Write(dir string, w *history.World, at string) error {
	return record.Write(dir, Run(w, at))
}

var (
	revOnce sync.Once
	rev     string
)

// revision is the git revision the binary was built from, with "-dirty"
// when the tree had uncommitted changes: the build's stamp when there is
// one, else the working tree's, else "unknown".
func revision() string {
	revOnce.Do(func() {
		rev = "unknown"
		if info, ok := debug.ReadBuildInfo(); ok {
			r, dirty := "", false
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					r = s.Value
				case "vcs.modified":
					dirty = s.Value == "true"
				}
			}
			if r != "" {
				rev = r[:min(12, len(r))]
				if dirty {
					rev += "-dirty"
				}
				return
			}
		}
		out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
		if err != nil {
			return
		}
		rev = strings.TrimSpace(string(out))
		if st, err := exec.Command("git", "status", "--porcelain").Output(); err == nil && len(strings.TrimSpace(string(st))) > 0 {
			rev += "-dirty"
		}
	})
	return rev
}
