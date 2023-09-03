package lib

import (
	"fmt"
)

type NFA struct {
	States []*State
}

func (n *NFA) SetQty(min int, max int) error {
	if len(n.States) < 1 {
		return fmt.Errorf("quantity before repeatable item")
	}
	s := n.States[len(n.States)-1]
	s.Min = min
	s.Max = max
	return nil
}

func (n *NFA) AddRuneState(r rune) *State {
	n.States = append(n.States, &State{Match: []*Matcher{{Any: false, First: r, Last: r}}, Min: 1, Max: 1})
	return n.States[len(n.States)-1]
}

func (n *NFA) AddDotState() *State {
	n.States = append(n.States, &State{Match: []*Matcher{{Any: true}}, Min: 1, Max: 1})
	return n.States[len(n.States)-1]
}

type Matcher struct {
	Any   bool
	First rune
	Last  rune
}

type State struct {
	Match []*Matcher // items in an 'or' group (e.g. a|b|c)
	Min   int        // min matches
	Max   int        // max matches or -1 for many
}
