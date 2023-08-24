package pcrec_test

import (
	"fmt"
	"github.com/jettero/pcrec"
	"testing"
)

var shouldCompileBase = []string{`.`, `a`, `ab`, `[a]`, `[ab]`, `[a-b]`}
var shouldCompileQuant = []string{"", "?", "*", "+", "*?", "+?", "{2,}", "{,3}", "{2,3}"}

func TestCompile(t *testing.T) {
	for _, gfmt := range []string{"%s%s", "(%s%s)", "(%s)%s"} {
		for _, pat := range shouldCompileBase {
			for _, mod := range shouldCompileQuant {
				cpat := fmt.Sprintf(gfmt, pat, mod)
				t.Run(fmt.Sprintf("pcrec.Compile(pat=`%s`)", cpat), func(t *testing.T) {
					_, err := pcrec.Compile(cpat)
					if err != nil {
						t.Error(fmt.Sprintf("`%s` should compile but didn't: %s", cpat, err))
					}
				})
			}
		}
	}
}
