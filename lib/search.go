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
	searchTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("RE_SEARCH_TRACE")
	res = &REsult{}

	if searchTrace {
		fmt.Printf("---=: SearchRunes(\"%s\")\n", PrintableizeRunes(candidate, 0))
	}

outer:
	for mpos, cpos := 0, 0; cpos < len(candidate); cpos++ {
		mpos = cpos
		res.Groups = res.Groups[:0]
		if searchTrace {
			fmt.Printf("  -- candidate[%d:]=\"%s\"\n", mpos, PrintableizeRunes(candidate[mpos:], 20))
		}
		for _, sta := range r.States {
			if adj, bak, ok := sta.SearchRunes(res, candidate[mpos:]); ok {
				mpos += adj
				if bak > 0 && searchTrace {
					fmt.Printf("    -- TODO mpos=%d bak=%d len=%d\n", mpos, bak, len(candidate))
				}
			} else {
				// fmt.Printf("    -- nomatch\n")
				continue outer // match states in order or longjump to the outer loop
			}
		}
		if searchTrace {
			fmt.Printf("    -- FIN: MATCHED\n")
		}
		res.Matched = true // if the States loop finishes, then we matched
		return             // so we only continue from the inner loop
	}
	if searchTrace {
		fmt.Printf("  -- FIN: nomatch\n")
	}
	res.Matched = false // this is implied, but spelled out because it looks cool
	return
}

func (g *Group) SearchRunes(res *REsult, candidate []rune) (adj int, bak int, ok bool) {
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
			if a, b, o := s.SearchRunes(res, candidate[adj:]); o {
				adj += a
				bak += b
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

func (s *State) SearchRunes(res *REsult, candidate []rune) (adj int, bak int, ok bool) {
	var max int = len(candidate)
	if s.Max >= 0 && s.Max < max {
		max = s.Max
	}
	for adj = 0; adj < max; adj++ {
		if s.Matches(candidate[adj]) {
			ok = true
			if adj > s.Min {
				bak = adj - s.Min
			}
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
