package record

import "worldgen/internal/flow"

// Tally counts what a people did in war and peace, for the batch reports.
type Tally struct {
	Declared    int  `json:"declared"`
	Fought      int  `json:"fought"`
	Taken       int  `json:"taken"`
	Lost        int  `json:"lost"`
	Glassed     int  `json:"glassed"`
	Fleets      int  `json:"fleets"`
	Native      int  `json:"native"`
	Scouts      int  `json:"scouts"`
	Relief      int  `json:"relief"`
	Pacts       int  `json:"pacts"`
	Refused     int  `json:"refused"`
	Betrayals   int  `json:"betrayals"`
	Called      int  `json:"called"`
	Capitulated bool `json:"capitulated"`
	// exploration
	Surveys    int     `json:"surveys"`
	Charted    int     `json:"charted"`
	Blind      int     `json:"blind"`
	BlindLost  int     `json:"blind_lost"`
	FindSurvey int     `json:"find_survey"`
	FindSettle int     `json:"find_settle"`
	FindChance int     `json:"find_chance"`
	FindOwn    int     `json:"find_own"`
	MetTouch   int     `json:"met_touch"`
	MetHeard   int     `json:"met_heard"`
	MetSurvey  int     `json:"met_survey"`
	MetShip    int     `json:"met_ship"`
	Searched   float64 `json:"searched"` // kyr with the Sight turned outward; kyr holding it
	Sighted    float64 `json:"sighted"`
	// tellings
	Tales      int `json:"tales"`
	Witnessed  int `json:"witnessed"`
	Told       int `json:"told"`
	Read       int `json:"read"`
	Inherited  int `json:"inherited"`
	Forgot     int `json:"forgot"`
	Myths      int `json:"myths"`
	Revised    int `json:"revised"`
	Blamed     int `json:"blamed"`
	Testaments int `json:"testaments"`
	Restored   int `json:"restored"`
	// flows: ticks lived, ticks shedding, both by era, and both alone (one system)
	Ticks       int    `json:"ticks"`
	Lean        int    `json:"lean"`
	TicksAt     [5]int `json:"ticks_at"`
	LeanAt      [5]int `json:"lean_at"`
	AloneAt     [5]int `json:"alone_at"`
	LeanAloneAt [5]int `json:"lean_alone_at"`
	// trade: what was sent and what came, over the life; partners ever, and partners ever sent to
	Sent     flow.Income `json:"sent"`
	Got      flow.Income `json:"got"`
	Partners int         `json:"partners"`
	Fed      int         `json:"fed"`
	// ships: built, lost in battle, rotted laid up; ship-kyr of flow spent building; ticks starfaring, at the want, with ships laid up
	Built     int     `json:"built"`
	ShipsLost int     `json:"ships_lost"`
	Rotted    int     `json:"rotted"`
	Building  float64 `json:"building"`
	StarTicks int     `json:"star_ticks"`
	AtWant    int     `json:"at_want"`
	LaidTick  int     `json:"laid_tick"`
	// battles: fought as the attacker, won on the roll, worlds taken with nothing in the sky; garrison moves and musters ordered
	Battles   int `json:"battles"`
	Won       int `json:"won"`
	EmptySky  int `json:"empty_sky"`
	Garrisons int `json:"garrisons"`
	Musters   int `json:"musters"`
	// sightings: fleets seen by own eyes, interceptors sent, meetings fought, own fleets turned back or broken in the dark, pickets sent, ships crewed from fields
	Sightings  int `json:"sightings"`
	Intercepts int `json:"intercepts"`
	Meetings   int `json:"meetings"`
	Caught     int `json:"caught"`
	Pickets    int `json:"pickets"`
	Salvaged   int `json:"salvaged"`
	// wisdom: peoples fathomed, fathomings lost to a dark age, brokered attempts made for others, wars that ended unfathomed, remains sealed by looking before the leap; messages dropped unread; councils' verdicts and the sum of the acted-on odds' distance from the mean
	Fathomed      int     `json:"fathomed"`
	Unfathomed    int     `json:"unfathomed"`
	Brokered      int     `json:"brokered"`
	Misunderstood int     `json:"misunderstood"`
	Leaps         int     `json:"leaps"`
	Dropped       int     `json:"dropped"`
	Judged        int     `json:"judged"`
	ActedGap      float64 `json:"acted_gap"`
	// contracts: bought, sold, broken by this people, sold out of by it, tributes paid, sightings sold
	Hired         int `json:"hired"`
	Sold          int `json:"sold"`
	Broke         int `json:"broke"`
	BoughtOff     int `json:"bought_off"`
	Tributes      int `json:"tributes"`
	SoldSightings int `json:"sold_sightings"`
	// slights: taken in all; councils the offence alone held back
	Slights  float64 `json:"slights"`
	Deterred int     `json:"deterred"`
	// plagues: caught, cured, ticks contained, worlds lost, cults formed from it; senders closed out, messages dropped for it
	Sickened   int `json:"sickened"`
	Cured      int `json:"cured"`
	Contained  int `json:"contained"`
	WorldsSick int `json:"worlds_sick"`
	Cults      int `json:"cults"`
	Refusals   int `json:"refusals"`
	Shut       int `json:"shut"`
	Attempts   int `json:"attempts"` // plagues made and tried: attempts, ones that took, ones seen, ones that got out at discovery, ones that leaked while held
	Poisoned   int `json:"poisoned"`
	Detected   int `json:"detected"`
	Breakouts  int `json:"breakouts"`
	Leaks      int `json:"leaks"`
	Ridden     int `json:"ridden"` // peoples ridden by this parasite; risings against a rider
	Risen      int `json:"risen"`
	// ossification: facings of the filter, renaissances, times set, breaks, and the stiffness summed at each facing; see ossify.go
	OssFaced   int     `json:"oss_faced"`
	OssRenewed int     `json:"oss_renewed"`
	OssSet     int     `json:"oss_set"`
	OssBroke   int     `json:"oss_broke"`
	OssStiff   float64 `json:"oss_stiff"`
	// kinds: see eldritch.go and waking.go
	Appeared int `json:"appeared"`
	Deepened int `json:"deepened"`
	Tithed   int `json:"tithed"`
	Sleeps   int `json:"sleeps"`
	Wakings  int `json:"wakings"`
	Demands  int `json:"demands"`
	Unmade   int `json:"unmade"`
	Eaten    int `json:"eaten"` // ships grown by eating; worlds stripped and held empty
	Consumed int `json:"consumed"`
	Hunts    int `json:"hunts"` // hunts declared on a hole in the ledger; times the shape drifted
	Drifts   int `json:"drifts"`
}
