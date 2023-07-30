package pcrec_test

import re "github.com/jettero/pcrec"

func TestMatch(t *testing.T) {
	m, err := re.Match(`.`, "aba", 0)
	if err == nil {
		t.Error("`.` should compile")
	}
	if m == nil {
		t.Error("`.` should amtch \"aba\"")
	}
}
