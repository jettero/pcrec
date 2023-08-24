package pcrec_test

import (
	"fmt"
	"github.com/jettero/pcrec"
	"testing"
)

var shouldCompileBase = []string{`.`, `a`, `ab`, `[a]`, `[ab]`}
var shouldCompileQuant = []string{"", "?", "*", "+", "*?", "+?"}

func TestCompile(t *testing.T) {
	for _, pat := range shouldCompileBase {
		for _, mod := range shouldCompileQuant {
			cpat := fmt.Sprintf("%s%s", pat, mod)
			t.Run(fmt.Sprintf("pcrec.Compile(pat=`%s`)", cpat), func(t *testing.T) {
				_, err := pcrec.Compile(cpat)
				if err != nil {
					t.Error(fmt.Sprintf("`%s` should compile but didn't: %s", cpat, err))
				}
			})
		}
	}
}
