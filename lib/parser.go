package lib

import (
	"fmt"
	"strconv"
)

const (
	CTX_NONE int = iota
	CTX_SLASHED
	CTX_CCLASS
	CTX_GROUP
	CTX_NQUANT
)

const (
	SUB_INIT int = iota // we just pushed a context or just returned from one
	SUB_RET             // we just popped context and restored position
	SUB_LHS             // context specific
	SUB_RHS             // context specific
	SUB_QTY             // we just issued SetQty()
	SUB_REP             // we just queued a repeatable item (relating to qtys)
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
	n int // p.n at time of push
	i int // p.i at time of push
}

func (p *Parser) formatError(msg string) (*NFA, error) {
	firstPart := "ERROR"
	if len(msg) > 0 {
		firstPart = fmt.Sprintf("ERROR: %s, while", msg)
	}
	return p.top, fmt.Errorf("%s processing \"%s\": unexpected character '%c' at position %d", firstPart, string(p.pat), p.pat[p.i], p.i+1)
}

func (p *Parser) PushContext(c int) {
	p.mode = append(p.mode, Context{m: c, i: p.i, n: p.n})
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

func (p *Parser) QTop() *NFA {
	p.n = SUB_QTY
	return p.top
}

func (p *Parser) RTop() *NFA {
	p.n = SUB_REP
	return p.top
}

func (p *Parser) subParseNumber(pat []rune, bitSize int) (int64, error) {
	// for bitSize, see 'go doc strconv.ParseInt', we also set up our filters
	// with it though

	nreg := []rune{}
	switch {
	case bitSize == 8:
		for _, r := range pat {
			if '0' <= r && r <= '8' {
				nreg = append(nreg, r)
				p.i++
				if len(nreg) >= 3 {
					break
				}
			} else {
				break
			}
		}
	case bitSize == 16:
		// XXX: we should change from 2 to 4 chars here if we see \x{1234} (unicode)
		// XXX: such matching could possibly do the \u1234 syntax too
		for _, r := range pat {
			if ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F') {
				nreg = append(nreg, r)
				p.i++
				if len(nreg) >= 2 {
					break
				}
			} else {
				break
			}
		}
	}

	return strconv.ParseInt(string(nreg), bitSize, 0)
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
				p.RTop().AddDotState()
			case p.r == '\n':
				p.RTop().AddRuneState(p.r) // this may be $ sometimes with /m
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
					p.RTop().AddRuneState(p.r)
					p.n = SUB_REP
				} else {
					p.PushContext(CTX_NQUANT)
				}

			case p.r == '?':
				switch {
				case p.n == SUB_QTY:
					s := p.top.States[len(p.top.States)-1]
					s.Greedy = false
					p.n = SUB_INIT
				case p.n == SUB_REP:
					p.QTop().SetQty(0, 1)
				default:
					return p.formatError("quantifier without preceeding repeatable")
				}
			case p.r == '*':
				switch {
				case p.n == SUB_REP:
					p.QTop().SetQty(0, -1)
				default:
					return p.formatError("quantifier without preceeding repeatable")
				}
			case p.r == '+':
				switch {
				case p.n == SUB_REP:
					p.QTop().SetQty(1, -1)
				default:
					return p.formatError("quantifier without preceeding repeatable")
				}
			case p.r == ')':
				return p.formatError("")

			default:
				p.Printf(" AddRuneState(%c) default\n", p.r)
				p.RTop().AddRuneState(p.r)
			}
		case p.m == CTX_SLASHED:
			switch {
			case p.r == 'x':
				if num, err := p.subParseNumber(pat[p.i+1:len(p.pat)], 16); err == nil {
					p.r = rune(num)
					p.Printf(" AddRuneState(%d) hex\n", p.r)
					p.RTop().AddRuneState(p.r)
				} else {
					return p.formatError("failed to parse hex")
				}
			case '0' <= p.r && p.r <= '8':
				if num, err := p.subParseNumber(pat[p.i:len(p.pat)], 8); err == nil {
					p.r = rune(num)
					p.Printf(" AddRuneState(%d) octal\n", p.r)
					p.RTop().AddRuneState(p.r)
				} else {
					return p.formatError("failed to parse octal")
				}
			default:
				return p.formatError("TODO")
			}
			p.PopContext(false)
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
					return p.formatError("")
				}
			case p.r == '}':
				if len(p.m_rreg1) > 0 || len(p.m_rreg2) > 0 {
					if p.mode[len(p.mode)-1].n != SUB_REP {
						return p.formatError("quantifier without preceeding repeatable")
					}
					var a int = 0
					var b int = -1
					if num, err := strconv.ParseInt(string(p.m_rreg1), 10, 0); err == nil {
						a = int(num)
					}
					if num, err := strconv.ParseInt(string(p.m_rreg2), 10, 0); err == nil {
						b = int(num)
					}
					p.PopContext(false)
					p.QTop().SetQty(a, b)
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
