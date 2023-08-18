package pcrec_test

import (
	"fmt"
	"github.com/jettero/pcrec"
	"testing"
)

var shouldCompile = []string{
	`.`,
	`a`,
	`ab`,
	`[a]`,
	`[ab]`,
}

func TestCompile(t *testing.T) {
	for _, mod := range []string{"", "?", "*", "+", "*?", "+?"} {
		for _, pat := range shouldCompile {
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
