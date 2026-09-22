package legends

import (
	"encoding/json"
	"fmt"
	"strings"

	"worldgen/internal/record"
)

// The words for numbers: counts, spans, shares, as the legends say them.

// year prints a year relative to the present.
func (v *view) year(y Year) string {
	y -= v.present
	switch {
	case y == 0:
		return "present"
	case y <= -1_000_000_000:
		return fmt.Sprintf("%.2f Gyr ago", -float64(y)/1e9)
	case y <= -1_000_000:
		return fmt.Sprintf("%.2f Myr ago", -float64(y)/1e6)
	default:
		return fmt.Sprintf("%s ago", commas(-int64(y)))
	}
}

func commas(n int64) string {
	s := fmt.Sprint(n)
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	return out
}

func systems(n int) string {
	if n == 1 {
		return "a single world"
	}
	return sprintf("%d systems", n)
}

// shipsWord says a count of ships.
func shipsWord(n int) string {
	switch n {
	case 1:
		return "one ship"
	case 2:
		return "two ships"
	case 3:
		return "three ships"
	}
	return sprintf("%d ships", n)
}

func span(y Year) string {
	switch {
	case y < 1000:
		return "a few centuries"
	case y < 2000:
		return "a thousand years"
	default:
		return sprintf("%d thousand years", y/1000)
	}
}

func worlds(n int) string {
	switch n {
	case 0:
		return "no worlds"
	case 1:
		return "one world"
	}
	return sprintf("%d worlds", n)
}

func ordinal(n int) string {
	switch n {
	case 2:
		return "second"
	case 3:
		return "third"
	case 4:
		return "fourth"
	case 5:
		return "fifth"
	}
	return sprintf("%dth", n)
}

func warsOf(n int) string {
	switch n {
	case 1:
		return "a war"
	case 2:
		return "two wars"
	}
	return sprintf("%d wars", n)
}

// warSpan says how long a war ran and what it cost, from those.
func warSpan(e *record.Event) string {
	if e.Bool("short") {
		return sprintf("a short war; %s changed hands or burned", worlds(e.Int("gone")))
	}
	return sprintf("%s of war; %s changed hands or burned", span(e.YearOf("years")), worlds(e.Int("gone")))
}

func shareWord(n, total int) string {
	switch f := float64(n) / float64(max(total, 1)); {
	case f >= 0.6:
		return "most"
	case f >= 0.4:
		return "half"
	case f >= 0.25:
		return "a third"
	}
	return "a part"
}

// depthWord says a depth: a tenth, a fifth, a third, half, most.
func depthWord(d float64) string {
	switch {
	case d < 0.15:
		return "a tenth"
	case d < 0.25:
		return "a fifth"
	case d < 0.4:
		return "a third"
	case d < 0.6:
		return "half"
	}
	return "most"
}

// partWord says a fraction in the register the legends use rather than
// as a number.
func partWord(x float64) string {
	switch {
	case x < 0.02:
		return "almost none"
	case x < 0.08:
		return "a twentieth"
	case x < 0.15:
		return "a tenth"
	case x < 0.25:
		return "a fifth"
	case x < 0.4:
		return "a third"
	case x < 0.6:
		return "half"
	case x < 0.9:
		return "most"
	}
	return "almost all"
}

func percent(x float64) string {
	if x < 0.01 {
		return "less than a hundredth"
	}
	return sprintf("%.0f%%", x*100)
}

// list joins names as prose: "a, b and c".
func list(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " and " + xs[len(xs)-1]
}

// capital raises the first letter of a name.
func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func toInt(x any) int {
	switch n := x.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

func toFloat(x any) float64 {
	switch n := x.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func toStrs(x any) []string {
	switch xs := x.(type) {
	case []string:
		return xs
	case []any:
		out := make([]string, 0, len(xs))
		for _, s := range xs {
			str, _ := s.(string)
			out = append(out, str)
		}
		return out
	}
	return nil
}
