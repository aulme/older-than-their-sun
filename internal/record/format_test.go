package record

import (
	"reflect"
	"strings"
	"testing"
)

// TestFormatDocumented: every field the writer emits is documented in
// FORMAT.md. Each record type has a section headed by its name, and
// every JSON field of it appears in that section in backticks.
func TestFormatDocumented(t *testing.T) {
	b, err := FormatDoc.ReadFile("FORMAT.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	sections := map[string]string{}
	parts := strings.Split(doc, "\n### ")
	for _, p := range parts[1:] {
		name, body, _ := strings.Cut(p, "\n")
		sections[strings.TrimSpace(name)] = body
	}
	if !strings.Contains(doc, "format "+itoa(Format)) {
		t.Errorf("FORMAT.md does not name the format version %d", Format)
	}
	seen := map[reflect.Type]bool{}
	var walk func(rt reflect.Type)
	walk = func(rt reflect.Type) {
		for rt.Kind() == reflect.Ptr || rt.Kind() == reflect.Slice || rt.Kind() == reflect.Map || rt.Kind() == reflect.Array {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct || seen[rt] {
			return
		}
		seen[rt] = true
		if rt.PkgPath() != reflect.TypeOf(Run{}).PkgPath() {
			return // a type of another package, documented where it is used
		}
		if rt.Name() == "Run" {
			for i := 0; i < rt.NumField(); i++ {
				walk(rt.Field(i).Type)
			}
			return
		}
		body, ok := sections[rt.Name()]
		if !ok {
			t.Errorf("FORMAT.md has no section for %s", rt.Name())
			return
		}
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			tag, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if tag == "" || tag == "-" {
				continue
			}
			if !strings.Contains(body, "`"+tag+"`") {
				t.Errorf("FORMAT.md's %s section does not document `%s`", rt.Name(), tag)
			}
			walk(f.Type)
		}
	}
	walk(reflect.TypeOf(Run{}))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
