package lib

import "fmt"

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

	for cpos := 0; cpos < len(candidate); cpos++ {
		r.NFA().SearchRunes(candidate[cpos:], res)
		if res.Matched {
			break
		}
	}

	return res
}

func (nfa *NFA) SearchRunes(candidate []rune, res *REsult) {
	searchTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("SEARCH_TRACE")
	defer func() { searchTrace = false }()
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
			fmt.Printf("    -- %s", s.Describe(0))
		}
		if m.Matches(r) {
			if searchTrace {
				fmt.Printf(" => matched(%s)\n", Printableize(r, true))
			}
			return true
		}
	}
	if searchTrace {
		fmt.Printf(" => fail(%s)\n", Printableize(r, true))
	}
	return false
}
