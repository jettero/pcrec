package pcrec

import "github.com/jettero/pcrec/lib"

func Parse(pat string) (*lib.RE, error) {
	return lib.Parse([]rune(pat))
}

func Search(pat string, candidate string) (*lib.REsult, error) {
	if _, err := Parse(pat); err != nil {
		return nil, err
	} else {
		return &lib.REsult{}, nil
		//return n.SearchRunes([]rune(candidate)), nil
	}
}
