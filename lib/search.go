package lib

import "fmt"

type REsult struct { // <--- I think this is hilarious, sorry
	Groups  []*[]rune
	Matched bool
}

func (n *NFA) Search(candidate string) (ret *REsult) {
	return n.SearchRunes([]rune(candidate))
}

var searchTrace bool

func (n *NFA) SearchRunes(candidate []rune) (res *REsult) {
	searchTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("RE_SEARCH_TRACE")
	res = &REsult{}

	if searchTrace {
		fmt.Printf("---=: SearchRunes(\"%s\")\n", PrintableizeRunes(candidate, 0))
	}

outer:
	for cpos := 0; cpos < len(candidate); cpos++ {
		mpos := cpos
		res.Groups = res.Groups[:0]
		if searchTrace {
			fmt.Printf("  -- candidate[%d:]=\"%s\"\n", cpos, PrintableizeRunes(candidate[cpos:], 20))
		}
		for _, sta := range n.States {
			if adj, ok := sta.SearchRunes(res, candidate[mpos:]); ok {
				mpos += adj
			} else {
				continue outer // match states in order or longjump to the outer loop
			}
		}
		res.Matched = true // if the States loop finishes, then we matched
		return             // so we only continue from the inner loop
	}

	res.Matched = false // this is implied, but spelled out because it looks cool
	return
}

func (g *Group) SearchRunes(res *REsult, candidate []rune) (adj int, ok bool) {
	/// g.States[0][∀] || g.States[1][∀] || …
	var cidx int
	if g.Capture {
		// we don't know what the capture actualy is yet, but we make room
		// for it at the position of the group
		res.Groups = append(res.Groups, nil)
		cidx = len(res.Groups) - 1
	}
	for _, sl := range g.States {
		ok = true // assume this whole chain is true
		for _, s := range sl {
			if adj_, ok_ := s.SearchRunes(res, candidate[adj:]); ok_ {
				adj += adj_
			} else {
				ok = false // until we learn it's not true
				break
			}
		}
		if ok { // seems that whole chain matched
			if g.Capture {
				// replace the empty string we put in the REsult (above)
				matched := candidate[:adj]
				res.Groups[cidx] = &matched
			}
			return // adj,true
		}
		adj = 0 // backtrack
		// oddly, if the group didn't match, we still leave the empty capture
		// result in the REsult
	}
	return // 0,false
}

func (s *State) SearchRunes(res *REsult, candidate []rune) (adj int, ok bool) {
	var max int = len(candidate)
	if s.Max >= 0 && s.Max < max {
		max = s.Max
	}
	for adj = 0; adj < max; adj++ {
		if s.Matches(candidate[adj]) {
			ok = true
		} else if adj < s.Min {
			ok = false
			break
		} else {
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
