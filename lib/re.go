package lib

import (
	"fmt"
	"strings"
)

// RE is a compiled regex.
type RE struct {
	Root        Op
	NumCaptures int // not counting group 0 (the full match)
}

func (r *RE) Describe(indent int) string {
	if r == nil || r.Root == nil {
		return strings.Repeat("  ", indent) + "<empty RE>"
	}
	return fmt.Sprintf("%sRE: captures=%d\n%s",
		strings.Repeat("  ", indent), r.NumCaptures, r.Root.describe(indent+1))
}

// matcher is the per-search runtime state.
type matcher struct {
	input    []rune
	captures []Capture
}

// Capture is one group's span, in rune offsets. Start < 0 means unset.
type Capture struct{ Start, End int }
