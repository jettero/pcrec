package lib

type REsult struct { // <--- I think this is hilarious, sorry
	Group   []string
	Matched bool
}

type Sinfo struct {
	state Stateish
	i     int
	j     int
	p     int
}

func (n *NFA) SearchRunes(candidate []rune) (ret *REsult) {
	ret = &REsult{}
	sil := []Sinfo{{state: n}}
	si := sil[0]

	cpos := 0

	for {
		switch typed := si.state.(type) {
		case *NFA:
			if si.i < len(typed.States) {
				si = Sinfo{state: typed.States[si.i]}
				sil = append(sil, si)
				si.i++
			} else {
				return
			}
		case *Group:
			if si.i < len(typed.States) {
				if si.j < len(typed.States[i]) {
					si = Sinfo{state: typed.States[si.i][si.j]}
					sil = append(sil, si)
					si.j++
				} else {
					si.i++
				}
			} else {
				return
			}
		}
	}

	return
}

func (n *NFA) Search(candidate string) (ret *REsult) {
	return n.SearchRunes([]rune(candidate))
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
