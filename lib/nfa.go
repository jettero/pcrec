package lib

type NFA struct {
	States []*State
}

func (n *NFA) SetQty(min int, max int) {
	s := n.States[len(n.States)-1]
	s.Min = min
	s.Max = max
}

func (n *NFA) AddRuneState(runes ...rune) *State {
	n.States = append(n.States, &State{Min: 1, Max: 1, Greedy: true, Capture: false})
	s := n.States[len(n.States)-1]
	for _, r := range runes {
		s.AppendMatch(r, false)
	}
	return n.States[len(n.States)-1]
}

func (n *NFA) AddInvertedRuneState(runes ...rune) *State {
	n.States = append(n.States, &State{Min: 1, Max: 1, Greedy: true, Capture: false})
	s := n.States[len(n.States)-1]
	for _, r := range runes {
		s.AppendMatch(r, true)
	}
	return n.States[len(n.States)-1]
}

func (n *NFA) AddDotState() *State {
	n.States = append(n.States, &State{Match: []*Matcher{{Any: true}}, Min: 1, Max: 1, Greedy: true, Capture: false})
	return n.States[len(n.States)-1]
}

type Matcher struct {
	Inverse bool
	Any     bool
	First   rune
	Last    rune
}

func (m *Matcher) Matches(r rune) bool {
	if m.Any {
		return true
	}
	return m.Inverse != (m.First <= r && r <= m.Last) // inverse ^ between
}

type State struct {
	Match   []*Matcher // items in an 'or' group (e.g. a|b|c)
	Min     int        // min matches
	Max     int        // max matches or -1 for many
	Greedy  bool       // usually we want to capture/match as many as possible
	Capture bool       // we can match in groups without capturing
}

func (s *State) Matches(r rune) bool {
	for _, m := range s.Match {
		if m.Matches(r) {
			return true
		}
	}
	return false
}

func (s *State) AppendMatch(r rune, inverse bool) {
	for _, m := range s.Match {
		if m.Inverse == inverse && m.First <= r && r <= m.Last {
			return
		}
		if m.Inverse == inverse && r == m.First-1 {
			m.First--
			return
		}
		if m.Inverse == inverse && r == m.Last+1 {
			m.Last++
			return
		}
	}
	s.Match = append(s.Match, &Matcher{First: r, Last: r, Inverse: inverse})
}
