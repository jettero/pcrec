package pcrec_test

import (
	"fmt"
	"github.com/jettero/pcrec"
	"testing"
)

var shouldCompile = []string{
	`(a|b|ab|(a)(b)|(ab)|[ab][a][b])`,

	// seems ludicrous, but these should compile, but the {}'s are literal:
	`{}`, `{,}`,
	`.{}`, `.{,}`,
}
var shouldNotCompile = []string{`?ab`, `ab}`, `ab)`, `ab]`, `{}`, `{}ab`}

var populateCompileBase = []string{`.`, `a`, `ab`, `[a]`, `[ab]`, `[a-b]`}
var populateCompileQuant = []string{"", "?", "*", "+", "*?", "+?", "{2,}", "{,3}", "{2,3}"}

func populateShouldCompile() {
	for _, gfmt := range []string{"%s%s", "(%s%s)", "(%s)%s"} {
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
			_, err := pcrec.ParseString(pat)
			if err != nil {
				t.Error(fmt.Sprintf("`%s` should compile but did not: %s", pat, err))
			}
		})
	}
}

func TestNotCompile(t *testing.T) {
	for _, pat := range shouldNotCompile {
		t.Run(fmt.Sprintf("pat=`%s`", pat), func(t *testing.T) {
			_, err := pcrec.ParseString(pat)
			if err == nil {
				t.Error(fmt.Sprintf("`%s` should not compile but did", pat))
			}
		})
	}
}
