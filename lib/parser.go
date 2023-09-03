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
	SUB_INIT int = iota // we just pushed a context or just returned from one
	SUB_RET             // we just popped context and restored position
	SUB_LHS             // context specific
	SUB_RHS             // context specific
	SUB_QTY             // we just issued SetQty
)

type Parser struct {
	trace bool
	mode  []Context
	top   *NFA
	pat   []rune

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
	p.Printf("  PushContext() => %d.%d\n", p.m, p.n)
}

func (p *Parser) PopContext(restore_i bool) {
	if len(p.mode) > 1 {
		o := p.mode[len(p.mode)-1]
		p.mode = p.mode[:len(p.mode)-1]
		p.m = p.mode[len(p.mode)-1].m
		if restore_i {
			p.i = o.i - 1 // gets incremented immediately after pop, so go one early
			p.r = p.pat[p.i]
			p.n = SUB_RET
		} else {
			p.n = SUB_INIT
		}
		p.Printf("  PopContext(%v) => p.i: %d; p.r: %c; p.mn: %d.%d\n", restore_i, p.i, p.r, p.m, p.n)
	}
}

func (p *Parser) Printf(format string, args ...interface{}) {
	if p.trace {
		fmt.Printf(format, args...)
	}
}

func (p *Parser) Parse(pat []rune) (*NFA, error) {
	p.trace = TruthyEnv("PCREC_TRACE") || TruthyEnv("RE_PARSE_TRACE")
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
					p.Printf(" AddRuneState(%c) NQUANT-RETURN\n", p.r)
					p.top.AddRuneState(p.r)
					p.n = SUB_INIT
				} else {
					p.PushContext(CTX_NQUANT)
				}

			case p.r == '?':
				if p.n == SUB_QTY {
					s := p.top.States[len(p.top.States)-1]
					s.Greedy = false
				} else if err := p.top.SetQty(0, 1); err != nil {
					return p.top, p.formatError(err.Error())
				}
				p.n = SUB_QTY
			case p.r == '*':
				if err := p.top.SetQty(0, -1); err != nil {
					return p.top, p.formatError(err.Error())
				}
				p.n = SUB_QTY
			case p.r == '+':
				if err := p.top.SetQty(1, -1); err != nil {
					return p.top, p.formatError(err.Error())
				}
				p.n = SUB_QTY
			case p.r == ')':
				return p.top, p.formatError("")

			default:
				p.Printf(" AddRuneState(%c) default\n", p.r)
				p.top.AddRuneState(p.r)
				p.n = SUB_INIT
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
					p.n = SUB_QTY
				} else {
					p.PopContext(true)
				}
			}
		}
		p.Printf("---=: p.pat[%d]: %c; p.mode: %+v; p.mn: %d.%d\n", p.i, p.r, p.mode, p.m, p.n)
	}
	return p.top, nil
}

func Parse(pat []rune) (*NFA, error) {
	parser := Parser{}
	return parser.Parse(pat)
}
