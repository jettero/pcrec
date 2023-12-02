package lib

type State struct {
	Xfer   map[*Transition]*State // *State is nil for Accept?
	Accept bool                   // or sould we have an explicit accepting state?
	Qty    *Qty                   // XXX do states have quantities? are they shared with Transitions?
}

type Qty struct {
	Min int
	Max int // -1 indicates ∞

	// Is this quantity capturing?
	Capture int // 0, or the number of the capture group
}

type Transition struct {
	// What we consume in the transition
	First rune // stated as a range
	Last  rune // of contiguous runes

	// although, perhaps we say
	Inverse bool // anything except the above

	// and we can only use this transition if the
	Qty *Qty // XXX does nil mean anything?
	// XXX This is a ptr so we can share the Qty between things
	// XXX Maybe multiple Transitions?
	// XXX Or what do States store?
	// XXX should this even be here? should it be in State?
	// XXX prolly, if *Qty is nil, we mean exactly 1? or do we always populate this?
}
