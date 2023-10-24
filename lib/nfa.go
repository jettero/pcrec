package lib

import "fmt"

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

	children    []*NFA
	Transitions map[*State][]*NFA
	Whence      Stateish
}

func makeNFA(whence Stateish) (ret *NFA) {
	ret = &NFA{Whence: whence, Transitions: make(map[*State][]*NFA)}
	if TruthyEnv("DEBUG_NFA_TRANSITION_BUILDER") {
		fmt.Printf("[DNTB] makeNFA(%s) => %s\n", GetTag(whence), GetTag(ret))
	}
	switch typed := whence.(type) {
	case *Group:
		for _, slist := range typed.States {
			for _, sti := range slist {
				if !TagDefined(sti) {
					ret.children = append(ret.children, makeNFA(sti))
				}
			}
		}
	}
	return
}

func FTag(item interface{}) (t string) {
	t = GetTag(item)
	if t[len(t)-1] == '?' {
		return "F"
	}
	return
}

func (n *NFA) FindNFA(s Stateish) *NFA {
	if n.Whence == s {
		return n
	}
	for _, item := range n.children {
		if item.Whence == s {
			return item
		}
	}
	return nil
}

func (this *NFA) addTransitions(next *NFA) {
	if TruthyEnv("DEBUG_NFA_TRANSITION_BUILDER") {
		fmt.Printf("[DNTB] %s.addTransitions(%s)\n", GetTag(this), FTag(next))
	}
	switch typed := this.Whence.(type) {
	case *State:
		if TruthyEnv("DEBUG_NFA_TRANSITION_BUILDER") {
			fmt.Printf("[DNTB]   %s -> %s\n", GetTag(typed), FTag(next))
		}
		this.Transitions[typed] = append(this.Transitions[typed], next)
	case *Group:
		for _, slist := range typed.States { // slist OR slist OR slist
			var last *NFA
			for _, sti := range slist { // sti . sti . sti
				nsti := this.FindNFA(sti)
				if last != nil {
					if TruthyEnv("DEBUG_NFA_TRANSITION_BUILDER") {
						fmt.Printf("[DNTB]   %s.%s.%s => %s\n", GetTag(this), GetTag(typed), GetTag(last), GetTag(nsti))
					}
					last.addTransitions(nsti)
				}
				last = nsti
			}
			last.addTransitions(next)
		}
	}
}

func BuildNFA(r *RE) (ret *NFA) {
	var last *NFA
	var this *NFA
	if TruthyEnv("DEBUG_NFA_TRANSITION_BUILDER") {
		fmt.Println("[DNTB] BuildNFA :: start")
	}
	for _, stateish := range r.States {
		this = makeNFA(stateish)
		if ret == nil {
			ret = this
		} else {
			last.addTransitions(this)
		}
		last = this
	}
	if TruthyEnv("DEBUG_NFA_TRANSITION_BUILDER") {
		fmt.Println("[DNTB] BuildNFA :: add accept")
	}
	last.addTransitions(nil)
	if TruthyEnv("DEBUG_NFA_TRANSITION_BUILDER") {
		fmt.Println("")
	}
	return
}
