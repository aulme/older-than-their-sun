package flow

// Share is trade's arithmetic: one people's surplus of a kind, offered to
// its partners against what each wants. The pool is the surplus times the
// largest cap, or the sum of the wants if that is less; it is split by
// want share, and each partner's take is clipped to the surplus times its
// own cap. Nothing sent exceeds the surplus, nobody gets more than it
// wants, and what was surplus costs the sender nothing, so the sum of
// working uses can only rise. Caps are per partner: the sender's drive,
// zero when out of reach.
func Share(surplus float64, wants, caps []float64) []float64 {
	out := make([]float64, len(wants))
	if surplus <= 0 {
		return out
	}
	total, capMax := 0.0, 0.0
	for i := range wants {
		if caps[i] <= 0 || wants[i] <= 0 {
			continue
		}
		total += wants[i]
		capMax = max(capMax, caps[i])
	}
	if total <= 0 {
		return out
	}
	pool := min(surplus*capMax, total)
	for i := range wants {
		if caps[i] <= 0 || wants[i] <= 0 {
			continue
		}
		out[i] = min(pool*wants[i]/total, surplus*caps[i])
	}
	return out
}
