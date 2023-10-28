package lib

import (
	"fmt"
	"os"
)

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

var nfaTrace bool

func makeNFA(whence Stateish) (ret *NFA) {
	ret = &NFA{Whence: whence, Transitions: make(map[*State][]*NFA)}
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] makeNFA(%s) => %s\n", GetTag(whence), GetTag(ret))
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

func (this *NFA) addTransitions(next *NFA) (leaf []*State) {
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] %s.addTransitions(%s)\n", GetTag(this), FTag(next))
	}
	switch typed := this.Whence.(type) {
	case *State:
		if nfaTrace {
			fmt.Fprintf(os.Stderr, "[DNTB]   %s -> %s\n", GetTag(typed), FTag(next))
		}
		if typed.Max > 0 || typed.Max < 0 {
			this.Transitions[typed] = append(this.Transitions[typed], next)
			if next == nil {
				if nfaTrace {
					fmt.Fprintf(os.Stderr, "[DNTB]   %s is a leaf\n", GetTag(typed))
				}
				leaf = append(leaf, typed)
			}
			if typed.Max < 0 || typed.Max > 1 {
				this.Transitions[typed] = append(this.Transitions[typed], this)
			}
		}
	case *Group:
		if typed.Max < 0 || typed.Max > 0 {
			for i, slist := range typed.States { // slist OR slist OR slist
				if len(slist) < 1 {
					continue
				}
				var last *NFA
				for j, sti := range slist { // sti . sti . sti
					nsti := this.FindNFA(sti)
					if last == nil {
						if nfaTrace {
							fmt.Fprintf(os.Stderr, "[DNTB]   %s -> %s\n", "ε", GetTag(nsti))
						}
						this.Transitions[nil] = append(this.Transitions[nil], nsti)
					} else {
						if nfaTrace {
							fmt.Fprintf(os.Stderr, "[DNTB]   %s.%s[%d,%d].%s => %s\n",
								GetTag(this), GetTag(typed), i, j, GetTag(last), GetTag(nsti))
						}
						for _, item := range last.addTransitions(nsti) {
							leaf = append(leaf, item)
						}
					}
					last = nsti
				}
				if nfaTrace {
					fmt.Fprintf(os.Stderr, "[DNTB]   %s.%s.%s => %s\n",
						GetTag(this), GetTag(typed), GetTag(last), FTag(next))
				}
				for _, item := range last.addTransitions(next) {
					leaf = append(leaf, item)
					if nfaTrace {
						fmt.Fprintf(os.Stderr, "[DNTB]   %s ~> %s\n", GetTag(item), GetTag(this))
					}
					nitem := this.FindNFA(item)
					nitem.Transitions[item] = append(nitem.Transitions[item], this)
				}
			}
		}
	}
	return
}

func BuildNFA(r *RE) (ret *NFA) {
	nfaTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("NFA_TRACE")
	defer func() { nfaTrace = false }()

	var last *NFA
	var this *NFA
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] BuildNFA :: start\n")
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
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] BuildNFA :: add accept\n")
	}
	last.addTransitions(nil)
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "\n")
	}
	return
}

func (r *RE) NFA() *NFA {
	if r.nfa == nil {
		r.nfa = BuildNFA(r)
	}
	return r.nfa
}
