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
	Transitions  map[*OldState][]*NFATrans
	Whence       Stateish
	Capture      bool
	CaptureGroup int
}

type NFATrans struct {
	Capture []int
	NFA     *NFA
}

var nfaTrace bool

func makeTrans(nfa *NFA, caps ...int) *NFATrans {
	caps = append([]int{0}, caps...)
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB]   makeTrans(%s, %v)\n", GetTag(nfa), caps)
	}
	return &NFATrans{NFA: nfa, Capture: caps}
}

func makeNFA(whence Stateish, gctr *int, top *NFA) (ret *NFA) {
	ret = &NFA{Whence: whence, Transitions: make(map[*OldState][]*NFATrans)}
	if top == nil {
		top = ret
	}
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] makeNFA(%s) => %s\n", GetTag(whence), GetTag(ret))
	}
	switch typed := whence.(type) {
	case *Group:
		if typed.Capture {
			*gctr++
			ret.Capture = true
			ret.CaptureGroup = *gctr
			if nfaTrace {
				fmt.Fprintf(os.Stderr, "[DNTB]   new capture group: %d\n", ret.CaptureGroup)
			}
		}
		for _, slist := range typed.States {
			for _, sti := range slist {
				if stin := top.FindNFA(sti); stin == nil {
					ret.Children = append(ret.Children, makeNFA(sti, gctr, top))
				} else {
					ret.Children = append(ret.Children, stin)
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
	for _, item := range n.Children {
		if d := item.FindNFA(s); d != nil {
			return d
		}
	}
	return nil
}

func (this *NFA) leaves() (ret []*NFA) {
	if len(this.Children) == 0 {
		return []*NFA{this}
	}
	for _, c := range this.Children {
		ret = append(ret, c.leaves()...)
	}
	return
}

func (this *NFA) addTransitions(next *NFA) (leaf []*OldState) {
	if nfaTrace {
		fmt.Fprintf(os.Stderr, "[DNTB] %s.addTransitions(%s)\n",
			GetTag(this), GetFTag(next))
	}
	switch typed := this.Whence.(type) {
	case *OldState:
		if nfaTrace {
			fmt.Fprintf(os.Stderr, "[DNTB]   %s -> %s\n", GetTag(typed), GetFTag(next))
		}
		if typed.Max > 0 || typed.Max < 0 {
			this.Transitions[typed] = append(this.Transitions[typed], makeTrans(next))
			if next == nil {
				if nfaTrace {
					fmt.Fprintf(os.Stderr, "[DNTB]   %s is a leaf\n", GetTag(typed))
				}
				leaf = append(leaf, typed)
			}
			if typed.Max < 0 || typed.Max > 1 {
				this.Transitions[typed] = append(this.Transitions[typed], makeTrans(this))
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
						this.Transitions[nil] = append(this.Transitions[nil], makeTrans(nsti))
						if typed.Max < 0 || typed.Max > 1 {
							for _, n := range nsti.leaves() {
								if nfaTrace {
									fmt.Fprintf(os.Stderr, "[DNTB]   %s -> %s\n", "ε", GetTag(nsti))
								}
								n.Transitions[nil] = append(n.Transitions[nil], makeTrans(this))
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
					nitem.Transitions[item] = append(nitem.Transitions[item], makeTrans(this))
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
		this = makeNFA(stateish, &gctr, ret)
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
