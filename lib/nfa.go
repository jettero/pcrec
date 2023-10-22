package lib

type NFA struct {
	/********************************************************
	* · a finite set of states Q                            *
	* · a finite set of input symbols called the alphabet Σ *
	* · a transition function δ : Q × Σ → Q                 *
	* · an initial or start state q 0 ∈ Q q_{0}\in Q        *
	* · a set of accept states F ⊆ Q F\subseteq Q           *
	*                                                       *
	* That's what a DFA is ... but we're going to do NFA,   *
	* and prolly not quite like the above.                  *
	********************************************************/

	Transitions map[*State][]*NFA
	Whence      Stateish
}

func makeNFA(whence Stateish) (ret *NFA) {
	ret = &NFA{Whence: whence, Transitions: make(map[*State][]*NFA)}
	ret.addTransitions(whence, nil)
	return
}

func (this *NFA) addTransitions(stateish Stateish, next *NFA) {
	switch typed := stateish.(type) {
	case *State:
		this.Transitions[typed] = append(this.Transitions[typed], next)
	case *Group:
		for _, slist := range typed.States { // slist OR slist OR slist
			var first, last, ithis *NFA
			for _, sti := range slist { // sti . sti . sti
				ithis = makeNFA(sti)
				if first == nil {
					first = ithis
					this.addTransitions(sti, ithis)
				} else {
					last.addTransitions(sti, ithis)
				}
				last = ithis
			}
		}
	}
}

func BuildNFA(r *RE) (ret *NFA) {
	//	var last *NFA
	var this *NFA
	for _, stateish := range r.States {
		this = makeNFA(stateish)
		if ret == nil {
			ret = this
			// } else {
			//     last.addTransitions(stateish, this)
		}
		//		last = this
	}
	return
}
