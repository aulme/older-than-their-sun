package record

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"worldgen/codex"
	"worldgen/data"
)

// FormatDoc is FORMAT.md, the contract, kept beside these types and
// copied into every run directory.
//
//go:embed FORMAT.md
var FormatDoc embed.FS

// The files of a run directory.
const (
	DossierFile   = "dossier.json"
	StateFile     = "state.json"
	ChronicleFile = "chronicle.jsonl"
	TellingsFile  = "tellings.jsonl"
)

// Write writes a run to a directory: the four files, and copies of the
// lookups, the codex and the contract, so the directory is
// self-contained. The directory is made if it is not there.
func Write(dir string, r *Run) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, DossierFile), r.Dossier); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, StateFile), r.State); err != nil {
		return err
	}
	if err := writeLines(filepath.Join(dir, ChronicleFile), len(r.Chronicle), func(i int) any { return r.Chronicle[i] }); err != nil {
		return err
	}
	if err := writeLines(filepath.Join(dir, TellingsFile), len(r.Tellings), func(i int) any { return r.Tellings[i] }); err != nil {
		return err
	}
	if err := copyFS(data.FS, filepath.Join(dir, "data")); err != nil {
		return err
	}
	if err := copyFS(codex.FS, filepath.Join(dir, "codex")); err != nil {
		return err
	}
	return copyFS(FormatDoc, dir)
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		f.Close()
		return err
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func writeLines(path string, n int, at func(int) any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)
	for i := 0; i < n; i++ {
		if err := enc.Encode(at(i)); err != nil {
			f.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func copyFS(src fs.FS, dir string) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		out := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
}

// Load reads a run directory back.
func Load(dir string) (*Run, error) {
	r := &Run{Dossier: &Dossier{}, State: &State{}}
	if err := readJSON(filepath.Join(dir, DossierFile), r.Dossier); err != nil {
		return nil, err
	}
	if r.Dossier.Format != Format {
		return nil, fmt.Errorf("record: %s is format %d, this build reads %d", dir, r.Dossier.Format, Format)
	}
	if err := readJSON(filepath.Join(dir, StateFile), r.State); err != nil {
		return nil, err
	}
	if err := readLines(filepath.Join(dir, ChronicleFile), func(dec *json.Decoder) error {
		e := &Event{}
		if err := dec.Decode(e); err != nil {
			return err
		}
		r.Chronicle = append(r.Chronicle, e)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := readLines(filepath.Join(dir, TellingsFile), func(dec *json.Decoder) error {
		t := &Tale{}
		if err := dec.Decode(t); err != nil {
			return err
		}
		r.Tellings = append(r.Tellings, t)
		return nil
	}); err != nil {
		return nil, err
	}
	return r, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func readLines(path string, each func(*json.Decoder) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(bufio.NewReaderSize(f, 1<<20))
	for dec.More() {
		if err := each(dec); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}

// Roundtrip passes a run through JSON and back, so that a run built in
// memory holds exactly what one read from a directory would.
func Roundtrip(r *Run) *Run {
	b, err := json.Marshal(r)
	if err != nil {
		panic("record: " + err.Error())
	}
	out := &Run{}
	if err := json.Unmarshal(b, out); err != nil {
		panic("record: " + err.Error())
	}
	return out
}
