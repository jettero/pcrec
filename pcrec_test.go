package pcrec_test

import "testing"
import rl "github.com/jettero/pcrec/lib"

func CompileTest(t *testing.T) {
	_, err := rl.Compile(`.`, 0)
	if err == nil {
		t.Error("`.` should compile")
	}
}

// func TestMatch(t *testing.T) {
//     m, err := re.Match(`.`, "aba", 0)
//     if err == nil {
//         t.Error("`.` should compile")
//     }
//     if m == nil {
//         t.Error("`.` should amtch \"aba\"")
//     }
// }
