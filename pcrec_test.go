package pcrec_test

import (
	"fmt"
	"testing"

	"github.com/jettero/pcrec"
)

var shouldCompile = []string{
	`(a|b|ab|(a)(b)|(ab)|[ab][a][b])`,
	`[abcdefABCDEF]`,
	`[a-fA-F]`,
	`[a]`,
	`[ab]`,
	`[a-b]`,
	`[a-b-]`,

	// seems ludicrous, but these should compile
	// the seemingly erroneous symbols just become literals
	`.{}`, `.{,}`, `.]`, `.}`,
	`a{}`, `a{,}`, `a]`, `a}`,
}
var shouldNotCompile = []string{
	`)`,
	`.)`,
	`ab)`,
	`(`,
	`(.`,
	`(ab`,
	`*ab`,
	`?ab`,
	`ab)`,
	`.**`,
	`.+*`,
	`.*+`,
	`.?+`,
	`.?*`,
	`.{1,2}{1,2}`,
}

var letteredEscapes = []string{`a`, `g`, `t`, `r`, `n`, `0`, `s`}
var populateCompileBase = []string{`.`, `a`, `ab`, `[a]`, `[ab]`, `[a-b]`}
var populateCompileQuant = []string{"", "?", "*", "+", "*?", "+?", "{2,}", "{,3}", "{2,3}"}
var patQuantFormats = []string{"%s%s", "(%s%s)", "(%s)%s"}

func populateShouldCompile() {
	for _, s := range letteredEscapes {
		populateCompileBase = append(populateCompileBase, fmt.Sprintf("\\%s", s))
	}
	for _, gfmt := range patQuantFormats {
		for _, pat := range populateCompileBase {
			for _, mod := range populateCompileQuant {
				cpat := fmt.Sprintf(gfmt, pat, mod)
				shouldCompile = append(shouldCompile, cpat)
			}
		}
	}
}

func TestCompile(t *testing.T) {
	populateShouldCompile()
	for _, pat := range shouldCompile {
		t.Run(fmt.Sprintf("pat=`%s`", pat), func(t *testing.T) {
			_, err := pcrec.Parse(pat)
			if err != nil {
				t.Error(fmt.Sprintf("`%s` should compile but did not:\n%s", pat, err))
			}
		})
	}
}

func TestNotCompile(t *testing.T) {
	for _, pat := range shouldNotCompile {
		t.Run(fmt.Sprintf("pat=`%s`", pat), func(t *testing.T) {
			_, err := pcrec.Parse(pat)
			if err == nil {
				t.Error(fmt.Sprintf("`%s` should not compile but did", pat))
			}
		})
	}
}

type LineMatchThing struct {
	line    string
	pat     string
	matches bool
	groups  []*string
}

func sp(s string) *string {
	return &s
}

// I should find a way to read tese from a csv or yaml or something
var LMT []LineMatchThing = []LineMatchThing{
	{"Testacular test blarg19 s3v3n.", "test", true, []*string{}},
	{"Testacular test blarg19 s3v3n.", "(test)", true, []*string{sp("test")}},
	{"Testacular test blarg19 s3v3n.", "(t(xx|es)t)", true, []*string{sp("test"), sp("es")}},
	{"Testacular test blarg19 s3v3n.", "(t(xx)t|t(es)t)", true, []*string{sp("test"), nil, sp("es")}},
	{"Testacular test blarg19 s3v3n.", "t(xx)t|t(es)t", true, []*string{nil, sp("es")}},
	{"Testacular test blarg19 s3v3n.", `(\w+)`, true, []*string{sp("test")}},
	{"Testacular test blarg19 s3v3n.", `(\D+)`, true, []*string{sp("test")}},
	{"Testacular test blarg19 s3v3n.", `(\w+)(\d+)`, true, []*string{sp("test"), sp("19")}},
}

func TestLineMatchThings(t *testing.T) {
	for i, lmt := range LMT {
		t.Run(fmt.Sprintf("%03d/%s", i, lmt.pat), func(t *testing.T) {
			res, err := pcrec.Search(lmt.pat, lmt.line)
			if err != nil {
				t.Error(fmt.Sprintf("`%s` should compile but did not", lmt.pat))
			} else {
				if lmt.matches && !res.Matched {
					t.Error(fmt.Sprintf("`%s` should match `%s`, but did not", lmt.pat, lmt.line))
				} else if !lmt.matches && res.Matched {
					t.Error(fmt.Sprintf("`%s` should not match `%s`, but did", lmt.pat, lmt.line))
				} else if len(lmt.groups) != len(res.Groups) {
					t.Error(fmt.Sprintf("|lmt.groups|=%d != |res.Groups|=%d", len(lmt.groups), len(res.Groups)))
				}
			}
		})
	}
}
