package lib

import (
	"fmt"
	"strings"
)

var searchTrace bool

type REsult struct { // <--- I think this is hilarious, sorry
	Groups  []*[]rune
	Matched bool
}

func (r *RE) Search(candidate string) (ret *REsult) {
	return r.SearchRunes([]rune(candidate))
}

func (r *RE) SearchRunes(candidate []rune) (res *REsult) {
	res = &REsult{}
	r.NFA().SearchRunes(candidate, res, false)
	return res
}

func (nfa *NFA) SearchRunes(candidate []rune, res *REsult, anchored bool) {
	searchTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("SEARCH_TRACE")
	defer func() { searchTrace = false }()

	if searchTrace {
		fmt.Printf("[SRCH] --------=: search :=--------\n")
	}

	for cpos := 0; !res.Matched && cpos < len(candidate); cpos++ {
		for s, nl := range nfa.Transitions {
			for _, n := range nl {
				if searchTrace {
					fmt.Printf("[SRCH] cpos=%d nfa=%s [s=%s]=> n=%s\n",
						cpos, GetTag(nfa), GetTag(s), GetTag(n))
				}
			}
		}

		if anchored {
			break
		}
	}
}

func (m *Matcher) Matches(r rune) bool {
	if m.Any {
		return true
	}
	return m.Inverse != (m.First <= r && r <= m.Last) // inverse ^ between
}

func (s *State) Matches(r rune) bool {
	for _, m := range s.Match {
		if searchTrace {
			fmt.Printf("[SRCH] ** \"%s\"", strings.ReplaceAll(s.Describe(0), "\n", "\n[SRCH] "))
		}
		if m.Matches(r) {
			if searchTrace {
				fmt.Printf("[SRCH] => matched(%s)\n", Printableize(r, true))
			}
			return true
		}
	}
	if searchTrace {
		fmt.Printf("[SRCH] => fail(%s)\n", Printableize(r, true))
	}
	return false
}
