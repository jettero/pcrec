package lib

type RE struct {
	States []Stateish
}

type Stateish interface {
	SetQty(min, max int)
	SetGreedy(f bool)
	LastStateish() Stateish
	LastOpenGroup() *Group
	Describe(indent int) string
	SearchRunes(res *REsult, candidate []rune) (int, int, bool)
}

type State struct {
	Match  []*Matcher // items in an 'or' group (e.g. a|b|c)
	Min    int        // min matches
	Max    int        // max matches or -1 for many
	And    bool       // a&b&c, useful for: [^abc] => (^a&^b&^c) ≡ ^(a|b|c)
	Greedy bool
}

type Group struct {
	States   [][]Stateish // []Stateish OR []Stateish OR …
	Min      int          // min matches
	Max      int          // max matches or -1 for many
	Capture  bool
	Greedy   bool
	Implicit bool // we can't close implicit groups with ')'

	// parser flags
	Closed bool
}

type Matcher struct {
	Inverse bool
	Any     bool
	First   rune
	Last    rune
}
