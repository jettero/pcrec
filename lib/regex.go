package lib

import (
	"fmt"
)

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

const (
	CTX_NONE int = iota
	CTX_SLASHED
	CTX_CCLASS
	CTX_UNICODE
	CTX_HEX
	CTX_GROUP
	CTX_NQUANT
)

const (
	SUB_INIT int = iota
	SUB_LHS
	SUB_RHS
)

func showError(pat []rune, pos int) error {
	return fmt.Errorf("ERROR processing \"%s\": unexpected character '%c' at position %d", string(pat), pat[pos], pos+1)
}

func Parse(pat []rune) (*NFA, error) {
	mode := []int{CTX_NONE}
	nfa := &NFA{}
	ret := nfa

	// places to store things during the parse
	var m_rreg1 []rune
	var m_rreg2 []rune

	var r rune // current rune
	var m int  // current mode number
	var n int  // current sub-mode number
	for i := 0; i < len(pat); i++ {
		r = pat[i]
		m = mode[len(mode)-1]

		fmt.Printf("---=: pat[%d]: %c; mode: %+v\n", i, r, mode)

		switch {
		case m == CTX_NONE:
			switch {
			// matchers
			case r == '.':
			case r == '\n':
			case r == '[':
				n = SUB_INIT
				mode = append(mode, CTX_CCLASS)

				// new context
			case r == '(':
				n = SUB_INIT
				mode = append(mode, CTX_GROUP)
			case r == '\\':
				n = SUB_INIT
				mode = append(mode, CTX_SLASHED)

				// quantities
			case r == '?':
			case r == '*':
			case r == '+':
			case r == '{':
				n = SUB_INIT
				mode = append(mode, CTX_NQUANT)

				// we shouldn't encounter the following here
			case r == ')':
				fallthrough
			case r == ']':
				fallthrough
			case r == '|':
				fallthrough
			case r == '}':
				return ret, showError(pat, i)

			default:
				nfa.AddRuneState(r)
			}
		case m == CTX_SLASHED:
			switch {
			case r == 'u':
				mode = append(mode, CTX_UNICODE)
			case r == 'x':
				mode = append(mode, CTX_HEX)
			}
		case m == CTX_NQUANT:
			if n == SUB_INIT {
				m_rreg1 = []rune{}
				m_rreg2 = []rune{}
				n = SUB_LHS
			}
			switch {
			case '0' <= r && r <= '9':
				switch {
				case n == SUB_LHS:
					m_rreg1 = append(m_rreg1, r)
				case n == SUB_RHS:
					m_rreg2 = append(m_rreg2, r)
				}
			case r == ',':
				switch {
				case n == SUB_LHS:
					n = SUB_RHS
				default:
					return ret, showError(pat, i)
				}
			case r == '}':
				fmt.Printf("    nquant{%s, %s}\n", string(m_rreg1), string(m_rreg2))
				mode = mode[:len(mode)-1]
			}
		}
	}
	return ret, nil
}
