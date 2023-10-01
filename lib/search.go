package lib

type REsult struct { // <--- I think this is hilarious, sorry
	Groups  []string
	Matched bool
}

func (n *NFA) Search(candidate string) (ret *REsult) {
	return n.SearchRunes([]rune(candidate))
}

func (n *NFA) SearchRunes(candidate []rune) (res *REsult) {
	res = &REsult{}

	// States[0] && States[1] && …
	for cpos, npos := 0, 0; cpos < len(candidate) && npos < len(n.States); npos++ {
		if adj, ok := n.States[npos].SearchRunes(res, candidate[cpos:]); ok {
			cpos += adj
		} else {
			return // if any of these are false, matching has failed
		}
	}

	res.Matched = true
	return
}

func (g *Group) SearchRunes(res *REsult, candidate []rune) (adj int, ok bool) {
	/// g.States[0][∀] || g.States[1][∀] || …
	var cidx int
	for _, sl := range g.States {
		ok = true // assume this whole chain is true
		if g.Capture {
			res.Groups = append(res.Groups, "")
			cidx = len(res.Groups) - 1
		}
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
				res.Groups[cidx] = string(candidate[:adj])
			}
			return // adj,true
		}
		adj = 0 // backtrack
		if g.Capture {
			// remove our capture item if applicable
			res.Groups = append(res.Groups[:cidx], res.Groups[cidx+1:]...)
		}
	}
	return // 0,false
}

func (s *State) SearchRunes(res *REsult, candidate []rune) (int, bool) {
	if len(candidate) < 1 {
		return 0, false
	}
	if s.Matches(candidate[0]) {
		return 1, true
	}
	return 0, false
}

func (m *Matcher) Matches(r rune) bool {
	if m.Any {
		return true
	}
	return m.Inverse != (m.First <= r && r <= m.Last) // inverse ^ between
}

func (s *State) Matches(r rune) bool {
	for _, m := range s.Match {
		if m.Matches(r) {
			return true
		}
	}
	return false
}
