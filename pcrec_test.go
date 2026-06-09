package pcrec_test

import (
	"fmt"
	"testing"

	"github.com/jettero/pcrec"
)

type matchCase struct {
	pat    string
	input  string
	match  bool
	span   [2]int // rune offsets, only checked if match=true
	groups []string
}

func sg(s string) string { return s }

var basicMatches = []matchCase{
	// literal
	{"a", "a", true, [2]int{0, 1}, nil},
	{"a", "xay", true, [2]int{1, 2}, nil},
	{"a", "xyz", false, [2]int{}, nil},
	{"ab", "xabc", true, [2]int{1, 3}, nil},

	// any
	{".", "a", true, [2]int{0, 1}, nil},
	{".", "", false, [2]int{}, nil},

	// classes
	{"[abc]", "x b y", true, [2]int{2, 3}, nil},
	{"[a-c]", "zzzbzzz", true, [2]int{3, 4}, nil},
	{"[^abc]", "abcdef", true, [2]int{3, 4}, nil},
	{"[a-z0-9]", "...4...", true, [2]int{3, 4}, nil},

	// alternation
	{"a|b", "xby", true, [2]int{1, 2}, nil},
	{"a|b|c", "xcy", true, [2]int{1, 2}, nil},

	// groups
	{"(a)", "xay", true, [2]int{1, 2}, []string{"a"}},
	{"(a|b)", "xby", true, [2]int{1, 2}, []string{"b"}},
	{"(ab|cd)", "xcdy", true, [2]int{1, 3}, []string{"cd"}},
	{"(ab|cd)", "xabxcdy", true, [2]int{1, 3}, []string{"ab"}}, // leftmost-first
	{"ab|cd", "xycdz", true, [2]int{2, 4}, nil},
	{"a(bc|de)f", "xadefy", true, [2]int{1, 5}, []string{"de"}},
	{"(x(y|z))+", "abxyxzab", true, [2]int{2, 6}, []string{"xz", "z"}},

	// quantifiers
	{"a*", "bbbaaa", true, [2]int{0, 0}, nil}, // empty match at start
	{"a+", "bbbaaa", true, [2]int{3, 6}, nil},
	{"a?", "bbb", true, [2]int{0, 0}, nil},
	{"a{2}", "aaaa", true, [2]int{0, 2}, nil},
	{"a{2,4}", "aaaaaa", true, [2]int{0, 4}, nil},
}

func TestBasicMatches(t *testing.T) {
	for i, tc := range basicMatches {
		t.Run(fmt.Sprintf("%02d/%s/%s", i, tc.pat, tc.input), func(t *testing.T) {
			re, err := pcrec.Parse(tc.pat)
			if err != nil {
				t.Fatalf("parse %q: %v", tc.pat, err)
			}
			res := re.Search(tc.input)
			if res.Matched != tc.match {
				t.Fatalf("Matched=%v want %v; result=%s\nre=%s", res.Matched, tc.match, res.Describe(0), re.Describe(0))
			}
			if !tc.match {
				return
			}
			if res.Start != tc.span[0] || res.End != tc.span[1] {
				t.Errorf("span [%d,%d] want [%d,%d]\n%s", res.Start, res.End, tc.span[0], tc.span[1], res.Describe(0))
			}
			for gi, want := range tc.groups {
				if gi >= len(res.Groups) {
					t.Errorf("group %d missing; have %d groups", gi+1, len(res.Groups))
					continue
				}
				got := string(res.Groups[gi])
				if got != want {
					t.Errorf("group %d = %q want %q", gi+1, got, want)
				}
			}
		})
	}
}
