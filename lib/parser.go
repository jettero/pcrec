package lib

import (
	"fmt"
)

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
	SUB_RET
	SUB_LHS
	SUB_RHS
)

type Parser struct {
	mode []int
	top  *NFA

	r rune // current rune
	m int  // current mode number
	n int  // current sub-mode number
	i int

	m_rreg1 []rune // tmp storage
	m_rreg2 []rune
}

func (p Parser) formatError(pat []rune) error {
	return fmt.Errorf("ERROR processing \"%s\": unexpected character '%c' at position %d", string(pat), pat[p.i], p.i+1)
}

func (p Parser) PushContext(c int) {
	p.n = SUB_INIT
	p.mode = append(p.mode, c)
	p.m = c
}

func (p Parser) PopContext() {
	p.mode = p.mode[:len(p.mode)-1]
	p.m = p.mode[len(p.mode)-1]
	p.n = SUB_RET
}

func (p Parser) Parse(pat []rune) (*NFA, error) {
	p.mode = []int{CTX_NONE}
	p.top = &NFA{}

	for p.i = 0; p.i < len(pat); p.i++ {
		p.r = pat[p.i]
		fmt.Printf("---=: p.pat[%d]: %c; p.mode: %+v\n", p.i, p.r, p.mode)

		switch {
		case p.m == CTX_NONE:
			switch {
			// matchers
			case p.r == '.':
				p.top.AddDotState()
			case p.r == '\n':
			case p.r == '[':
				p.PushContext(CTX_CCLASS)

				// new context
			case p.r == '(':
				p.PushContext(CTX_GROUP)
			case p.r == '\\':
				p.PushContext(CTX_SLASHED)

				// quantities
			case p.r == '{':
				p.PushContext(CTX_NQUANT)

				// without being in a capture context, these are wrong
			case p.r == '?':
			case p.r == '*':
			case p.r == '+':
			case p.r == ')':
				return p.top, p.formatError(pat)

			default:
				p.top.AddRuneState(p.r)
			}
		case p.m == CTX_SLASHED:
			switch {
			case p.r == 'u':
				p.PushContext(CTX_UNICODE)
			case p.r == 'x':
				p.PushContext(CTX_HEX)
			}
		case p.m == CTX_NQUANT:
			if p.n == SUB_INIT {
				p.m_rreg1 = []rune{}
				p.m_rreg2 = []rune{}
				p.n = SUB_LHS
			}
			switch {
			case '0' <= p.r && p.r <= '9':
				switch {
				case p.n == SUB_LHS:
					p.m_rreg1 = append(p.m_rreg1, p.r)
				case p.n == SUB_RHS:
					p.m_rreg2 = append(p.m_rreg2, p.r)
				}
			case p.r == ',':
				switch {
				case p.n == SUB_LHS:
					p.n = SUB_RHS
				default:
					return p.top, p.formatError(pat)
				}
			case p.r == '}':
				fmt.Printf("    nquant{%s, %s}\n", string(p.m_rreg1), string(p.m_rreg2))
				p.PopContext()
			}
		}
	}
	return p.top, nil
}

func Parse(pat []rune) (*NFA, error) {
	parser := Parser{}
	return parser.Parse(pat)
}
