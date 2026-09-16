package galaxy

import "math"

// Features are the named things of the real galaxy: the great hole at the
// centre, the arms and the bar, the nurseries and the clusters, the dead
// stars and the ones about to die, the globular clusters of the halo, and
// the two clouds beyond the rim. A region of the galaxy is described by
// what it lies near, and the laws of the place are bent by them. Beyond
// their reach they are only sky.
//
// Positions are heliocentric galactic coordinates (l, b in degrees, d in
// kpc) as catalogued, converted at start. Distances are the commonly
// quoted ones; a few are argued over by a factor of two. It does not
// matter here.

type FeatureKind int

const (
	BlackHole FeatureKind = iota
	NeutronStar
	Magnetar
	Remnant // a supernova remnant
	Nebula  // a star-forming region
	Cluster // a young massive cluster or association
	Globular
	Giant     // a massive star near its end
	Structure // a large-scale feature: bar, bubbles, waves, rings
	Sky       // beyond the galaxy: only ever seen
)

var kindNames = map[FeatureKind]string{BlackHole: "black hole", NeutronStar: "neutron star", Magnetar: "magnetar", Remnant: "supernova remnant",
	Nebula: "nebula", Cluster: "cluster", Globular: "globular cluster", Giant: "doomed giant", Structure: "structure", Sky: "beyond the galaxy"}

func (k FeatureKind) String() string { return kindNames[k] }

// Feature is one catalogued thing.
type Feature struct {
	Name     string
	Kind     FeatureKind
	L, B, D  float64 // heliocentric galactic, degrees and kpc
	Pos      Vec     // galactocentric, computed
	Radius   float64 // kpc, its physical extent or the range of its influence
	Strength float64 // relative weight for the laws: 1 is ordinary
	Desc     string  // how a people living near it would tell it
	Fact     string  // what is actually known
	When     int64   // years before the present of a known event, 0 if none
	Event    string  // what the sky did then, if When is set
}

// Reach is how far away the feature still counts as near.
func (f *Feature) Reach() float64 {
	switch f.Kind {
	case Structure:
		return math.Max(1.0, 1.5*f.Radius)
	case Globular, Magnetar, BlackHole:
		return 1.5
	case Sky:
		return math.Inf(1)
	}
	return 1.0
}

// Features is the catalogue.
var Features = []*Feature{
	// the centre
	{Name: "Sagittarius A*", Kind: BlackHole, L: 359.944, B: -0.046, D: 8.2, Radius: 0.01, Strength: 8,
		Desc: "the great hole at the heart: four million suns of mass behind a veil of dust, and a sky of stars so close they blur into one light",
		Fact: "the supermassive black hole at the galactic centre, 4.3 million solar masses, orbited by the S-stars at a few hundred astronomical units"},
	{Name: "the Nuclear Cluster", Kind: Cluster, L: 359.944, B: -0.046, D: 8.2, Radius: 0.01, Strength: 20,
		Desc: "ten million stars packed into a few light years around the hole, old and young together, with no night anywhere",
		Fact: "the nuclear star cluster, about 25 million solar masses within 5 pc, with ongoing star formation despite the tides"},
	{Name: "the Central Molecular Zone", Kind: Structure, L: 0, B: 0, D: 8.2, Radius: 0.25, Strength: 3,
		Desc: "a thick belt of cloud and fire around the centre, where stars are born by the thousand and die by the hundred",
		Fact: "the CMZ: 60 million solar masses of dense molecular gas within 250 pc of the centre, a tenth of the galaxy's star formation"},
	{Name: "Sagittarius B2", Kind: Nebula, L: 0.667, B: -0.036, D: 8.2, Radius: 0.02, Strength: 30,
		Desc: "the largest cloud in the galaxy, laced with alcohol and sugar: the materials of life, gathered at the one place life cannot use them",
		Fact: "a giant molecular cloud of about 3 million solar masses near the centre, rich in complex organic molecules"},
	{Name: "the Arches Cluster", Kind: Cluster, L: 0.121, B: 0.017, D: 8.2, Radius: 0.005, Strength: 15,
		Desc: "a knot of stars each a hundred times the Sun, none older than a few million years, all of them about to die",
		Fact: "the densest known cluster in the galaxy, about 150 O-type stars within a parsec, 2 to 3 million years old"},
	{Name: "the Quintuplet Cluster", Kind: Cluster, L: 0.16, B: -0.06, D: 8.2, Radius: 0.005, Strength: 12,
		Desc: "five dust-shrouded giants and their court, among them the Pistol Star, which throws off a sun's mass of gas at a time",
		Fact: "a young massive cluster 30 pc from the centre, home to the Pistol Star, one of the most luminous stars known"},
	{Name: "the Fermi Bubbles", Kind: Structure, L: 0, B: 30, D: 8.2, Radius: 4, Strength: 1,
		Desc: "two lobes of thin fire standing above and below the centre, the breath of something that happened there a few million years ago",
		Fact: "twin lobes of gamma-ray emitting gas extending about 8 kpc above and below the centre, the outflow of an outburst of Sgr A* a few million years ago"},
	{Name: "the Bar", Kind: Structure, L: 0, B: 0, D: 8.2, Radius: 4.5, Strength: 1,
		Desc: "the old red heart of the galaxy, a lozenge of ancient stars turning end over end, sweeping the space within it clear of the clouds that make new ones",
		Fact: "the Galactic bar, half-length about 4.5 kpc, at about 30 degrees to the Sun-centre line; old, metal-rich stars"},
	{Name: "the 3-kpc Arm", Kind: Structure, L: 0, B: 0, D: 4.9, Radius: 0.6, Strength: 1,
		Desc: "a ring of gas streaming outward from the bar's ends, the inner shore of the disc",
		Fact: "the near 3-kpc arm, a gas structure at about 3.3 kpc from the centre expanding at about 50 km/s, with a far-side counterpart"},
	{Name: "the Molecular Ring", Kind: Structure, L: 25, B: 0, D: 4.5, Radius: 1.0, Strength: 1,
		Desc: "the great belt of cloud where the arms are born, thick with nurseries and with the wrecks of what they made",
		Fact: "a ring-like concentration of molecular gas at 4 to 5 kpc from the centre, where the major arms attach to the bar; most of the galaxy's star formation"},
	{Name: "the Scutum Red Supergiant Clusters", Kind: Cluster, L: 25.27, B: -0.16, D: 6.6, Radius: 0.1, Strength: 10,
		Desc: "a crowd of dying red giants at the root of the great arm, each one a supernova waiting its turn",
		Fact: "RSGC1 to 3 near l=25, clusters containing dozens of red supergiants 10 to 20 Myr old, at the near end of the bar"},
	{Name: "Westerlund 1", Kind: Cluster, L: 339.55, B: -0.4, D: 4.0, Radius: 0.01, Strength: 12,
		Desc: "a burning city of giants in Ara, with a magnetar at its heart that should have been a black hole",
		Fact: "the most massive young cluster in the galaxy, about 100,000 solar masses, 4 to 5 Myr old, hosting the magnetar CXOU J164710"},
	{Name: "NGC 3603", Kind: Cluster, L: 291.62, B: -0.52, D: 7.0, Radius: 0.01, Strength: 10,
		Desc: "a compact knot of giants in Carina, bright enough to be seen across the arm",
		Fact: "a dense young cluster in the Carina arm, one of the most massive HII regions in the galaxy"},
	{Name: "Westerlund 2", Kind: Cluster, L: 284.27, B: -0.33, D: 4.2, Radius: 0.01, Strength: 6,
		Desc: "a young cluster still wrapped in the cloud that made it",
		Fact: "a young massive cluster in the Carina arm, about 2 Myr old, within the Gum 29 HII region"},
	{Name: "Terzan 5", Kind: Globular, L: 3.84, B: 1.69, D: 5.9, Radius: 0.01, Strength: 1,
		Desc: "an ancient ball of stars lost in the dust of the bulge, older than the disc",
		Fact: "a globular cluster in the bulge with two stellar populations, possibly a remnant of the bulge's original building blocks"},

	// the arms toward the centre
	{Name: "the Carina Nebula", Kind: Nebula, L: 287.6, B: -0.63, D: 2.3, Radius: 0.05, Strength: 12,
		Desc: "a bright wound in the sky of Carina, cradle of giants and of the one giant that will end it",
		Fact: "one of the largest HII regions in the galaxy, home to Eta Carinae and several of the most massive stars known"},
	{Name: "Eta Carinae", Kind: Giant, L: 287.6, B: -0.63, D: 2.3, Radius: 0.03, Strength: 3,
		Desc: "a star of a hundred suns, wrapped in the cloud it threw off when it nearly died; when it dies, the arm will know",
		Fact: "a luminous blue variable of about 100 solar masses, which underwent a great eruption in the 1840s; expected to end as a supernova or hypernova"},
	{Name: "WR 104", Kind: Giant, L: 6.44, B: -0.48, D: 2.58, Radius: 0.03, Strength: 3,
		Desc: "a pinwheel of dust spun by two dying giants, its axis pointed, more or less, at the Sun",
		Fact: "a Wolf-Rayet binary with a spiral dust plume, its rotation axis within about 30 degrees of Earth's line of sight; a gamma-ray burst candidate"},
	{Name: "the Lagoon Nebula", Kind: Nebula, L: 5.96, B: -1.17, D: 1.25, Radius: 0.03, Strength: 4,
		Desc: "a glowing shore in Sagittarius, the first of the nurseries along the road inward",
		Fact: "M8, a giant HII region in the Sagittarius arm, 1.25 kpc away"},
	{Name: "the Omega Nebula", Kind: Nebula, L: 15.05, B: -0.67, D: 1.6, Radius: 0.03, Strength: 5,
		Desc: "a swan of light on the inner arm, hiding a cluster of hot young stars",
		Fact: "M17, one of the brightest and most massive star-forming regions in the galaxy"},
	{Name: "the Eagle Nebula", Kind: Nebula, L: 16.95, B: 0.79, D: 1.7, Radius: 0.03, Strength: 4,
		Desc: "pillars of cold dust standing in a wind of starlight, with stars condensing at their tips",
		Fact: "M16, the Pillars of Creation, a young cluster and its parent cloud in the Sagittarius arm"},
	{Name: "W49A", Kind: Nebula, L: 43.16, B: 0.01, D: 11, Radius: 0.03, Strength: 15,
		Desc: "a hidden furnace on the far side of the disc, making stars faster than anywhere outside the centre",
		Fact: "one of the most luminous star-forming regions in the galaxy, deeply embedded, on the far side of the Sagittarius arm"},
	{Name: "W51", Kind: Nebula, L: 49.4, B: -0.24, D: 5.4, Radius: 0.05, Strength: 10,
		Desc: "a great dark cloud in Aquila lit from within",
		Fact: "a giant molecular cloud complex with active massive star formation, at the tangent point of the Sagittarius arm"},
	{Name: "Kepler's Supernova", Kind: Remnant, L: 4.53, B: 6.82, D: 5, Radius: 0.02, Strength: 1, When: 420,
		Event: "A new star burns in Ophiuchus for a year, bright enough to see by day; then it fades.",
		Fact:  "SN 1604, the last supernova seen in the galaxy, a white dwarf explosion about 5 kpc away"},

	// the Sun's neighbourhood
	{Name: "the Local Bubble", Kind: Structure, L: 0, B: 0, D: 0, Radius: 0.1, Strength: 1,
		Desc: "a cavity of hot thin gas a few hundred light years across, blown out by the dead stars of the Scorpion; the Sun is passing through it",
		Fact: "a low-density region of hot ionised gas around the Sun, about 300 ly across, excavated by supernovae in the last 10 to 20 Myr"},
	{Name: "the Radcliffe Wave", Kind: Structure, L: 150, B: -5, D: 0.35, Radius: 1.2, Strength: 1,
		Desc: "a wave of cold gas nine thousand light years long undulating through the Local arm; every nearby nursery lies along it like beads on a string",
		Fact: "a coherent 2.7 kpc long structure of molecular gas discovered in 2020, connecting Orion, Taurus, Perseus, Cepheus and Cygnus, oscillating through the plane"},
	{Name: "the Scorpius-Centaurus Association", Kind: Cluster, L: 315.47, B: 16.2, D: 0.13, Radius: 0.07, Strength: 3,
		Desc: "the nearest nursery of great stars, spread across a quarter of the southern sky; its dead have already reached the Sun as iron on the sea floor",
		Fact: "the nearest OB association, about 400 ly away, 5 to 20 Myr old; supernovae in it deposited iron-60 on Earth 2 to 3 Myr ago"},
	{Name: "the Orion Nebula", Kind: Nebula, L: 209.01, B: -19.38, D: 0.41, Radius: 0.05, Strength: 8,
		Desc: "the nearest of the great nurseries, a lit cloud below the belt of the Hunter, still making stars",
		Fact: "M42, the nearest massive star-forming region, about 1,350 ly away, with the Trapezium cluster at its heart"},
	{Name: "Betelgeuse", Kind: Giant, L: 199.79, B: -8.96, D: 0.17, Radius: 0.03, Strength: 1,
		Desc: "a red giant on the Hunter's shoulder, swollen and unsteady, near the end",
		Fact: "a red supergiant about 550 ly away, expected to explode as a supernova within about 100,000 years; far enough to be harmless"},
	{Name: "Antares", Kind: Giant, L: 351.95, B: 15.07, D: 0.17, Radius: 0.03, Strength: 1,
		Desc: "the red heart of the Scorpion, a dying giant with a hot blue companion",
		Fact: "a red supergiant about 550 ly away in the Scorpius-Centaurus association, near the end of its life"},
	{Name: "Rigel", Kind: Giant, L: 209.24, B: -25.25, D: 0.26, Radius: 0.03, Strength: 1,
		Desc: "the Hunter's blue foot, a star of twenty suns burning through its life in a few million years",
		Fact: "a blue supergiant about 860 ly away, one of the most luminous stars in the local sky"},
	{Name: "Spica", Kind: Giant, L: 316.11, B: 50.84, D: 0.077, Radius: 0.02, Strength: 0.6,
		Desc: "a close pair of hot stars high above the plane, the nearest star that will one day tear itself apart",
		Fact: "a close binary of B-type stars about 250 ly away, the primary massive enough to end as a supernova"},
	{Name: "the Pleiades", Kind: Cluster, L: 166.64, B: -23.46, D: 0.136, Radius: 0.01, Strength: 1,
		Desc: "seven sisters and their hundreds of dimmer kin, young and blue, passing through a cloud that is not theirs",
		Fact: "an open cluster of about 1,000 stars, 100 Myr old, 444 ly away"},
	{Name: "the Hyades", Kind: Cluster, L: 180.08, B: -22.32, D: 0.047, Radius: 0.01, Strength: 0.2,
		Desc: "the nearest cluster, a loose crowd in the Bull's face, already growing old and drifting apart",
		Fact: "the nearest open cluster, about 150 ly away, 625 Myr old"},
	{Name: "the Taurus Cloud", Kind: Nebula, L: 173.16, B: -15.91, D: 0.14, Radius: 0.03, Strength: 1.5,
		Desc: "a dark cloud in the Bull making small stars quietly, no giants among them",
		Fact: "the Taurus molecular cloud, the nearest large star-forming region, about 450 ly away, forming only low-mass stars"},
	{Name: "the Ophiuchus Cloud", Kind: Nebula, L: 353.24, B: 16.58, D: 0.14, Radius: 0.02, Strength: 1.5,
		Desc: "a cloud of colours near Antares, where the newest stars in the local sky are lighting",
		Fact: "the Rho Ophiuchi cloud complex, about 460 ly away, one of the nearest star-forming regions"},
	{Name: "the Vela Pulsar", Kind: NeutronStar, L: 263.55, B: -2.79, D: 0.29, Radius: 0.05, Strength: 1, When: 11_000,
		Event: "A star dies in Vela, nine hundred light years off. For weeks it is brighter than the moon; then a spinning cinder remains, ticking.",
		Desc:  "a spinning cinder in a torn veil of light, the newest dead star near the Sun",
		Fact:  "a pulsar about 950 ly away, remnant of a supernova about 11,000 years ago; its remnant spans 8 degrees of sky"},
	{Name: "Geminga", Kind: NeutronStar, L: 195.13, B: 4.27, D: 0.25, Radius: 0.05, Strength: 0.8, When: 342_000,
		Event: "A star dies in the Twins, eight hundred light years off. The sky is bright for a year; the space around the Sun is swept clear.",
		Desc:  "a dead star in the Twins that shines only in light no eye can see",
		Fact:  "a nearby neutron star about 800 ly away, radio-quiet, seen only in gamma rays and X-rays; its supernova may have helped shape the Local Bubble"},
	{Name: "RX J1856.5-3754", Kind: NeutronStar, L: 358.6, B: -17.21, D: 0.123, Radius: 0.03, Strength: 0.6,
		Desc: "a bare dead star drifting through the Southern Crown, the nearest of its kind, cooling in silence",
		Fact: "one of the Magnificent Seven isolated neutron stars, about 400 ly away, the closest known"},
	{Name: "PSR J0437-4715", Kind: NeutronStar, L: 253.39, B: -41.96, D: 0.157, Radius: 0.03, Strength: 0.5,
		Desc: "a dead star spun up by its companion until it turns a hundred times a second, the steadiest clock in the sky",
		Fact: "the nearest millisecond pulsar, about 510 ly away, with a white dwarf companion"},
	{Name: "the Cygnus Loop", Kind: Remnant, L: 73.98, B: -8.56, D: 0.74, Radius: 0.05, Strength: 1, When: 20_000,
		Event: "A star dies in the Swan, two and a half thousand light years off. Its veil is still expanding.",
		Desc:  "a torn veil of light in the Swan, the wreck of a star that died when the ice was retreating",
		Fact:  "the Veil Nebula, a supernova remnant about 2,400 ly away, some 20,000 years old"},
	{Name: "the Monogem Ring", Kind: Remnant, L: 204.96, B: 6.43, D: 0.3, Radius: 0.1, Strength: 0.5, When: 86_000,
		Event: "A star dies between the Twins and the Unicorn, a thousand light years off.",
		Desc:  "a faint ring of old fire in the sky between the Twins and the Unicorn",
		Fact:  "a large old supernova remnant about 1,000 ly away, associated with the pulsar PSR B0656+14"},
	{Name: "Gaia BH1", Kind: BlackHole, L: 22.63, B: 18.05, D: 0.48, Radius: 0.02, Strength: 1,
		Desc: "a sun like the Sun in Ophiuchus, circling nothing: the nearest of the quiet holes",
		Fact: "the nearest known black hole, about 1,560 ly away, 9.6 solar masses, in a wide orbit with a Sun-like star; dormant"},
	{Name: "Gaia BH3", Kind: BlackHole, L: 51.67, B: -3.49, D: 0.59, Radius: 0.02, Strength: 1.5,
		Desc: "an old and metal-poor star in the Eagle bound to a hole of thirty suns, a stranger from the halo passing through the disc",
		Fact: "a 33 solar mass dormant black hole about 1,900 ly away, the most massive stellar black hole known in the galaxy, on a halo orbit"},
	{Name: "A0620-00", Kind: BlackHole, L: 209.96, B: -6.54, D: 1.06, Radius: 0.02, Strength: 1,
		Desc: "a hole in the Unicorn feeding slowly on a small star, and flaring once in a lifetime",
		Fact: "V616 Monocerotis, a black hole X-ray transient about 3,300 ly away, 6.6 solar masses; outbursts in 1917 and 1975"},
	{Name: "Gaia BH2", Kind: BlackHole, L: 310.4, B: 2.78, D: 1.16, Radius: 0.02, Strength: 1,
		Desc: "a red giant in the Centaur in a long slow orbit around nothing",
		Fact: "a 9 solar mass dormant black hole about 3,800 ly away with a red giant companion in a 1,277 day orbit"},
	{Name: "Cygnus X-1", Kind: BlackHole, L: 71.34, B: 3.07, D: 2.22, Radius: 0.05, Strength: 3,
		Desc: "a blue giant in the Swan being eaten by its unseen companion, screaming in X-rays as it goes",
		Fact: "the first black hole identified, 21 solar masses, about 7,200 ly away, accreting from a blue supergiant"},
	{Name: "V404 Cygni", Kind: BlackHole, L: 73.12, B: -2.09, D: 2.39, Radius: 0.03, Strength: 2,
		Desc: "a hole in the Swan that wakes once in a generation and outshines the arm",
		Fact: "a black hole X-ray transient about 7,800 ly away, 9 solar masses; outbursts in 1938, 1956, 1989 and 2015"},
	{Name: "Cygnus X", Kind: Nebula, L: 80.22, B: 0.8, D: 1.5, Radius: 0.1, Strength: 10,
		Desc: "a great lit cloud along the Swan, the nearest place where giants are born by the hundred",
		Fact: "one of the richest star-forming complexes in the galaxy, about 4,600 ly away, containing the Cygnus OB2 association"},
	{Name: "Deneb", Kind: Giant, L: 84.28, B: 2.0, D: 0.8, Radius: 0.03, Strength: 1.5,
		Desc: "the Swan's tail, a white giant so bright it is seen across the whole arm",
		Fact: "a blue-white supergiant about 2,600 ly away, among the most luminous stars known; its distance is uncertain"},
	{Name: "VY Canis Majoris", Kind: Giant, L: 239.35, B: -5.07, D: 1.2, Radius: 0.03, Strength: 1.5,
		Desc: "a red star so swollen that it would swallow the orbits of the outer planets, shedding itself in clouds",
		Fact: "a red hypergiant about 3,900 ly away, one of the largest stars known, losing mass in great eruptions"},
	{Name: "the Rosette Nebula", Kind: Nebula, L: 206.47, B: -1.64, D: 1.6, Radius: 0.04, Strength: 4,
		Desc: "a wreath of lit cloud in the Unicorn around a young cluster",
		Fact: "a large HII region about 5,200 ly away with the open cluster NGC 2244 at its centre"},
	{Name: "the Coalsack", Kind: Nebula, L: 302.77, B: 0.37, D: 0.18, Radius: 0.02, Strength: 0.5,
		Desc: "a hole in the southern sky by the Cross where no stars show: a dark cloud near enough to hide them",
		Fact: "a dark nebula about 600 ly away, one of the most prominent in the southern sky"},

	// the outer arms
	{Name: "the Crab", Kind: NeutronStar, L: 184.56, B: -5.78, D: 2.0, Radius: 0.03, Strength: 1, When: 970,
		Event: "A new star burns in the Bull, bright enough to see by day for weeks. Where it stood, a filament nebula grows around a dead star spinning thirty times a second.",
		Desc:  "a filament wreck in the Bull, a young dead star at its centre spinning thirty times a second",
		Fact:  "the remnant of SN 1054, about 6,500 ly away in the Perseus arm, powered by the Crab pulsar"},
	{Name: "Cassiopeia A", Kind: Remnant, L: 111.73, B: -2.13, D: 3.4, Radius: 0.03, Strength: 1.5, When: 340,
		Event: "A star dies in Cassiopeia, unseen behind the dust. Its wreck is the brightest thing in the sky at radio wavelengths.",
		Desc:  "the wreck of a star that died unseen behind the dust of the Perseus arm, loud in a light no eye sees",
		Fact:  "the youngest known supernova remnant in the galaxy, about 11,000 ly away, from a supernova around 1680 that went unrecorded"},
	{Name: "Tycho's Supernova", Kind: Remnant, L: 120.08, B: 1.42, D: 2.8, Radius: 0.02, Strength: 1, When: 450,
		Event: "A new star burns in Cassiopeia for over a year, brighter than Venus; then it fades.",
		Fact:  "SN 1572, a white dwarf explosion about 9,000 ly away in the Perseus arm"},
	{Name: "the Double Cluster", Kind: Cluster, L: 134.76, B: -3.69, D: 2.3, Radius: 0.02, Strength: 3,
		Desc: "two young clusters side by side in Perseus, marking the near edge of the outer arm",
		Fact: "h and Chi Persei, two young open clusters about 7,500 ly away in the Perseus arm"},

	// the inner disc, farther
	{Name: "SGR 1806-20", Kind: Magnetar, L: 10.0, B: -0.24, D: 8.7, Radius: 0.05, Strength: 3, When: 21,
		Event: "A dead star on the far side of the centre lets go of its field. For a fifth of a second it outshines the galaxy; the flash strips the day side of every world for a thousand light years.",
		Desc:  "a dead star wound so tight with magnetism that it cracks, and each crack is a flash seen across the galaxy",
		Fact:  "a magnetar about 28,000 ly away; its giant flare of December 2004 was the brightest event ever recorded from outside the solar system"},
	{Name: "SGR 1900+14", Kind: Magnetar, L: 43.02, B: 0.77, D: 12.5, Radius: 0.05, Strength: 2, When: 27,
		Event: "A magnetar in the Eagle flares. The ionosphere of every world for a thousand light years is knocked flat for a night.",
		Fact:  "a magnetar in Aquila whose 1998 giant flare ionised Earth's upper atmosphere"},
	{Name: "SS 433", Kind: BlackHole, L: 39.69, B: -2.24, D: 5.5, Radius: 0.05, Strength: 3,
		Desc: "a thing in the Eagle that fires two beams of matter at a quarter the speed of light, sweeping them around the sky like a lighthouse",
		Fact: "a microquasar about 18,000 ly away: a compact object with precessing relativistic jets, inside the remnant W50"},
	{Name: "GRS 1915+105", Kind: BlackHole, L: 45.37, B: -0.22, D: 8.6, Radius: 0.05, Strength: 3,
		Desc: "a heavy hole in the Eagle, spinning as fast as a hole can spin, with a giant for its meal",
		Fact: "a 12 solar mass black hole about 28,000 ly away with a near-maximally spinning horizon, a microquasar with superluminal jets"},
	{Name: "Cygnus X-3", Kind: BlackHole, L: 79.85, B: 0.7, D: 7.4, Radius: 0.05, Strength: 2,
		Desc: "a dead star and a naked giant in the Swan locked in an orbit of hours, throwing jets",
		Fact: "a microquasar about 24,000 ly away with a Wolf-Rayet companion in a 4.8 hour orbit"},

	// the halo
	{Name: "Omega Centauri", Kind: Globular, L: 309.1, B: 14.97, D: 5.2, Radius: 0.05, Strength: 3,
		Desc: "ten million stars in a ball a hundred light years across, older than the disc, with no night at its heart and nothing there that could bear to live",
		Fact: "the largest globular cluster of the galaxy, about 17,000 ly away, possibly the stripped core of a dwarf galaxy; may hold an intermediate-mass black hole"},
	{Name: "47 Tucanae", Kind: Globular, L: 305.9, B: -44.89, D: 4.5, Radius: 0.04, Strength: 2,
		Desc: "a second great ball of ancient stars beneath the southern pole, crowded with dead stars spun to a hum",
		Fact: "the second brightest globular cluster, about 14,700 ly away, rich in millisecond pulsars"},
	{Name: "Messier 4", Kind: Globular, L: 350.97, B: 15.97, D: 1.85, Radius: 0.03, Strength: 1,
		Desc: "the nearest of the ancient balls of stars, near Antares, holding a world that was old before the Sun was lit",
		Fact: "the nearest globular cluster, about 6,000 ly away; contains the pulsar planet PSR B1620-26 b, one of the oldest known planets"},
	{Name: "Messier 13", Kind: Globular, L: 59.01, B: 40.91, D: 7.1, Radius: 0.03, Strength: 1.5,
		Desc: "the great cluster in Hercules, to which a message was once sent from Earth that will arrive in twenty-five thousand years",
		Fact: "a globular cluster about 22,000 ly away, target of the 1974 Arecibo message"},
	{Name: "Messier 22", Kind: Globular, L: 9.89, B: -7.55, D: 3.2, Radius: 0.03, Strength: 1.5,
		Desc: "a bright ball of old stars toward the centre, with two small holes at its heart",
		Fact: "a globular cluster about 10,000 ly away toward the bulge, with two candidate stellar-mass black holes"},
	{Name: "NGC 6397", Kind: Globular, L: 338.16, B: -11.96, D: 2.3, Radius: 0.03, Strength: 1,
		Desc: "a collapsed ball of old stars in Ara, its core a knot of dead stars",
		Fact: "one of the two nearest globular clusters, about 7,800 ly away, with a core-collapsed centre of white dwarfs and neutron stars"},
	{Name: "the Sagittarius Stream", Kind: Sky, L: 5.6, B: -14.08, D: 26, Radius: 5, Strength: 1,
		Desc: "a river of stars wrapped around the galaxy, the remains of a small one being eaten; it has passed through the disc more than once",
		Fact: "the tidal stream of the Sagittarius dwarf galaxy, whose core is about 70,000 ly away on the far side of the centre; its passages through the disc may have triggered episodes of star formation"},

	// beyond
	{Name: "the Large Cloud", Kind: Sky, L: 280.47, B: -32.89, D: 50, Radius: 5, Strength: 1,
		Desc: "a small galaxy hanging in the southern sky like a torn piece of the Milky Way, with a nursery in it brighter than any in the disc",
		Fact: "the Large Magellanic Cloud, about 160,000 ly away, home of the Tarantula Nebula and SN 1987A"},
	{Name: "the Small Cloud", Kind: Sky, L: 302.8, B: -44.3, D: 62, Radius: 3, Strength: 1,
		Desc: "a fainter smudge beside the Large Cloud, joined to it by a bridge of gas",
		Fact: "the Small Magellanic Cloud, about 200,000 ly away"},
	{Name: "Andromeda", Kind: Sky, L: 121.17, B: -21.57, D: 765, Radius: 30, Strength: 1,
		Desc: "the other great galaxy, a faint oval in the northern sky, falling toward this one across an emptiness that will take four billion years to cross",
		Fact: "M31, about 2.5 million ly away, on a collision course with the Milky Way in about 4.5 billion years"},
}

var byName = map[string]*Feature{}

func init() {
	for _, f := range Features {
		f.Pos = FromSun(f.L, f.B, f.D)
		byName[f.Name] = f
	}
}

// FeatureByName looks a feature up, case-insensitively, by name or by a
// distinctive part of it.
func FeatureByName(name string) *Feature {
	n := lower(name)
	for _, f := range Features {
		if lower(f.Name) == n {
			return f
		}
	}
	for _, f := range Features {
		if contains(lower(f.Name), n) {
			return f
		}
	}
	return nil
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
