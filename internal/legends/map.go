package legends

import (
	"fmt"
	"io"
	"math"

	"worldgen/internal/galaxy"
)

// Map prints a top-down chart of the galaxy with the Sun and a region
// marked, then a survey of the laws from the centre to the rim along the
// Sun's line, so the reader can see what changes going in or out.
func Map(out io.Writer, rg galaxy.Region) {
	const cols, rows = 65, 33
	const span = 16.0 // kpc each way
	fmt.Fprintf(out, "The Milky Way from above, %d kpc across; the Sun's side at the bottom; rotation is clockwise.\n", int(2*span))
	fmt.Fprintln(out, "  # arm ridge  + arm  . disc  % bar and bulge  * the centre  S the Sun  @ this region  o feature")
	grid := make([][]byte, rows)
	for i := range grid {
		grid[i] = make([]byte, cols)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	// X (toward the Sun) runs down the page: the Sun at X=-8.2 is at the bottom
	cell := func(v galaxy.Vec) (int, int) {
		r := int(math.Round((span - v.X) / (2 * span) * float64(rows-1)))
		c := int(math.Round((v.Y + span) / (2 * span) * float64(cols-1)))
		return r, c
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x := span - float64(r)/float64(rows-1)*2*span
			y := -span + float64(c)/float64(cols-1)*2*span
			v := galaxy.Vec{X: x, Y: y}
			R := v.R()
			if R > 17 {
				continue
			}
			ch := byte(' ')
			d := galaxy.Density(v)
			switch {
			case R < 0.4:
				ch = '*'
			case galaxy.InBar(v) || R < 2.5:
				ch = '%'
			default:
				armw := 0.0
				for _, a := range galaxy.Arms {
					if wt := a.Dist(v); !math.IsInf(wt, 1) && wt < 0.5*a.Width {
						armw = math.Max(armw, 1)
					} else if !math.IsInf(wt, 1) && wt < 1.2*a.Width {
						armw = math.Max(armw, 0.5)
					}
				}
				switch {
				case armw >= 1:
					ch = '#'
				case armw > 0:
					ch = '+'
				case d > 0.08:
					ch = '.'
				}
			}
			grid[r][c] = ch
		}
	}
	for _, f := range galaxy.Features {
		if f.Kind == galaxy.Sky || f.Kind == galaxy.Structure {
			continue
		}
		if r, c := cell(f.Pos); r >= 0 && r < rows && c >= 0 && c < cols && f.Pos.R() > 0.5 {
			grid[r][c] = 'o'
		}
	}
	if r, c := cell(galaxy.Sun); r >= 0 && r < rows {
		grid[r][c] = 'S'
	}
	if r, c := cell(rg.Pos); r >= 0 && r < rows && c >= 0 && c < cols {
		grid[r][c] = '@'
	}
	for _, row := range grid {
		fmt.Fprintf(out, "  %s\n", string(row))
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "This region: %s\n", rg.Describe())
	fmt.Fprintln(out)
	fmt.Fprintln(out, "The laws along the Sun's line, from the heart to the rim (all relative to the Sun's neighbourhood; metals in dex):")
	fmt.Fprintf(out, "  %-9s %-16s %8s %8s %7s %7s %6s %7s %7s %7s %7s\n", "R kpc", "zone", "density", "spacing", "youth", "metals", "rocky", "glare", "hazard", "crowd", "exotic")
	for _, R := range []float64{0.01, 0.1, 0.3, 1, 2, 3, 4.5, 5, 6, 6.7, 7.5, 8.2, 9, 10.1, 11.5, 13.5, 15.5, 18} {
		v := galaxy.Vec{X: -R, Y: 0, Z: 0.02}
		l := galaxy.At(v)
		sp := galaxy.MeanSpacing(400, 150*l.Spacing(), 40*l.Spacing())
		fmt.Fprintf(out, "  %-9.2f %-16s %8.2f %6.0f ly %7.2f %7.2f %6.2f %7.2f %7.2f %7.1f %7.2f\n", R, l.Zone, l.Density, sp, l.Youth, l.Metals, l.Rocky(), l.Glare, l.Hazard(), l.Crowd, l.Exotic)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Named places (-at): ")
	for _, p := range galaxy.Presets {
		l := galaxy.At(p.Pos)
		fmt.Fprintf(out, "  %-12s %-34s %s (density %.1f, youth %.1f, glare %.1f)\n", p.Key, p.Name, p.Desc, l.Density, l.Youth, l.Glare)
	}
	fmt.Fprintln(out, "Any feature by name: ")
	for _, f := range galaxy.Features {
		if f.Kind == galaxy.Sky {
			continue
		}
		fmt.Fprintf(out, "  %-38s %-18s %6.0f ly from Earth\n", f.Name, f.Kind, f.D*galaxy.LyPerKpc)
	}
}
