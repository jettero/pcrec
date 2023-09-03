package lib

import (
	"fmt"
	"strconv"
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
	mode []Context
	top  *NFA
	pat  []rune

	r rune // current rune
	m int  // current mode number
	n int  // current sub-mode number
	i int  // pos in pattern

	m_rreg1 []rune // tmp storage
	m_rreg2 []rune
}

type Context struct {
	m int // this mode
	i int // p.i at time of push
}

func (p *Parser) formatError(msg string) error {
	firstPart := "ERROR"
	if len(msg) > 0 {
		firstPart = fmt.Sprintf("ERROR: %s, while", msg)
	}
	return fmt.Errorf("%s processing \"%s\": unexpected character '%c' at position %d", firstPart, string(p.pat), p.pat[p.i], p.i+1)
}

func (p *Parser) PushContext(c int) {
	p.mode = append(p.mode, Context{m: c, i: p.i})
	p.m = c
	p.n = SUB_INIT
	fmt.Printf("  PushContext() => %d.%d\n", p.m, p.n)
}

func (p *Parser) PopContext(restore_i bool) {
	if len(p.mode) > 1 {
		o := p.mode[len(p.mode)-1]
		p.mode = p.mode[:len(p.mode)-1]
		p.m = p.mode[len(p.mode)-1].m
		p.n = SUB_RET
		if restore_i {
			p.i = o.i - 1 // gets incremented immediately after pop, so go one early
			p.r = p.pat[p.i]
		}
		fmt.Printf("  PopContext(%v) => p.i: %d; p.r: %c; p.mn: %d.%d\n", restore_i, p.i, p.r, p.m, p.n)
	}
}

func (p Parser) Parse(pat []rune) (*NFA, error) {
	p.mode = []Context{{m: CTX_NONE, i: 0}}
	p.top = &NFA{}
	p.pat = pat

	for p.i = 0; p.i < len(p.pat); p.i++ {
		p.r = p.pat[p.i]

		switch {
		case p.m == CTX_NONE:
			switch {
			// matchers
			case p.r == '.':
				p.top.AddDotState()
			case p.r == '\n':
				p.top.AddRuneState(p.r) // this may be $ sometimes with /m
			case p.r == '[':
				p.PushContext(CTX_CCLASS)

				// new context
			case p.r == '(':
				p.PushContext(CTX_GROUP)
			case p.r == '\\':
				p.PushContext(CTX_SLASHED)

				// quantities
			case p.r == '{':
				if p.n == SUB_RET {
					fmt.Printf(" AddRuneState(%c) NQUANT-RETURN\n", p.r)
					p.top.AddRuneState(p.r)
				} else {
					p.PushContext(CTX_NQUANT)
				}

				// without being in a capture context, these are wrong
			case p.r == '?':
				if err := p.top.SetQty(0, 1); err != nil {
					return p.top, p.formatError(err.Error())
				}
			case p.r == '*':
				if err := p.top.SetQty(0, -1); err != nil {
					return p.top, p.formatError(err.Error())
				}
			case p.r == '+':
				if err := p.top.SetQty(1, -1); err != nil {
					return p.top, p.formatError(err.Error())
				}
			case p.r == ')':
				return p.top, p.formatError("")

			default:
				fmt.Printf(" AddRuneState(%c) default\n", p.r)
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
					return p.top, p.formatError("")
				}
			case p.r == '}':
				if len(p.m_rreg1) > 0 || len(p.m_rreg2) > 0 {
					var a int = 0
					var b int = -1
					if num, err := strconv.ParseInt(string(p.m_rreg1), 10, 0); err == nil {
						a = int(num)
					}
					if num, err := strconv.ParseInt(string(p.m_rreg2), 10, 0); err == nil {
						b = int(num)
					}
					if err := p.top.SetQty(a, b); err != nil {
						return p.top, p.formatError(err.Error())
					}
					p.PopContext(false)
				} else {
					p.PopContext(true)
				}
			}
		}
		fmt.Printf("---=: p.pat[%d]: %c; p.mode: %+v; p.mn: %d.%d\n", p.i, p.r, p.mode, p.m, p.n)
	}
	return p.top, nil
}

func Parse(pat []rune) (*NFA, error) {
	parser := Parser{}
	return parser.Parse(pat)
}
