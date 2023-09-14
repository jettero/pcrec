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
	m_sreg  *State
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

func (p *Parser) Top(sub int) *NFA {
	p.n = sub
	return p.top
}

func (p *Parser) subParseNumber(subpat []rune, bitSize int) (int64, error) {
	// for bitSize, see 'go doc strconv.ParseInt', we also set up our filters
	// with it though
	nreg := []rune{}
	var lim int
	var needsClosingBrace bool
	switch {
	case bitSize == 8:
		for _, r := range subpat {
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
		// XXX: such matching could possibly do the \u1234 syntax too
		lim = 2
		p.Printf(" HEX")
		for _, r := range subpat {
			if r == '{' {
				lim = 4
				needsClosingBrace = true
				p.Printf(" UNICODE")
				p.i++ // consume '{'
			} else if ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F') {
				nreg = append(nreg, r)
				p.i++ // consume r
				if len(nreg) >= lim {
					p.Printf(" l-break %d", p.i)
					break
				}
			} else {
				p.Printf(" d-break %d", p.i)
				break
			}
		}
	}

	if needsClosingBrace {
		p.Printf(" NEED-CLOSE-BRACE \"...%s\"", string(p.pat[p.i:len(p.pat)]))
		if p.pat[p.i] == '}' {
			p.i++ // consume '}'
			p.Printf(" CLOSE-BRACE %d", p.i)
		} else {
			p.Printf(" ERROR-BRACE\n")
			return 0, fmt.Errorf("failed to parse \\x{...} syntax")
		}
	}

	p.Printf(" parse-num\n")
	if p.i < len(p.pat) {
		p.r = p.pat[p.i]
	} else {
		p.r = 0
	}

	return strconv.ParseInt(string(nreg), bitSize, 0)
}

func (p *Parser) GrokSlashed() ([]rune, bool, error) {
	var ret []rune
	var old_i int = p.i
	switch p.r {
	case 'x':
		p.i++ // skip the 'x'
		if num, err := p.subParseNumber(p.pat[p.i:len(p.pat)], 16); err == nil {
			ret = append(ret, rune(num))
			p.Printf(" GrokSlashed(hex) => %+v\n", ret)
			return ret, false, nil
		} else {
			return nil, false, err
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if num, err := p.subParseNumber(p.pat[p.i:len(p.pat)], 8); err == nil {
			ret = append(ret, rune(num))
			p.Printf(" GrokSlashed(oct) => %+v\n", ret)
			return ret, false, nil
		} else {
			p.i = old_i
			return nil, false, err
		}
	case 'g', 'a':
		ret = append(ret, '\a')
		p.Printf(" GrokSlashed(\\a) => %+v\n", ret)
		return ret, false, nil
	case 't':
		ret = append(ret, '\t')
		p.Printf(" GrokSlashed(\\t) => %+v\n", ret)
		return ret, false, nil
	case 'r':
		ret = append(ret, '\r')
		p.Printf(" GrokSlashed(\\r) => %+v\n", ret)
		return ret, false, nil
	case 'n':
		ret = append(ret, '\n')
		p.Printf(" GrokSlashed(\\n) => %+v\n", ret)
		return ret, false, nil
	case 's':
		ret = append(ret, ' ', '\t', '\n', '\r')
		p.Printf(" GrokSlashed(\\s) => %+v\n", ret)
		return ret, false, nil
	case 'S':
		ret = append(ret, ' ', '\t', '\n', '\r')
		p.Printf(" GrokSlashed(\\S) => !%+v\n", ret)
		return ret, true, nil
	default:
		ret := []rune{p.r}
		p.Printf(" GrokSlashed(??) => %+v\n", ret)
		return ret, false, nil
	}
}

func (p *Parser) Parse(pat []rune) (*NFA, error) {
	p.trace = TruthyEnv("PCREC_TRACE") || TruthyEnv("RE_PARSE_TRACE")
	p.mode = []Context{{m: CTX_NONE, i: 0}}
	p.top = &NFA{}
	p.pat = pat

	for p.i = 0; p.i < len(p.pat); p.i++ {
		p.r = p.pat[p.i]

		switch p.m {
		case CTX_NONE:
			switch p.r {
			case '[':
				p.PushContext(CTX_CCLASS)
			case '(':
				p.PushContext(CTX_GROUP)
			case '\\':
				p.PushContext(CTX_SLASHED)
			case '{':
				if p.n == SUB_RET {
					p.Printf(" AddRuneState(%c) NQUANT-RETURN\n", p.r)
					p.Top(SUB_REP).AddRuneState(p.r)
				} else {
					p.PushContext(CTX_NQUANT)
				}
			case '?':
				switch {
				case p.n == SUB_QTY:
					p.Top(SUB_INIT).SetGreedy(false)
				case p.n == SUB_REP:
					p.Top(SUB_QTY).SetQty(0, 1)
				default:
					return p.formatError("quantifier without preceeding repeatable")
				}
			case '*':
				switch {
				case p.n == SUB_REP:
					p.Top(SUB_QTY).SetQty(0, -1)
				default:
					return p.formatError("quantifier without preceeding repeatable")
				}
			case '+':
				switch {
				case p.n == SUB_REP:
					p.Top(SUB_QTY).SetQty(1, -1)
				default:
					return p.formatError("quantifier without preceeding repeatable")
				}
			case ')':
				return p.formatError("")
			case '.':
				p.Top(SUB_REP).AddDotState()
			case '\n':
				p.Top(SUB_REP).AddRuneState(p.r) // this may be $ sometimes with /m
			default:
				p.Printf(" AddRuneState(%c) default\n", p.r)
				p.Top(SUB_REP).AddRuneState(p.r)
			}

		case CTX_CCLASS:
			if p.n == SUB_INIT {
				p.n = SUB_LHS
			}
			switch p.n {
			case SUB_LHS:
				p.m_sreg = p.Top(SUB_RHS).AddRuneState(p.r)
			case SUB_RHS:
				switch p.r {
				case '-':
					// XXX
				case ']':
					p.n = SUB_LHS
				default:
					if err := p.m_sreg.AppendToLastMatch(p.r, false); err != nil {
						return p.formatError(err.Error())
					}
				}
			}

		case CTX_SLASHED:
			runes, inverted, err := p.GrokSlashed()
			if err != nil {
				return p.formatError(err.Error())
			}
			p.r = runes[0]
			if inverted {
				p.Printf(" AddRuneState(%+v) CTX_SLASHED\n", runes)
				p.Top(SUB_REP).AddInvertedRuneState(runes...)
			} else {
				p.Printf(" AddInvertedRuneState(%+v) CTX_SLASHED\n", runes)
				p.Top(SUB_REP).AddRuneState(runes...)
			}

		case CTX_NQUANT:
			if p.n == SUB_INIT {
				p.m_rreg1 = []rune{}
				p.m_rreg2 = []rune{}
				p.n = SUB_LHS
			}
			switch p.r {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				switch p.n {
				case SUB_LHS:
					p.m_rreg1 = append(p.m_rreg1, p.r)
				case SUB_RHS:
					p.m_rreg2 = append(p.m_rreg2, p.r)
				}
			case ',':
				switch p.n {
				case SUB_LHS:
					p.n = SUB_RHS
				default:
					return p.formatError("")
				}
			case '}':
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
					p.Top(SUB_QTY).SetQty(a, b)
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
