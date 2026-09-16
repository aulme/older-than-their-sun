// Command mkcatalog builds internal/galaxy/catalog_data.go from two public
// datasets: the HYG star database (astronexus.com, CC BY-SA 4.0) and the
// NASA Exoplanet Archive's composite planet table. It keeps every star
// within 150 ly that has a proper name, hosts a known planet, or lies
// within 25 ly, and attaches the known planets to their hosts.
//
//	curl -o hyg.csv https://raw.githubusercontent.com/astronexus/HYG-Database/main/hyg/CURRENT/hygdata_v41.csv
//	curl -o planets.csv "https://exoplanetarchive.ipac.caltech.edu/TAP/sync?query=select+pl_name,hostname,sy_dist,sy_snum,sy_pnum,pl_bmasse,pl_rade,pl_orbper,pl_orbsmax,pl_eqt,pl_orbeccen,st_spectype,st_teff,st_mass,st_rad,st_age,st_met,ra,dec,discoverymethod,disc_year+from+pscomppars+where+sy_dist<50&format=csv"
//	go run ./cmd/mkcatalog -hyg hyg.csv -planets planets.csv > internal/galaxy/catalog_data.go
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const lyPerPc = 3.26156
const maxLy = 150.0

type planet struct {
	Name                                        string
	MassE, RadE, Period, SMA, Teq, Ecc          float64
	Method                                      string
	Year                                        int
	ra, dec, dist, stTeff, stMass, stAge, stMet float64
	vmag                                        float64
	spect                                       string
	snum                                        int
}

type star struct {
	id                 string
	Name, Alt, Spect   string
	L, B, Dist         float64 // deg, deg, ly
	Mag, Lum           float64
	Var                string
	Mult               int
	Comps              []string
	Planets            []planet
	StMass, StAge, Met float64
	Teff               float64
	ra, dec            float64
	near               bool
}

func f(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// galactic converts J2000 equatorial degrees to galactic longitude and latitude.
func galactic(raDeg, decDeg float64) (float64, float64) {
	const d = math.Pi / 180
	aG, dG, lNCP := 192.85948*d, 27.12825*d, 122.93192*d
	a, dc := raDeg*d, decDeg*d
	sb := math.Sin(dc)*math.Sin(dG) + math.Cos(dc)*math.Cos(dG)*math.Cos(a-aG)
	b := math.Asin(sb)
	y := math.Cos(dc) * math.Sin(a-aG)
	x := math.Sin(dc)*math.Cos(dG) - math.Cos(dc)*math.Sin(dG)*math.Cos(a-aG)
	l := lNCP - math.Atan2(y, x)
	for l < 0 {
		l += 2 * math.Pi
	}
	for l >= 2*math.Pi {
		l -= 2 * math.Pi
	}
	return l / d, b / d
}

func angSep(ra1, dec1, ra2, dec2 float64) float64 {
	const d = math.Pi / 180
	c := math.Sin(dec1*d)*math.Sin(dec2*d) + math.Cos(dec1*d)*math.Cos(dec2*d)*math.Cos((ra1-ra2)*d)
	return math.Acos(math.Min(1, math.Max(-1, c))) / d
}

func main() {
	hygPath := flag.String("hyg", "hyg.csv", "HYG csv")
	plPath := flag.String("planets", "planets.csv", "exoplanet archive csv")
	flag.Parse()

	rows := readCSV(*hygPath)
	col := index(rows[0])
	type hrow struct {
		id, primary, proper, bf, gl, hd, hip, spect, comp, vr string
		ra, dec, dist, mag, lum                               float64
	}
	var hs []hrow
	byID := map[string]*hrow{}
	for _, r := range rows[1:] {
		d := f(r[col["dist"]])
		if d <= 0 || d*lyPerPc > maxLy+5 {
			continue
		}
		h := hrow{id: r[col["id"]], primary: r[col["comp_primary"]], proper: r[col["proper"]], bf: r[col["bf"]], gl: r[col["gl"]],
			hd: r[col["hd"]], hip: r[col["hip"]], spect: r[col["spect"]], comp: r[col["comp"]], vr: r[col["var"]],
			ra: f(r[col["ra"]]) * 15, dec: f(r[col["dec"]]), dist: d * lyPerPc, mag: f(r[col["mag"]]), lum: f(r[col["lum"]])}
		hs = append(hs, h)
	}
	for i := range hs {
		byID[hs[i].id] = &hs[i]
	}
	// group by primary
	groups := map[string][]*hrow{}
	for i := range hs {
		p := hs[i].primary
		if p == "" || byID[p] == nil {
			p = hs[i].id
		}
		groups[p] = append(groups[p], &hs[i])
	}
	stars := map[string]*star{}
	for pid, g := range groups {
		sort.Slice(g, func(i, j int) bool { return g[i].comp < g[j].comp })
		p := byID[pid]
		if p == nil {
			p = g[0]
		}
		s := &star{id: pid, Dist: p.dist, Mag: p.mag, Lum: p.lum, Var: p.vr, ra: p.ra, dec: p.dec, Spect: p.spect}
		s.L, s.B = galactic(p.ra, p.dec)
		s.Name = p.proper
		alts := []string{}
		if p.bf != "" {
			alts = append(alts, cleanBF(p.bf))
		}
		if p.gl != "" {
			alts = append(alts, strings.Replace(strings.Replace(p.gl, "Gl ", "GJ ", 1), "GJ  ", "GJ ", 1))
		}
		if p.hd != "" {
			alts = append(alts, "HD "+p.hd)
		}
		if p.hip != "" {
			alts = append(alts, "HIP "+p.hip)
		}
		if s.Name == "" && len(alts) > 0 {
			s.Name, alts = alts[0], alts[1:]
		}
		if len(alts) > 0 {
			s.Alt = alts[0]
		}
		s.Mult = min(3, len(g))
		for _, c := range g {
			if c != p && c.spect != "" {
				s.Comps = append(s.Comps, c.spect)
			}
		}
		s.near = s.Dist <= 25
		stars[pid] = s
	}
	// planets
	prow := readCSV(*plPath)
	pc := index(prow[0])
	hosts := map[string][]planet{}
	for _, r := range prow[1:] {
		p := planet{Name: r[pc["pl_name"]], MassE: f(r[pc["pl_bmasse"]]), RadE: f(r[pc["pl_rade"]]), Period: f(r[pc["pl_orbper"]]),
			SMA: f(r[pc["pl_orbsmax"]]), Teq: f(r[pc["pl_eqt"]]), Ecc: f(r[pc["pl_orbeccen"]]), Method: r[pc["discoverymethod"]],
			Year: int(f(r[pc["disc_year"]])), ra: f(r[pc["ra"]]), dec: f(r[pc["dec"]]), dist: f(r[pc["sy_dist"]]) * lyPerPc,
			stTeff: f(r[pc["st_teff"]]), stMass: f(r[pc["st_mass"]]), stAge: f(r[pc["st_age"]]), stMet: f(r[pc["st_met"]]),
			spect: r[pc["st_spectype"]], snum: int(f(r[pc["sy_snum"]])), vmag: f(r[pc["sy_vmag"]])}
		if p.vmag == 0 {
			p.vmag = 99
		}
		hosts[r[pc["hostname"]]] = append(hosts[r[pc["hostname"]]], p)
	}
	added := 0
	for host, ps := range hosts {
		p0 := ps[0]
		if p0.dist > maxLy {
			continue
		}
		// nearest HYG row by position, then its group
		var best *hrow
		bd := 0.25
		for i := range hs {
			h := &hs[i]
			if math.Abs(h.dist-p0.dist)/p0.dist > 0.2 {
				continue
			}
			if d := angSep(h.ra, h.dec, p0.ra, p0.dec); d < bd {
				best, bd = h, d
			}
		}
		var s *star
		if best != nil {
			pid := best.primary
			if pid == "" || byID[pid] == nil {
				pid = best.id
			}
			s = stars[pid]
		}
		if s == nil {
			l, b := galactic(p0.ra, p0.dec)
			s = &star{id: "pl:" + host, Name: host, Spect: p0.spect, L: l, B: b, Dist: p0.dist, Mult: max(1, p0.snum), ra: p0.ra, dec: p0.dec, Mag: p0.vmag}
			stars[s.id] = s
			added++
		} else if s.Alt == "" && s.Name != host && !strings.EqualFold(s.Name, host) {
			s.Alt = host
		}
		sort.Slice(ps, func(i, j int) bool { return ps[i].SMA < ps[j].SMA })
		s.Planets = append(s.Planets, ps...)
		s.StMass, s.StAge, s.Met, s.Teff = p0.stMass, p0.stAge, p0.stMet, p0.stTeff
	}
	var out []*star
	for _, s := range stars {
		if s.Dist > maxLy {
			continue
		}
		if s.Name != "" && (isProper(s.Name) || len(s.Planets) > 0 || s.near) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dist < out[j].Dist })
	fmt.Fprintf(os.Stderr, "%d stars (%d planet hosts added from the archive alone)\n", len(out), added)

	fmt.Println("// Code generated by cmd/mkcatalog from the HYG database (CC BY-SA 4.0) and the NASA Exoplanet Archive; DO NOT EDIT.")
	fmt.Println()
	fmt.Println("package galaxy")
	fmt.Println()
	fmt.Println("var catalogStars = []CatStar{")
	for _, s := range out {
		fmt.Printf("\t{Name: %q, Alt: %q, Spect: %q, L: %.3f, B: %.3f, Dist: %.2f, Mag: %.2f, Lum: %.3g, Var: %q, Mult: %d, Comps: %s, Mass: %.2f, Age: %.1f, Met: %.2f, Teff: %.0f",
			cleanBF(s.Name), cleanBF(s.Alt), s.Spect, s.L, s.B, s.Dist, s.Mag, s.Lum, s.Var, s.Mult, strs(s.Comps), s.StMass, s.StAge, s.Met, s.Teff)
		if len(s.Planets) > 0 {
			fmt.Printf(", Planets: []CatPlanet{")
			for i, p := range s.Planets {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("{%q, %.3g, %.3g, %.4g, %.4g, %.0f, %.2f, %q, %d}", p.Name, p.MassE, p.RadE, p.Period, p.SMA, p.Teq, p.Ecc, p.Method, p.Year)
			}
			fmt.Printf("}")
		}
		fmt.Println("},")
	}
	fmt.Println("}")
}

// isProper: HYG proper names are IAU or traditional; Bayer/Flamsteed and
// catalogue designations are not proper names.
func isProper(n string) bool {
	for _, p := range []string{"HD ", "HIP ", "GJ ", "Gl ", "NN ", "Wo ", "LHS", "LP ", "G ", "LTT", "BD", "CD", "L "} {
		if strings.HasPrefix(n, p) {
			return false
		}
	}
	return !isBF(n)
}

var greek = map[string]string{"Alp": "Alpha", "Bet": "Beta", "Gam": "Gamma", "Del": "Delta", "Eps": "Epsilon", "Zet": "Zeta", "Eta": "Eta", "The": "Theta",
	"Iot": "Iota", "Kap": "Kappa", "Lam": "Lambda", "Mu": "Mu", "Nu": "Nu", "Xi": "Xi", "Omi": "Omicron", "Pi": "Pi", "Rho": "Rho", "Sig": "Sigma",
	"Tau": "Tau", "Ups": "Upsilon", "Phi": "Phi", "Chi": "Chi", "Psi": "Psi", "Ome": "Omega"}

var bfRe = regexp.MustCompile(`^(\d*)([A-Z][a-z]{1,2})(\d?)\s*([A-Z][A-Za-z]{2})\b(.*)$`)

// cleanBF turns HYG's "9Alp CMa" into "Alpha CMa", "Zet2Ret" into "Zeta2 Ret"; other names pass through.
func cleanBF(bf string) string {
	bf = strings.TrimSpace(bf)
	if m := bfRe.FindStringSubmatch(bf); m != nil {
		if g, ok := greek[m[2]]; ok {
			return strings.TrimSpace(g + m[3] + " " + m[4] + m[5])
		}
	}
	i := 0
	for i < len(bf) && bf[i] >= '0' && bf[i] <= '9' {
		i++
	}
	num, rest := bf[:i], strings.TrimSpace(bf[i:])
	parts := strings.Fields(rest)
	if len(parts) >= 2 {
		if g, ok := greek[parts[0]]; ok {
			return strings.Join(append([]string{g}, parts[1:]...), " ")
		}
		if num != "" {
			return num + " " + rest
		}
	}
	if num != "" {
		return num + " " + rest
	}
	return rest
}

func isBF(n string) bool {
	parts := strings.Fields(n)
	if len(parts) < 2 {
		return false
	}
	if _, ok := greek[parts[0]]; ok {
		return true
	}
	for _, g := range greek {
		if parts[0] == g {
			return true
		}
	}
	_, err := strconv.Atoi(parts[0])
	return err == nil
}

func strs(xs []string) string {
	if len(xs) == 0 {
		return "nil"
	}
	q := make([]string, len(xs))
	for i, x := range xs {
		q[i] = strconv.Quote(x)
	}
	return "[]string{" + strings.Join(q, ", ") + "}"
}

func readCSV(path string) [][]string {
	fh, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	rd := csv.NewReader(fh)
	rd.FieldsPerRecord = -1
	rd.LazyQuotes = true
	rows, err := rd.ReadAll()
	if err != nil {
		panic(err)
	}
	return rows
}

func index(h []string) map[string]int {
	m := map[string]int{}
	for i, c := range h {
		m[strings.Trim(c, "\"")] = i
	}
	return m
}
