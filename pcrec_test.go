package pcrec_test

import (
	"fmt"
	"github.com/jettero/pcrec"
	"testing"
)

var shouldCompile = []string{
	`(a|b|ab|(a)(b)|(ab)|[ab][a][b])`,

	// seems ludicrous, but these should compile
	// the seemingly erroneous symbols just become literals
	`.{}`, `.{,}`, `.]`, `.}`,
	`a{}`, `a{,}`, `a]`, `a}`,
}
var shouldNotCompile = []string{
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
