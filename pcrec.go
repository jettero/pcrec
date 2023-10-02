package pcrec

import "github.com/jettero/pcrec/lib"

func Parse(pat string) (*lib.NFA, error) {
	return lib.Parse([]rune(pat))
}

func Search(pat string, candidate string) (*lib.REsult, error) {
	if n, err := Parse(pat); err != nil {
		return nil, err
	} else {
		return n.SearchRunes([]rune(candidate)), nil
	}
}
