package lib

import (
	"fmt"
	"strings"
)

type REsult struct {
	Matched    bool
	Start, End int      // rune offsets of the match
	Groups     [][]rune // captured substrings for groups 1..N (nil = unmatched)
}

func (r *REsult) Describe(indent int) string {
	is := strings.Repeat("  ", indent)
	if !r.Matched {
		return is + "no match"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%smatch [%d,%d]", is, r.Start, r.End)
	for i, g := range r.Groups {
		if g == nil {
			fmt.Fprintf(&b, "\n%s  $%d: <unset>", is, i+1)
		} else {
			fmt.Fprintf(&b, "\n%s  $%d: %q", is, i+1, string(g))
		}
	}
	return b.String()
}

func (r *RE) Search(s string) *REsult {
	return r.SearchRunes([]rune(s))
}

func (r *RE) SearchRunes(input []rune) *REsult {
	res := &REsult{}
	if r == nil || r.Root == nil {
		return res
	}
	m := &matcher{
		input:    input,
		captures: make([]Capture, r.NumCaptures+1),
	}
	for start := 0; start <= len(input); start++ {
		for i := range m.captures {
			m.captures[i] = Capture{-1, -1}
		}
		m.captures[0].Start = start
		matched := r.Root.match(m, start, func(np int) bool {
			m.captures[0].End = np
			return true
		})
		if matched {
			res.Matched = true
			res.Start = m.captures[0].Start
			res.End = m.captures[0].End
			res.Groups = make([][]rune, r.NumCaptures)
			for i := 0; i < r.NumCaptures; i++ {
				c := m.captures[i+1]
				if c.Start >= 0 && c.End >= c.Start {
					res.Groups[i] = input[c.Start:c.End]
				}
			}
			return res
		}
	}
	return res
}
