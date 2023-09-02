package lib

type NFA struct {
	States []*State
}

func (n *NFA) AddRuneState(r rune) {
	one := 1
	n.States = append(n.States, &State{})
	s := n.States[len(n.States)-1]
	s.min = &one
	s.max = &one
	m := s.LastOrNewMatcher()
	m.min = &r
	m.max = m.min
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

func (s *State) AddRuneMatch() {

}

func (s *State) LastOrNewMatcher() *Matcher {
	if len(s.Match) < 1 {
		s.Match = append(s.Match, &Matcher{})
	}
	return s.Match[len(s.Match)-1]
}
