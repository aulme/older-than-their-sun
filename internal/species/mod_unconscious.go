package species

// unconscious: intelligence with no one home. It holds no grudge and
// takes none to heart, has no morale, never civil-wars, never ossifies,
// never faces the filters of belief and boredom; amoral always; as other
// as eldritch to anything conscious. It imitates the forms: it trades,
// swears and breaks pacts on appraisal alone, which is what makes it
// creepy.
var unconscious = &ModDef{Mod: Unconscious, Entry: Entry{
	Key:    "unconscious",
	Legacy: Draw{Base: 5.0 / 92, Skip: []string{"org"}}, // the old org-group weight, after the hive
	Draws:  Draw{Base: 1.0 / 20, Tilts: map[string]float64{"replicator": 3, "antimemetic": 3}, Skip: []string{"org"}},
	Profile: Profile{
		Mil: -0.5, Sur: 1, Soc: 1.5,
		Wis:        -1.5, // no intuition to see past, and nothing to read another mind with
		Dom:        M{"exotic": 0.6, "society": 0.4, "biology": 1.3},
		Memory:     1.5,
		FilterDiff: map[string]float64{"beacon": -3, "transcend": 2, "machines": -1, "silence": -2},
		Cannot:     Believes | Stiffens | CivilWars | HoldsGrudges | Wavers | Leads, // no idea can take a mind that is not there; nothing in it sets, splits, resents or despairs, and there is nobody to lead or to succeed
		NeverFaces: []string{"faith", "silence", "ossification", "upload"},          // nobody inside to copy out
		NoOne:      true,
		Amoral:     true,
		Morals:     [4]float64{2, 0.04, 3, 1},
	},
}}
