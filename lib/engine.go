package lib

type REsult struct { // <--- I think this is hilarious, sorry
	Group   []string
	Matched bool
}

type Sinfo struct {
	state Stateish
	i     int
	j     int
}

func (n *NFA) SearchRunes(candidate []rune) (ret *REsult) {
	// NOTE: don't try to make this recursive, Matches() is fine during RE
	// parsing, but it'd be too slow to use here. Aim for tail recursion.
	ret = &REsult{}
	sil := []Sinfo{{state: n}}

	for i := 0; i < len(candidate); {
		// r := candidate[i]
		si := sil[len(sil)-1]
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
			case *State:
				return
			}
		}
		return
	}
	return
}

func (n *NFA) Search(candidate string) (ret *REsult) {
	return n.SearchRunes([]rune(candidate))
}
