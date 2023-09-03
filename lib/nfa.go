package lib

type NFA struct {
	States []*State
}

var one int = 1

func (n *NFA) AddRuneState(r rune) *State {
	n.States = append(n.States, &State{Match: []*Matcher{{min: &r, max: &r}}, min: &one, max: &one})
	return n.States[len(n.States)-1]
}

func (n *NFA) AddDotState() *State {
	n.States = append(n.States, &State{Match: []*Matcher{{}}, min: &one, max: &one})
	return n.States[len(n.States)-1]
}

func (n *NFA) LastOrNewState() *State {
	if len(n.States) < 1 {
		n.States = append(n.States, &State{})
	}
	return n.States[len(n.States)-1]
}

type Matcher struct {
	min *rune
	max *rune
}

type State struct {
	Match []*Matcher
	min   *int
	max   *int
}

func (s *State) LastOrNewMatcher() *Matcher {
	if len(s.Match) < 1 {
		s.Match = append(s.Match, &Matcher{})
	}
	return s.Match[len(s.Match)-1]
}
