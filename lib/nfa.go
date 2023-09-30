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
	LastOpenGroup() *Group
	Describe(indent int) string
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

func (n *NFA) AppendOrToGroupOrCreateGroup() bool {
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

	g := &Group{Implicit: true, Min: 1, Max: 1, States: [][]Stateish{n.States}}
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

func (n *NFA) LastOpenGroup() *Group {
	if i := len(n.States) - 1; i >= 0 {
		return n.States[i].LastOpenGroup()
	}
	return nil
}

func (n *NFA) AppendGroup() {
	if lg := n.LastOpenGroup(); lg != nil {
		lg.AppendGroup(1, 1, true)
	} else {
		n.States = append(n.States, &Group{Min: 1, Max: 1, Greedy: true})
	}
}

func (n *NFA) CloseGroup() error {
	if lg := n.LastOpenGroup(); lg != nil {
		lg.Closed = true
		return nil
	}
	return fmt.Errorf("unmatched closing parenthesis")
}

func (n *NFA) CloseImplicitTopGroups() {
	if log := n.LastOpenGroup(); log != nil && log.Implicit {
		// There should only be the one anyway, right? ... right??
		log.Closed = true
	}
}

func (n *NFA) AppendState(min int, max int, greedy bool) *State {
	if lg := n.LastOpenGroup(); lg != nil {
		return lg.AppendState(min, max, greedy)
	}
	ret := &State{Min: min, Max: max, Greedy: greedy}
	n.States = append(n.States, ret)
	return ret
}

func (n *NFA) AppendRuneState(runes ...rune) *State {
	s := n.AppendState(1, 1, true)
	for _, r := range runes {
		s.AppendMatch(r, false)
	}
	return s
}

func (n *NFA) AppendInvertedRuneState(runes ...rune) *State {
	s := n.AppendState(1, 1, true)
	s.And = true
	for _, r := range runes {
		s.AppendMatch(r, true)
	}
	return s
}

func (n *NFA) AppendDotState() *State {
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

type Group struct {
	States   [][]Stateish // []Stateish OR []Stateish OR …
	Min      int          // min matches
	Max      int          // max matches or -1 for many
	Capture  bool
	Greedy   bool
	Implicit bool // we can't close implicit groups with '('

	// parser flags
	Closed bool
}

func (g *Group) AppendGroup(min int, max int, greedy bool) {
	if i := len(g.States) - 1; i >= 0 {
		g.States[i] = append(g.States[i], &Group{Min: min, Max: max, Greedy: greedy})
	} else {
		g.States = append(g.States, []Stateish{&Group{Min: min, Max: max, Greedy: greedy}})
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

type State struct {
	Match  []*Matcher // items in an 'or' group (e.g. a|b|c)
	Min    int        // min matches
	Max    int        // max matches or -1 for many
	And    bool       // a&b&c, useful for: [^abc] => (^a&^b&^c) ≡ ^(a|b|c)
	Greedy bool
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
