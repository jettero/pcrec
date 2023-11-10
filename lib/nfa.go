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

	Children     []*NFA
	Transitions  map[*State][]*NFATrans
	Whence       Stateish
	Capture      bool
	CaptureGroup int
}

type NFATrans struct {
	Capture []int
	NFA     *NFA
}

var nfaTrace bool

func makeNFA(whence Stateish, gctr *int) (ret *NFA) {
	ret = &NFA{Whence: whence, Transitions: make(map[*State][]*NFATrans)}
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] makeNFA(%s) => %s\n", GetTag(whence), GetTag(ret))
	}
	switch typed := whence.(type) {
	case *Group:
		if typed.Capture {
			ret.Capture = true
			ret.CaptureGroup = *gctr
			*gctr++
			if nfaTrace {
				fmt.Fprintf(os.Stderr, "[DNTB]   new capture group: %d\n", ret.CaptureGroup+1)
			}
		}
		for _, slist := range typed.States {
			for _, sti := range slist {
				if !TagDefined(sti) {
					ret.Children = append(ret.Children, makeNFA(sti, gctr))
				}
			}
		}
	}
	return
}

func (n *NFA) FindNFA(s Stateish) *NFA {
	if n.Whence == s {
		return n
	}
	for _, item := range n.Children {
		if item.Whence == s {
			return item
		}
	}
	return nil
}

func (this *NFA) nonGroupNodes() (ret []*NFA) {
	switch this.Whence.(type) {
	case *Group:
		for _,c := range this.Children {
			ret = append(ret, c.nonGroupNodes()...)
		}
		return
	}
	return []*NFA{ this }
}

func (this *NFA) addTransitions(next *NFA) (leaf []*State) {
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] %s.addTransitions(%s)\n",
			GetTag(this), GetFTag(next))
	}
	switch typed := this.Whence.(type) {
	case *State:
		if nfaTrace {
			fmt.Fprintf(os.Stderr, "[DNTB]   %s -> %s\n", GetTag(typed), GetFTag(next))
		}
		if typed.Max > 0 || typed.Max < 0 {
			this.Transitions[typed] = append(this.Transitions[typed], &NFATrans{NFA: next})
			if next == nil {
				if nfaTrace {
					fmt.Fprintf(os.Stderr, "[DNTB]   %s is a leaf\n", GetTag(typed))
				}
				leaf = append(leaf, typed)
			}
			if typed.Max < 0 || typed.Max > 1 {
				this.Transitions[typed] = append(this.Transitions[typed], &NFATrans{NFA: this})
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
						this.Transitions[nil] = append(this.Transitions[nil], &NFATrans{NFA: nsti})
						if typed.Max < 0 || typed.Max > 1 {
							for _, n := range nsti.nonGroupNodes() {
								n.Transitions[nil] = append(n.Transitions[nil], &NFATrans{NFA: this})
							}
						}
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
						GetTag(this), GetTag(typed), GetTag(last), GetFTag(next))
				}
				for _, item := range last.addTransitions(next) {
					leaf = append(leaf, item)
					if nfaTrace {
						fmt.Fprintf(os.Stderr, "[DNTB]   %s ~> %s\n", GetTag(item), GetTag(this))
					}
					nitem := this.FindNFA(item)
					nitem.Transitions[item] = append(nitem.Transitions[item], &NFATrans{NFA: this})
				}
			}
		}
	}
	return
}

func BuildNFA(r *RE) (ret *NFA) {
	nfaTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("NFA_TRACE")
	defer func() { nfaTrace = false }()

	var gctr int

	var last *NFA
	var this *NFA
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] BuildNFA :: start\n")
	}
	for _, stateish := range r.States {
		this = makeNFA(stateish, &gctr)
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
