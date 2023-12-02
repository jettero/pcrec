package lib

import (
	"fmt"
	"os"
	"strings"
)

var searchTrace bool
var searchIndent int

type REsult struct { // <--- I think this is hilarious, sorry
	Groups  [][]rune
	Matched bool
}

func (r *RE) Search(candidate string) (ret *REsult) {
	return r.SearchRunes([]rune(candidate))
}

func (r *RE) SearchRunes(candidate []rune) (res *REsult) {
	return r.NFA().SearchRunes(candidate, false)
}

func si(i int) string {
	return strings.Repeat(" ", searchIndent+i)
}

func xlateTrans(ntrl []*NFATrans) (ret []*NFA) {
	for _, ntr := range ntrl {
		ret = append(ret, ntr.NFA)
	}
	return
}

func (nfa *NFA) continueSR(candidate []rune, res *REsult) {
	searchIndent += 1
	defer func() { searchIndent -= 1 }()
	if len(candidate) < 1 {
		if searchTrace {
			fmt.Fprintf(os.Stderr, "[SRCH] %s%s <EOL>\n", si(0), GetTag(nfa))
		}
		return
	}
	for s, nl := range nfa.Transitions {
		if searchTrace {
			fmt.Fprintf(os.Stderr, "[SRCH] %s%s.Transitions[%s] => {%s} candidate=\"%s\"\n",
				si(0), GetTag(nfa), GetFTag(s), GetFTagList(xlateTrans(nl)), PrintableizeRunes(candidate, 20, true))
		}
		if lb, ub, matched := s.Matches(candidate); matched {
			for b := ub; b >= lb; b-- {
				for _, n := range nl {
					if searchTrace {
						fmt.Fprintf(os.Stderr, "[SRCH] %s%s.%s.%s {%d,%d}:%d \n",
							si(0), GetTag(nfa), GetFTag(s), GetFTag(n.NFA), lb, ub, b)
					}
					if n.NFA == nfa && b == 0 {
						if searchTrace {
							// use si(1) because we don't actually descend
							fmt.Fprintf(os.Stderr, "[SRCH] %sboring zero-width self transition\n", si(1))
						}
						continue
					}
					if n.NFA == nil {
						res.Matched = true
						if searchTrace {
							// we don't actually transition to F, so use s(1) to show the pretend descent
							fmt.Fprintf(os.Stderr, "[SRCH] %sFIN\n", si(1))
						}
						return
					}
					if n.NFA.continueSR(candidate[b:], res); res.Matched {
						if searchTrace {
							fmt.Fprintf(os.Stderr, "[SRCH] %s%s.Transitions[%s]: FIN\n", si(0), GetTag(nfa), GetFTag(s))
						}
						return
					}
				}
			}
		}
	}
}

func (nfa *NFA) SearchRunes(candidate []rune, anchored bool) (res *REsult) {
	searchTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("PCREC_SEARCH_TRACE")
	defer func() { searchTrace = false }()

	res = &REsult{}

	if searchTrace {
		fmt.Fprintf(os.Stderr, "[SRCH] --------=: search :=--------\n")
		searchIndent = -1
	}

	for cpos := 0; !res.Matched && cpos < len(candidate); cpos++ {
		if nfa.continueSR(candidate[cpos:], res); res.Matched || anchored {
			break
		}
	}
	return
}

func (m *Matcher) Matches(r rune) bool {
	if m.Any {
		return true
	}
	return m.Inverse != (m.First <= r && r <= m.Last) // inverse ^ between
}

func (s *OldState) Matches(candidate []rune) (lb int, ub int, match bool) {
	var q int
	if s == nil {
		lb = 0
		ub = 0
		match = true
		if searchTrace {
			fmt.Fprintf(os.Stderr, "[SRCH] %sε => true, {0,0}\n", si(0))
		}
		return
	}
qty:
	for q = 0; q < len(candidate) && (s.Max < 0 || q <= s.Max); q++ {
		for _, m := range s.Match {
			if m.Matches(candidate[q]) {
				continue qty
			}
		}
		break
	}
	if q >= s.Min {
		lb = s.Min
		ub = q
		match = true
	}
	if searchTrace {
		fmt.Fprintf(os.Stderr, "[SRCH] %s%s => %v, {%d, %d}\n", si(0), s.medium(), match, lb, ub)
	}
	return
}
