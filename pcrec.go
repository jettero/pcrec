package pcrec

import "github.com/jettero/pcrec/lib"

func Compile(pat string) (*lib.NFA, error) {
	return lib.Parser.ParseString("-", pat)
}
