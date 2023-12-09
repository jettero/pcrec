package lib

import (
	"fmt"
	"os"
)

var searchTrace bool
var searchIndent int

type REsult struct { // <--- I think this name is hilarious, sorry
	Groups  [][]rune
	Matched bool
}

func (r *RE) Search(candidate string) (ret *REsult) {
	return r.SearchRunes([]rune(candidate))
}

func (r *RE) SearchRunes(candidate []rune) (res *REsult) {
	return r.Graph.SearchRunes(candidate, false)
}

func (g *Graph) SearchRunes(candidate []rune, anchored bool) (res *REsult) {
	searchTrace = TruthyEnv("PCREC_TRACE") || TruthyEnv("PCREC_SEARCH_TRACE")
	defer func() { searchTrace = false }()

	res = &REsult{}

	if searchTrace {
		fmt.Fprintf(os.Stderr, "[SRCH] --------=: search :=--------\n")
		searchIndent = -1
	}

	for cpos := 0; !res.Matched && cpos < len(candidate); cpos++ {
		// if nfa.continueSR(candidate[cpos:], res); res.Matched || anchored {
		//     break
		// }
	}

	return
}
