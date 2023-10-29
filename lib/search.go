package lib

import (
	"fmt"
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
	return r.NFA().SearchRunes(candidate, false)
}

func (nfa *NFA) continueSR(candidate []rune, res *REsult) {
	for s, nl := range nfa.Transitions {
		if searchTrace {
			fmt.Printf("[SRCH] %s.Transitions[%s] => {%s}\n",
				GetTag(nfa), GetTag(s), GetFTagList(nl))
		}
		if s.Matches(candidate[0]) {
			for _, n := range nl {
				if n == nil {
					res.Matched = true
					fmt.Printf("[SRCH]    FIN\n")
					break
				}
				if n.continueSR(candidate[1:], res); res.Matched {
					break
				}
			}
		}
	}
}

func (nfa *NFA) SearchRunes(candidate []rune, anchored bool) (res *REsult) {
	searchTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("SEARCH_TRACE")
	defer func() { searchTrace = false }()

	res = &REsult{}

	if searchTrace {
		fmt.Printf("[SRCH] --------=: search :=--------\n")
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

func (s *State) Matches(r rune) bool {
	var head string
	if searchTrace {
		head = fmt.Sprintf("[SRCH]    %s:", GetTag(s))
	}
	for _, m := range s.Match {
		if m.Matches(r) {
			if searchTrace {
				// mstr := strings.ReplaceAll(s.Describe(0), "\n", "\n" + head)
				fmt.Printf("%s => matched(%s)\n", head, Printableize(r, true))
			}
			return true
		}
	}
	if searchTrace {
		fmt.Printf("%s => fail(%s)\n", head, Printableize(r, true))
	}
	return false
}
