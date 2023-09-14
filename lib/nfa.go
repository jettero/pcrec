package lib

import (
	"fmt"
)

type NFA struct {
	States []Stateish
}

type Stateish interface {
	SetQty(min, max int)
	SetGreedy(f bool)
	LastStateish() Stateish
}

func (n *NFA) SetQty(min int, max int) {
	n.LastStateish().SetQty(min, max)
}

func (n *NFA) SetGreedy(f bool) {
	n.LastStateish().SetGreedy(f)
}

func (s *State) SetQty(min, max int) {
	s.Min = min
	s.Max = max
}

func (s *State) SetGreedy(f bool) {
	s.Greedy = f
}

func (g *Group) SetQty(min, max int) {
	g.Min = min
	g.Max = max
}

func (g *Group) SetGreedy(f bool) {
	g.Greedy = f
}

func (n *NFA) LastStateish() Stateish {
	ls := len(n.States) - 1
	if ls < 0 {
		return nil
	}
	return n.States[ls].LastStateish()
}

func (s *State) LastStateish() Stateish {
	return s
}

func (g *Group) LastStateish() Stateish {
	if g.closed {
		return g
	}
	lgs := len(g.States) - 1
	if lgs < 0 {
		return nil
	}
	lgss := len(g.States[lgs]) - 1
	if lgss < 0 {
		return nil
	}
	return g.States[lgs][lgss]
}

func (n *NFA) AppendState(min int, max int, greedy bool) *State {
	ls := n.LastStateish()
	switch typed := ls.(type) {
	case *Group:
		return typed.AppendState(min, max, greedy)
	}
	ret := &State{Min: min, Max: max, Greedy: greedy}
	n.States = append(n.States, ret)
	return ret
}

func (n *NFA) AddRuneState(runes ...rune) *State {
	s := n.AppendState(1, 1, true)
	for _, r := range runes {
		s.AppendMatch(r, false)
	}
	return s
}

func (n *NFA) AddInvertedRuneState(runes ...rune) *State {
	s := n.AppendState(1, 1, true)
	for _, r := range runes {
		s.AppendMatch(r, true)
	}
	return s
}

func (n *NFA) AddDotState() *State {
	s := n.AppendState(1, 1, true)
	s.AppendDotMatch()
	return s
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

type Group struct {
	States  [][]*State // []*State OR []*State OR ...
	Min     int        // min matches
	Max     int        // max matches or -1 for many
	Capture bool
	Greedy  bool

	// parser flags
	closed bool
}

func (g *Group) AppendState(min int, max int, greedy bool) *State {
	ret := &State{Min: min, Max: max, Greedy: greedy}
	lgs := len(g.States) - 1
	if lgs < 0 {
		g.States = append(g.States, []*State{ret})

	} else {
		g.States[lgs] = append(g.States[lgs], ret)
	}
	return ret
}

func (g *Group) GetOrCreateLastState() *State {
	lgs := len(g.States) - 1
	if lgs < 0 {
		return g.AppendState(1, 1, true)
	}
	lgss := len(g.States[lgs]) - 1
	if lgss < 0 {
		return g.AppendState(1, 1, true)
	}
	return g.States[lgs][lgss]
}

type State struct {
	Match  []*Matcher // items in an 'or' group (e.g. a|b|c)
	Min    int        // min matches
	Max    int        // max matches or -1 for many
	And    bool       // a&b&c, useful for: [^abc] => (^a&^b&^c) ≡ ^(a|b|c)
	Greedy bool
}

func (s *State) Matches(r rune) bool {
	for _, m := range s.Match {
		if m.Matches(r) {
			return true
		}
	}
	return false
}

func (s *State) AppendDotMatch() {
	s.Match = append(s.Match, &Matcher{Any: true})
}

func (s *State) AppendToLastMatch(r rune, inverse bool) error {
	i := len(s.Match) - 1
	if i < 0 {
		return fmt.Errorf("unable to append %d to last match as there are no matches in %+v", r, s)
	}
	if r < s.Match[i].First {
		return fmt.Errorf("unable to append %d to last match %+v, out of order", r, s.Match[i])
	}
	s.Match[i].Last = r
	return nil
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
