package lib

import (
	"fmt"
)

type RE struct {
	States []Stateish
	nfa    *NFA
}

type Stateish interface {
	SetQty(min, max int)
	SetGreedy(f bool)
	LastStateish() Stateish
	LastOpenGroup() *Group
	Describe(indent int) string
	short() string
	medium() string
}

type OldState struct {
	Match  []*Matcher // items in an 'or' group (e.g. a|b|c)
	Min    int        // min matches
	Max    int        // max matches or -1 for many
	And    bool       // a&b&c, useful for: [^abc] => (^a&^b&^c) ≡ ^(a|b|c)
	Greedy bool
}

type Group struct {
	States   [][]Stateish // []Stateish OR []Stateish OR …
	Min      int          // min matches
	Max      int          // max matches or -1 for many
	Capture  bool
	Greedy   bool
	Implicit bool // we can't close implicit groups with ')'

	// parser flags
	Closed bool
}

type Matcher struct {
	Inverse bool
	Any     bool
	First   rune
	Last    rune
}

func (r *RE) SetQty(min int, max int) {
	r.LastStateish().SetQty(min, max)
}

func (r *RE) SetGreedy(f bool) {
	r.LastStateish().SetGreedy(f)
}

func (r *RE) SetCapture(f bool) {
	switch typed := r.LastStateish().(type) {
	case *Group:
		typed.SetCapture(f)
	}
}

func (s *OldState) SetQty(min, max int) {
	s.Min = min
	s.Max = max
}

func (s *OldState) SetGreedy(f bool) {
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

func (r *RE) LastStateish() Stateish {
	ls := len(r.States) - 1
	if ls < 0 {
		return nil
	}
	return r.States[ls].LastStateish()
}

func (r *RE) AppendOrToGroupOrCreateGroup() bool {
	// you are here:
	// aba|
	// (a|b|
	if lg := r.LastOpenGroup(); lg != nil {
		lg.AppendOrClause()
		return false
	}

	// really, if we get an '|' pipe, it's either going to be in an open group
	// (above) or it's going to be in the top level expression …
	//
	// There's really no need to check anything further wrt that. Just replace
	// the top level states with the open group.

	g := &Group{States: [][]Stateish{r.States, {}},
		Min: 1, Max: 1, Capture: false, Greedy: true, Implicit: true}
	r.States = []Stateish{g}

	return true
}

func (s *OldState) LastStateish() Stateish {
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

func (s *OldState) LastOpenGroup() *Group {
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

func (r *RE) LastOpenGroup() *Group {
	if i := len(r.States) - 1; i >= 0 {
		return r.States[i].LastOpenGroup()
	}
	return nil
}

func (r *RE) AppendGroup() {
	if lg := r.LastOpenGroup(); lg != nil {
		lg.AppendGroup(1, 1, true, true)
	} else {
		r.States = append(r.States, &Group{Min: 1, Max: 1, Greedy: true, Capture: true})
	}
}

func (r *RE) CloseGroup() error {
	if lg := r.LastOpenGroup(); lg != nil {
		lg.Closed = true
		return nil
	}
	return fmt.Errorf("unmatched closing parenthesis")
}

func (r *RE) CloseImplicitTopGroups() {
	if log := r.LastOpenGroup(); log != nil && log.Implicit {
		// There should only be the one anyway, right? ... right??
		log.Closed = true
	}
}

func (r *RE) AppendState(min int, max int, greedy bool) *OldState {
	if lg := r.LastOpenGroup(); lg != nil {
		return lg.AppendState(min, max, greedy)
	}
	ret := &OldState{Min: min, Max: max, Greedy: greedy}
	r.States = append(r.States, ret)
	return ret
}

func (r *RE) AppendRuneState(runes ...rune) *OldState {
	s := r.AppendState(1, 1, true)
	for _, r := range runes {
		s.AppendMatch(r, false)
	}
	return s
}

func (r *RE) AppendInvertedRuneState(runes ...rune) *OldState {
	s := r.AppendState(1, 1, true)
	s.And = true
	for _, r := range runes {
		s.AppendMatch(r, true)
	}
	return s
}

func (r *RE) AppendDotState() *OldState {
	s := r.AppendState(1, 1, true)
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

func (g *Group) AppendState(min int, max int, greedy bool) *OldState {
	ret := &OldState{Min: min, Max: max, Greedy: greedy}
	if i := len(g.States) - 1; i >= 0 {
		g.States[i] = append(g.States[i], ret)
	} else {
		g.States = append(g.States, []Stateish{ret})
	}
	return ret
}

func (s *OldState) AppendDotMatch() {
	s.Match = append(s.Match, &Matcher{Any: true})
}

func (s *OldState) AppendToLastMatch(r rune, inverse bool) error {
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

func (s *OldState) AppendMatch(r rune, inverse bool) {
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
