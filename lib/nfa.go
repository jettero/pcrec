package lib

import (
	"fmt"
)

type NFA struct {
	States []*State
}

func (n *NFA) SetQty(min int, max int) error {
	if len(n.States) < 1 {
		return fmt.Errorf("quantifier before repeatable item")
	}
	s := n.States[len(n.States)-1]
	if s.QtySet {
		return fmt.Errorf("quantifier does not follow a repeatable item")
	}
	s.Min = min
	s.Max = max
	s.QtySet = true
	return nil
}

func (n *NFA) AddRuneState(r rune) *State {
	n.States = append(n.States, &State{Match: []*Matcher{{Any: false, First: r, Last: r}}, Min: 1, Max: 1, Greedy: true, Capture: false})
	return n.States[len(n.States)-1]
}

func (n *NFA) AddDotState() *State {
	n.States = append(n.States, &State{Match: []*Matcher{{Any: true}}, Min: 1, Max: 1, Greedy: true, Capture: false})
	return n.States[len(n.States)-1]
}

type Matcher struct {
	Any   bool
	First rune
	Last  rune
}

type State struct {
	Match   []*Matcher // items in an 'or' group (e.g. a|b|c)
	Min     int        // min matches
	Max     int        // max matches or -1 for many
	Greedy  bool       // usually we want to capture/match as many as possible
	Capture bool       // we can match in groups without capturing
	QtySet  bool       // whether we've already set a qty on this state
}
