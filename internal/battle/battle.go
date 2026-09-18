// Package battle is the arithmetic of a fight: a level is how good a
// people's arms are, ships are how many it has, and the two meet only
// here. Quality turns a level into a multiplier on every ship and gun a
// people fights with; a fleet's strength is its ships times its owner's
// quality; the roll between two strengths carries multiplicative noise;
// losses are drawn in strength and paid in ships at each side's own
// quality, so better ships die less as well as win more. The package is
// pure: values in, values out, a *rand.Rand for the rolls, and nothing of
// worlds, stars or peoples.
package battle

import (
	"math"
	"math/rand/v2"
)

// Base is the multiplier per level: three levels are twice the ship, ten
// levels nine times.
const Base = 1.25

// Noise is the standard deviation of the log of the roll's noise: one draw
// on the ratio of the two strengths.
const Noise = 0.35

// LossShare is the most either side loses in one battle, as a share of
// the other's strength: the loss is uniform from nothing to this.
const LossShare = 0.5

// Quality is the multiplier a level gives: Base to the level.
func Quality(mil float64) float64 { return math.Pow(Base, mil) }

// Levels is the inverse: how many levels a strength ratio is worth.
// Levels(Quality(m)) is m.
func Levels(ratio float64) float64 { return math.Log(ratio) / math.Log(Base) }

// Strength is what a fleet fights with: its ships at its owner's quality.
func Strength(ships int, q float64) float64 { return float64(ships) * q }

// Roll decides a battle between two strengths: the attacker wins when its
// strength, under one draw of noise, exceeds the defender's. A strength of
// nothing never wins and never loses to nothing.
func Roll(r *rand.Rand, atk, def float64) bool {
	if atk <= 0 {
		return false
	}
	if def <= 0 {
		return true
	}
	return atk*math.Exp(r.NormFloat64()*Noise) > def
}

// Losses draws what each side loses, in strength: uniform from nothing to
// LossShare of the other's strength.
func Losses(r *rand.Rand, atk, def float64) (la, ld float64) {
	return r.Float64() * LossShare * def, r.Float64() * LossShare * atk
}

// ToShips turns a loss in strength into ships at a quality, rounded
// stochastically, and never more than the ships there are.
func ToShips(r *rand.Rand, loss, q float64, ships int) int {
	if ships <= 0 || loss <= 0 {
		return 0
	}
	x := loss / q
	n := int(x)
	if r.Float64() < x-float64(n) {
		n++
	}
	return min(n, ships)
}

// Fight is one battle between an attacker and what holds a world: the
// roll and the losses in one call, with the defender's losses falling on
// its guns first and the rest on its ships. Guns and ships are the
// defender's at one quality; a defence of several qualities is summed by
// the caller and split by it. Returns whether the attacker won the roll
// and what each side lost in strength.
func Fight(r *rand.Rand, atk, def float64) (won bool, la, ld float64) {
	won = Roll(r, atk, def)
	la, ld = Losses(r, atk, def)
	return
}

// Hold runs a siege to the end on the rule as the world runs it, one
// battle a thousand years: an attacking fleet against guns alone, the
// guns losing first and the attacker fighting until it is wiped or no gun
// stands. With repair a gun is remade each thousand years the world is
// still held. Returns whether the guns held. It is what the proposal's
// table was drawn from.
func Hold(r *rand.Rand, ships int, q float64, guns int, gq float64, repair bool) bool {
	full := guns
	for ships > 0 && guns > 0 {
		atk, def := Strength(ships, q), Strength(guns, gq)
		_, la, ld := Fight(r, atk, def)
		ships -= ToShips(r, la, q, ships)
		guns -= ToShips(r, ld, gq, guns)
		if repair && guns > 0 && guns < full {
			guns++
		}
	}
	return guns > 0
}
