package pcrec_test

import (
	"fmt"
	"testing"

	"github.com/jettero/pcrec"
	"github.com/jettero/pcrec/lib"
)

func TestDotCompileBuild(t *testing.T) {
	re, err := pcrec.Parse(`.`)
	if err != nil {
		t.Error("`.` failed to compile")
	}
	if nfa := lib.BuildNFA(re); nfa == nil {
		t.Error(fmt.Sprintf("lib.BuildNFA(%v) failed to build", re))
	}
}
