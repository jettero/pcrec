package pcrec

import "github.com/jettero/pcrec/lib"

func Parse(pat []rune) (*lib.NFA, error) {
	return lib.Parse(pat)
}

func ParseString(pat string) (*lib.NFA, error) {
	return Parse([]rune(pat))
}
