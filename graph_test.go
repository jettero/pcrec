package pcrec_test

import (
	"testing"

	"github.com/jettero/pcrec/lib"
)

func TestDotCompileBuild(t *testing.T) {
	if g := lib.MakeGraph(); g == nil {
		t.Error("wut?")
	}
}
