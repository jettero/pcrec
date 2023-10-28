package lib

import (
	"fmt"
	"strconv"
)

const (
	CTX_NONE int = iota
	CTX_SLASHED
	CTX_CCLASS
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
	top   *RE
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

func (p *Parser) formatError(msg string) (*RE, error) {
	firstPart := "ERROR"
	if len(msg) > 0 {
		firstPart = fmt.Sprintf("ERROR: %s", msg)
	}
	p.Printf("  formatError() p.i=%d, p.r=%d, p.mn=%d.%d\n", p.i, p.r, p.m, p.n)
	if p.i >= len(p.pat) {
		return p.top, fmt.Errorf("%s at end of pattern", firstPart)
	}
	return p.top, fmt.Errorf("%s at position %d '%c'", firstPart, p.i+1, p.pat[p.i])
}

func (p *Parser) PushContextN(c int, n int) {
	p.mode = append(p.mode, Context{m: c, i: p.i, n: p.n})
	p.m = c
	p.n = n
	p.Printf("[PARS] PushContext(%d)\n", c)
}

func (p *Parser) PushContext(c int) {
	p.PushContextN(c, SUB_INIT)
}

func (p *Parser) PopContext(restore_i bool) {
	p.PopContextN(restore_i, SUB_INIT)
}

func (p *Parser) PopContextN(restore_i bool, n int) {
	if len(p.mode) > 1 {
		o := p.mode[len(p.mode)-1]
		p.mode = p.mode[:len(p.mode)-1]
		p.m = p.mode[len(p.mode)-1].m
		if restore_i {
			p.Seek(o.i - 1) // gets incremented immediately after pop, so go one early
			p.n = SUB_RET
		} else {
			p.n = n
		}
		p.Printf("[PARS] PopContext(%v)\n", restore_i)
	}
}

func (p *Parser) Printf(format string, args ...interface{}) {
	if p.trace {
		fmt.Printf(format, args...)
	}
}

func (p *Parser) Top(sub int) *RE {
	if sub >= SUB_INIT {
		p.n = sub
	}
	return p.top
}

func (p *Parser) subParseInt64(subpat []rune, bitSize int) (int64, error) {
	// for bitSize, see 'go doc strconv.ParseInt', we also set up our filters
	// with it though
	nreg := []rune{}
	var lim int
	var needsClosingBrace bool
	switch {
	case bitSize == 8:
		lim = 3
		for _, r := range subpat {
			if '0' <= r && r <= '8' {
				nreg = append(nreg, r)
				if len(nreg) >= lim {
					p.Printf("[PARS] l-break %d\n", p.i)
					break
				}
				p.Consume(1)
			} else {
				p.Printf("[PARS] d-break %d\n", p.i)
				p.Consume(-1) // put back the delimiter character
				break
			}
		}
	case bitSize == 16:
		// XXX: such matching could possibly do the \u1234 syntax too
		lim = 2
		p.Printf("[PARS] HEX\n")
		for _, r := range subpat {
			if r == '{' {
				lim = 4
				needsClosingBrace = true
				p.Printf("[PARS] UNICODE\n")
				p.Consume(1) // eat '{'
			} else if ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F') {
				nreg = append(nreg, r)
				if len(nreg) >= lim {
					p.Printf("[PARS] l-break %d\n", p.i)
					break
				}
				p.Consume(1)
			} else {
				p.Printf("[PARS] d-break %d\n", p.i)
				p.Consume(-1) // put back the delimiter character
				break
			}
		}
	}

	if needsClosingBrace {
		p.Printf("[PARS] NEED-CLOSE-BRACE \"...%s\"\n", string(p.pat[p.i:len(p.pat)]))
		if p.pat[p.i] == '}' {
			p.Consume(1)
			p.Printf("[PARS] CLOSE-BRACE %d\n", p.i)
		} else {
			p.Printf("[PARS] ERROR-BRACE\n")
			return 0, fmt.Errorf("failed to parse \\x{...} syntax")
		}
	}

	p.Printf("[PARS] parse-num\n")

	return strconv.ParseInt(string(nreg), bitSize, 0)
}

func (p *Parser) GrokSlashed() (ret []rune, inverse bool, rerr error) {
	var old_i int = p.i
	switch p.r {
	case 'x':
		p.Consume(1) // skip the 'x'
		if num, err := p.subParseInt64(p.pat[p.i:len(p.pat)], 16); err == nil {
			ret = append(ret, rune(num))
			p.Printf("[PARS] GrokSlashed(hex) => %+v\n", ret)
		} else {
			rerr = err
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if num, err := p.subParseInt64(p.pat[p.i:len(p.pat)], 8); err == nil {
			ret = append(ret, rune(num))
			p.Printf("[PARS] GrokSlashed(oct) => %+v\n", ret)
		} else {
			p.Seek(old_i)
			rerr = err
		}
	case 'g', 'a':
		ret = append(ret, '\a')
		p.Printf("[PARS] GrokSlashed(\\a) => %+v\n", ret)
	case 't':
		ret = append(ret, '\t')
		p.Printf("[PARS] GrokSlashed(\\t) => %+v\n", ret)
	case 'r':
		ret = append(ret, '\r')
		p.Printf("[PARS] GrokSlashed(\\r) => %+v\n", ret)
	case 'n':
		ret = append(ret, '\n')
		p.Printf("[PARS] GrokSlashed(\\n) => %+v\n", ret)
	case 'W':
		inverse = true
		p.Printf("[PARS] GrokSlashed(\\W) => !")
		fallthrough
	case 'w':
		for i := 'a'; i <= 'z'; i++ {
			ret = append(ret, i)
		}
		for i := 'A'; i <= 'Z'; i++ {
			ret = append(ret, i)
		}
		for i := '0'; i <= '9'; i++ {
			ret = append(ret, i)
		}
		ret = append(ret, '_')
		p.Printf("[PARS] GrokSlashed(\\w) => %+v\n", ret)
	case 'D':
		inverse = true
		p.Printf("[PARS] GrokSlashed(\\D) => !")
		fallthrough
	case 'd':
		for i := '0'; i <= '9'; i++ {
			ret = append(ret, i)
		}
		p.Printf("[PARS] GrokSlashed(\\d) => %+v\n", ret)
	case 'S':
		inverse = true
		p.Printf("[PARS] GrokSlashed(\\S) => !")
		fallthrough
	case 's':
		ret = append(ret, ' ', '\t', '\n', '\r')
		p.Printf("[PARS] GrokSlashed(\\s) => %+v\n", ret)
	default:
		ret := []rune{p.r}
		p.Printf("[PARS] GrokSlashed(??) => %+v\n", ret)
	}
	return
}

func (p *Parser) Seek(i int) {
	p.i = i
	if p.i >= 0 && p.i < len(p.pat) {
		p.r = p.pat[p.i]
	} else {
		p.r = 0
	}
}

func (p *Parser) Consume(i int) {
	p.i += i
	if p.i >= 0 && p.i < len(p.pat) {
		p.r = p.pat[p.i]
	} else {
		p.r = 0
	}
}

func (p *Parser) Parse(pat []rune) (*RE, error) {
	p.trace = TruthyEnv("PCREC_TRACE") || TruthyEnv("RE_PARSE_TRACE")
	p.mode = []Context{{m: CTX_NONE, i: 0}}
	p.top = &RE{}
	p.pat = pat

	p.Printf("[PARS] --------=: Parse(%s) :=--------\n", string(pat))

	for p.Seek(0); p.i < len(p.pat); p.Consume(1) {
		p.Printf("[PARS] -- p.pat[%d]: '%c'; |p.mode|: %d; p.mn: %d.%d\n", p.i, p.r, len(p.mode), p.m, p.n)

		switch p.m {
		case CTX_NONE:
			switch p.r {
			case '[':
				p.PushContext(CTX_CCLASS)
			case '(':
				p.Top(SUB_INIT).AppendGroup()
				p.Printf("[PARS] APPEND GROUP\n")
			case ')':
				p.Printf("[PARS] CLOSE GROUP\n")
				if err := p.Top(SUB_REP).CloseGroup(); err != nil {
					return p.formatError("unmatched closing parenthesis")
				}
			case '|':
				if p.Top(SUB_INIT).AppendOrToGroupOrCreateGroup() {
					p.Printf("[PARS] IMPLICIT TOP GROUP\n")
				}
			case '\\':
				p.PushContext(CTX_SLASHED)
			case '{':
				if p.n == SUB_RET {
					// the regex '{ab}' should work as literally the 4
					// character string to facilitate this, when {1,2} parsing
					// fails, we return here and add the dumb thing as
					// a literal.
					p.Printf("[PARS] AppendRuneState(%c) NQUANT-RETURN\n", p.r)
					p.Top(SUB_REP).AppendRuneState(p.r)
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
			case '.':
				p.Top(SUB_REP).AppendDotState()
			case '\n':
				// XXX: may mean '$' anchor, or sometimes is boring \s
				p.Top(SUB_REP).AppendRuneState(p.r)
			default:
				p.Top(SUB_REP).AppendRuneState(p.r)
			}

		case CTX_CCLASS:
			if p.n == SUB_INIT {
				p.m_sreg = p.Top(SUB_LHS).AppendRuneState(p.r)
				p.Printf("[PARS] INIT CCLASS\n")
			}
			switch p.r {
			case '\\':
				p.Consume(1)
				p.Printf("[PARS] GrokSlashed() in cclass p.i=%d, p.r=%d, p.mn=%d.%d\n",
					p.i, p.r, p.m, p.n)
				runes, inverted, err := p.GrokSlashed()
				if err != nil {
					return p.formatError(err.Error())
				}
				// XXX: what do we do if we get [abc\S] ?
				// this will be dead wrong, I think Matcher needs a rework
				// XXX: what does [^\S] even mean? Prolly it means \s, can we make it work?
				for _, r := range runes {
					switch p.n {
					case SUB_LHS:
						// Apparently AppendMatch tries to handle inverted internally, will this just work?
						p.m_sreg.AppendMatch(r, inverted)
					case SUB_RHS:
						// XXX: is [x-y] ever valid if y is \s or \g? sometimes maybe?
						return p.formatError("invalid range in character class")
					}
				}
			case ']':
				if p.n == SUB_RHS {
					p.m_sreg.AppendMatch('-', false)
					p.Printf("[PARS] APPEND DASH\n")
				}
				p.PopContextN(false, SUB_REP)
				p.Printf("[PARS] STOP CCLASS\n")
			default:
				switch p.n {
				case SUB_LHS:
					switch p.r {
					case '-':
						p.n = SUB_RHS
						p.Printf("[PARS] LHS->RHS\n")
					default:
						p.m_sreg.AppendMatch(p.r, false)
					}
				case SUB_RHS:
					if err := p.m_sreg.AppendToLastMatch(p.r, false); err != nil {
						return p.formatError(err.Error())
					}
					p.n = SUB_LHS
					p.Printf("[PARS] RHS->LHS\n")
				default:
					return p.formatError(fmt.Sprintf("internal error, bad cclas sub-state p.mn=%d,%d", p.m, p.n))
				}
			}

		case CTX_SLASHED:
			runes, inverted, err := p.GrokSlashed()
			if err != nil {
				return p.formatError(err.Error())
			}
			p.PopContext(false) // we're done with CTX_SLASHED
			if inverted {
				p.Top(SUB_REP).AppendInvertedRuneState(runes...)
			} else {
				p.Top(SUB_REP).AppendRuneState(runes...)
			}

		case CTX_NQUANT:
			if p.n == SUB_INIT {
				p.m_rreg1 = []rune{}
				p.m_rreg2 = []rune{}
				p.n = SUB_LHS
				p.Printf("[PARS] INIT NQUANT LHS\n")
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
					p.Printf("[PARS] INIT NQUANT RHS\n")
					p.n = SUB_RHS
				default:
					return p.formatError("unexpected comma")
				}
			case '}':
				if len(p.m_rreg1) > 0 || len(p.m_rreg2) > 0 {
					if p.mode[len(p.mode)-1].n != SUB_REP {
						return p.formatError("quantifier without preceeding repeatable")
					}
					var a int = -1337 // elite way to see if we missed something
					var b int = -1337
					if len(p.m_rreg1) == 0 {
						a = 0
					} else if num, err := strconv.ParseInt(string(p.m_rreg1), 10, 0); err == nil {
						a = int(num)
						b = a
					} else {
						return p.formatError(fmt.Sprintf("unable to parse \"%s\" as a number", string(p.m_rreg1)))
					}
					if len(p.m_rreg2) == 0 {
						b = -1
					} else if num, err := strconv.ParseInt(string(p.m_rreg2), 10, 0); err == nil {
						b = int(num)
					} else {
						return p.formatError(fmt.Sprintf("unable to parse \"%s\" as a number", string(p.m_rreg2)))
					}
					if a <= -1337 || b <= -1337 {
						panic("unknown error parsing quantity")
					}
					p.PopContext(false)
					p.Top(SUB_QTY).SetQty(a, b)
					p.Printf("[PARS] SetQty(%d,%d) NQUANT\n", a, b)
				} else {
					p.Printf("[PARS] abort NQUANT, reparse\n")
					p.PopContext(true)
				}
			}
		}
	}

	p.top.CloseImplicitTopGroups()

	if p.top.LastOpenGroup() != nil {
		return p.formatError("missing closing perenthesis")
	}

	return p.top, nil
}

func Parse(pat []rune) (*RE, error) {
	parser := Parser{}
	return parser.Parse(pat)
}
