package pcrec_test

import (
	"testing"

	"github.com/jettero/pcrec/lib"
)

func TestGraph01(t *testing.T) {
	g := lib.MakeGraph()
	if g == nil {
		t.Error("making an empty graph should totally work every time... but didn't")
	}
}
