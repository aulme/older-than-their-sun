package tech

// Descs says in one line what each node is: what a people has learned to
// do once they hold it. The mechanical effects live on the Node; this is
// the meaning.
var Descs = map[string]string{
	// era 0: the long dawn
	"tools":          "Shaped stone, bone and shell. The first lever over the world.",
	"fire":           "Kept flame: warmth, cooked food, the dark held back. Survival begins here.",
	"agriculture":    "Planting and herding. Surplus, settlement, and the first people who do not gather.",
	"writing":        "Memory outside the skull. Law, debt and story that outlive the teller.",
	"metallurgy":     "Smelting and forging. Blades, ploughs, and the arms race between them.",
	"mathematics":    "Number and proof. The habit of exactness that everything later rests on.",
	"astronomy":      "The sky as a clock and a map. The first hint that the lights are places.",
	"states":         "Kings, priests and tax. Force organised at the scale of a whole land.",
	"seafaring":      "Hulls and navigation. Coasts joined, strangers met, the world found to be round.",
	"cold_chemistry": "Making without burning: catalysis, pressure and patience. The fireless road to industry, for peoples who cannot burn.",
	// era 1: the turn to industry
	"printing":          "Cheap copies of anything written. Ideas spread faster than they can be stopped.",
	"scientific_method": "Asking the world questions and believing the answers. The engine of every later discovery.",
	"steam":             "Heat made to push. Work no longer limited by muscle.",
	"industrial":        "Factories, coal and rail. Everything made in quantity; nothing made by hand for long.",
	"chemistry":         "Matter understood as elements and bonds. Dyes, fertiliser, explosives, drugs.",
	"medicine":          "Germs, surgery and vaccines. Half of every generation stops dying young.",
	"firearms":          "Chemical energy thrown as metal. Killing at a distance becomes cheap.",
	"mass_politics":     "Newspapers, parties and crowds. Whole populations move as one, for good or ill.",
	"electricity":       "Power carried on wires. Light, motors and the first signals sent without a messenger.",
	// era 2: the atomic and the machine
	"mass_industry":   "Production at planetary scale. It strips a world if not watched, and no one watches at first.",
	"mechanised_war":  "Engines, armour and aircraft. Wars that consume nations rather than armies.",
	"physics":         "Relativity and the quantum. The rules underneath chemistry, and the first door to the atom.",
	"atomic":          "The split atom: unlimited power and a weapon that can end the species that holds it.",
	"rocketry":        "Machines in orbit. The world seen whole, and the first step off it.",
	"computers":       "Calculation by machine. Every later science runs on this.",
	"networks":        "Every mind on the world talking to every other. Society speeds up and thins out.",
	"genetics":        "Heredity read and edited. Crops, cures and the first deliberate changes to the body.",
	"ecology":         "The living world understood as one system, usually after nearly breaking it.",
	"orbital_weapons": "Weapons above the sky. Nowhere on the world is out of reach.",
	"fusion":          "A small star kept burning. Power without fuel worth fighting over.",
	"neuroscience":    "The mind mapped as a mechanism. The ground for machine minds and for changing one's own.",
	// era 3: the stars within reach
	"machine_minds":        "A mind that is not one of theirs. It speeds everything, or replaces its makers.",
	"closed_ecologies":     "Sealed, self-sustaining habitats. Living where nothing lives; arcologies at home.",
	"orbital_habitats":     "Cities in orbit and yards to build ships in. Population no longer bound to a surface.",
	"interplanetary":       "Fusion drives between the worlds of one star. The whole system becomes home.",
	"slow_interstellar":    "Generation ships and sleeper arks. Centuries to the nearest stars, but they arrive.",
	"self_replication":     "Machines that build machines. Industry without limit, and a swarm if the leash slips.",
	"life_extension":       "Ageing halted. Very long lives, and the question of whether anyone still wants anything.",
	"terraforming":         "A dead world remade in the image of home, over centuries.",
	"antimatter":           "Matter and its opposite, stored and burned. The densest energy there is.",
	"defence_grid":         "Guns that watch the sky. A world that cannot be taken cheaply.",
	"memetics":             "Ideas engineered to spread and to hold. Belief as a built thing.",
	"relativistic":         "Ships at a good fraction of light. Years to the stars instead of centuries.",
	"relativistic_weapons": "A fast enough rock ends any world. Everything that moves is a weapon now.",
	"uploading":            "Minds copied into machines. Death optional; identity negotiable.",
	"dyson":                "Swarms of collectors around the home star. Power on a scale that dims the star from outside.",
	"germline":             "The species redesigned at conception. Bodies fit for other worlds.",
	"quantum_computing":    "Computation on the quantum itself. The key to the exotic sciences.",
	"synthetic_biology":    "Life designed from scratch. Organisms as tools.",
	"deep_governance":      "Institutions that plan and hold across centuries and light-years.",
	"beamed_sails":         "Sails driven by lasers from home. Faster than arks, cheaper than starships.",
	"hibernation":          "Sleep across the crossing. Longer voyages become bearable.",
	// era 4: the deep tree
	"stellar_engineering": "Reaching into the star: shaping its output and, a little, moving it. It can flare.",
	"wormhole_physics":    "The proof that space can be folded. A proof, for now.",
	"exotic_matter":       "Matter with negative pressure, made and held. What a folded space needs to stay open.",
	"causal_physics":      "The structure of cause and effect itself, studied as a thing that can be leaned on.",
	"transcendence":       "The door out of the universe. Most who find it go through.",
	"star_lifting":        "Feeding and draining the star. A sun that will never be allowed to fail.",
	"deep_time":           "Reading the ages in the ash of dead stars: the galaxy has done this before. Needs patience and luck.",
	"vacuum_energy":       "Power drawn from the vacuum. Energy without a source anyone can take away.",
	"matter_compilers":    "Anything assembled atom by atom from feedstock. Scarcity ends, as a concept.",
	"world_engines":       "Whole planets moved, spun, warmed or built. Habitability by decree.",
	"substrate_minds":     "Minds run on whatever is to hand: rock, plasma, the quantum foam. Thought no longer needs a body.",
	"panspermia":          "Life seeded across the stars on purpose, tuned to grow into what is wanted.",
	"posthuman_law":       "Law for a people who are no longer one kind of thing. Holds societies of uploads, machines and flesh together.",
	"long_thought":        "A single argument carried on for millennia across a whole civilisation. Wisdom as infrastructure.",
	"near_light":          "Ships so close to light that a voyage is an afternoon inside and a lifetime outside.",
	"nova_bombs":          "Antimatter weapons that scour a world's face in one burst.",
	"stellar_weapons":     "The star itself turned into a gun. The last word in war that stays within physics.",
	// miracles: the powers apart from the tree
	"ansible":            "Speech across any distance with no delay. Every world in one room; every mind on one line.",
	"directed_evolution": "The species remade in a generation, a body for every world. No ships, no plague, no fear of the sky.",
	"ftl":                "A way between the stars faster than light, and so faster than cause. The stars are next door.",
	"unmaking":           "Matter ended at any distance that can be seen. Nothing can be defended against it.",
	"chorus":             "A thought that takes root in any mind that hears it. Whoever meets them joins them.",
	"foresight":          "Knowledge of what is coming. Half of every filter is seen and stepped around.",
}

func init() {
	for _, n := range Nodes {
		d, ok := Descs[n.Key]
		if !ok {
			panic("tech: no description for " + n.Key)
		}
		n.Desc = d
	}
}
