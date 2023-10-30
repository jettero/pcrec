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
			printableRunes := fmt.Sprintf("\"%s\"", PrintableizeRunes(candidate, 20, true))
			if len(candidate) < 1 {
				printableRunes = "<EOL>"
			}
			fmt.Printf("[SRCH] %s.Transitions[%s] => {%s}\n[SRCH]    candidate=%s\n",
				GetTag(nfa), GetTag(s), GetFTagList(nl), printableRunes)
		}
		if len(candidate) < 1 {
			break
		}
		if _, ub, matched := s.Matches(candidate); matched {
			for _, n := range nl {
				if n == nil {
					res.Matched = true
					fmt.Printf("[SRCH]    FIN\n")
					break
				}
				if n.continueSR(candidate[ub:], res); res.Matched {
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

func (s *State) Matches(candidate []rune) (lb int, ub int, match bool) {
	var q int
	var head string
	if searchTrace {
		head = fmt.Sprintf("[SRCH]    %s:", GetTag(s))
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
		fmt.Printf("%s => %v \"%s\" {%d,%d}\n", head, match,
			PrintableizeRunes(candidate[:q], 0, true), lb, ub)
	}
	return
}
