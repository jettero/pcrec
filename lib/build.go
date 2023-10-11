package lib

import (
	"fmt"
)

func (n *RE) SetQty(min int, max int) {
	n.LastStateish().SetQty(min, max)
}

func (n *RE) SetGreedy(f bool) {
	n.LastStateish().SetGreedy(f)
}

func (n *RE) SetCapture(f bool) {
	switch typed := n.LastStateish().(type) {
	case *Group:
		typed.SetCapture(f)
	}
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

func (g *Group) SetCapture(f bool) {
	g.Capture = f
}

func (n *RE) LastStateish() Stateish {
	ls := len(n.States) - 1
	if ls < 0 {
		return nil
	}
	return n.States[ls].LastStateish()
}

func (n *RE) AppendOrToGroupOrCreateGroup() bool {
	// you are here:
	// aba|
	// (a|b|
	if lg := n.LastOpenGroup(); lg != nil {
		lg.AppendOrClause()
		return false
	}

	// really, if we get an '|' pipe, it's either going to be in an open group
	// (above) or it's going to be in the top level expression …
	//
	// There's really no need to check anything further wrt that. Just replace
	// the top level states with the open group.

	g := &Group{States: [][]Stateish{n.States, {}},
		Min: 1, Max: 1, Capture: false, Greedy: true, Implicit: true}
	n.States = []Stateish{g}

	return true
}

func (s *State) LastStateish() Stateish {
	return s
}

func (g *Group) LastStateish() Stateish {
	if g.Closed {
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

func (s *State) LastOpenGroup() *Group {
	return nil
}

func (g *Group) LastOpenGroup() *Group {
	if i := len(g.States) - 1; i >= 0 {
		if j := len(g.States[i]) - 1; j >= 0 {
			if lg := g.States[i][j].LastOpenGroup(); lg != nil {
				return lg
			}
		}
	}
	if !g.Closed {
		return g
	}

	return nil
}

func (n *RE) LastOpenGroup() *Group {
	if i := len(n.States) - 1; i >= 0 {
		return n.States[i].LastOpenGroup()
	}
	return nil
}

func (n *RE) AppendGroup() {
	if lg := n.LastOpenGroup(); lg != nil {
		lg.AppendGroup(1, 1, true, true)
	} else {
		n.States = append(n.States, &Group{Min: 1, Max: 1, Greedy: true, Capture: true})
	}
}

func (n *RE) CloseGroup() error {
	if lg := n.LastOpenGroup(); lg != nil {
		lg.Closed = true
		return nil
	}
	return fmt.Errorf("unmatched closing parenthesis")
}

func (n *RE) CloseImplicitTopGroups() {
	if log := n.LastOpenGroup(); log != nil && log.Implicit {
		// There should only be the one anyway, right? ... right??
		log.Closed = true
	}
}

func (n *RE) AppendState(min int, max int, greedy bool) *State {
	if lg := n.LastOpenGroup(); lg != nil {
		return lg.AppendState(min, max, greedy)
	}
	ret := &State{Min: min, Max: max, Greedy: greedy}
	n.States = append(n.States, ret)
	return ret
}

func (n *RE) AppendRuneState(runes ...rune) *State {
	s := n.AppendState(1, 1, true)
	for _, r := range runes {
		s.AppendMatch(r, false)
	}
	return s
}

func (n *RE) AppendInvertedRuneState(runes ...rune) *State {
	s := n.AppendState(1, 1, true)
	s.And = true
	for _, r := range runes {
		s.AppendMatch(r, true)
	}
	return s
}

func (n *RE) AppendDotState() *State {
	s := n.AppendState(1, 1, true)
	s.AppendDotMatch()
	return s
}

func (g *Group) AppendGroup(min int, max int, greedy bool, capture bool) {
	if i := len(g.States) - 1; i >= 0 {
		g.States[i] = append(g.States[i], &Group{Min: min, Max: max, Greedy: greedy, Capture: capture})
	} else {
		g.States = append(g.States, []Stateish{&Group{Min: min, Max: max, Greedy: greedy, Capture: capture}})
	}
}

func (g *Group) AppendOrClause() {
	g.States = append(g.States, []Stateish{})
}

func (g *Group) AppendState(min int, max int, greedy bool) *State {
	ret := &State{Min: min, Max: max, Greedy: greedy}
	if i := len(g.States) - 1; i >= 0 {
		g.States[i] = append(g.States[i], ret)
	} else {
		g.States = append(g.States, []Stateish{ret})
	}
	return ret
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
