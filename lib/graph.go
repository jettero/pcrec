package lib

type State struct {
	Xfer  []*Transition // *State is nil for Accept?
	Qty   *Qty          // XXX do states have quantities? are they shared with Transitions?
	Graph *Graph        // what graph am I in?
}

type Qty struct {
	Min int
	Max int // -1 indicates ∞

	// Is this quantity capturing?
	Capture []int // the numbers of any capture groups
}

type Transition struct {
	// What we consume in the transition
	First rune // stated as a range
	Last  rune // of contiguous runes

	// although, perhaps we say
	Inverse bool // anything except the above

	// should Transitions have a Qty? is it he same as the parent state?
	Qty *Qty // XXX does nil mean anything?

	From *State // keep pointers for the source state
	To   *State // and the target state
}

// XXX maybe to figure out if the above would work, we just try to manually
// populate a couple and see what's missing

type Graph struct {
	States []*State
	Start  *State // this is really just for reference. we can start a search anywhere we'd like
	Accept *State

	CaptureGroups int // counter to keep track of what capture group we're on
}

func MakeGraph() (g *Graph) {
	g = &Graph{}
	g.Start = g.MakeState()
	g.Accept = g.MakeState()
	return
}

func (g *Graph) MakeState() (s *State) {
	s = &State{Qty: &Qty{Min: 1, Max: 1}}
	g.States = append(g.States, s)
	return
}

func (s *State) AddTransition(first rune, last rune, inverse bool, target *State) (t *Transition) {
	t = &Transition{Qty: s.Qty, From: s, To: target}
	s.Xfer = append(s.Xfer, t)
	return
}
