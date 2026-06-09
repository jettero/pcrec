package pcrec

import "github.com/jettero/pcrec/lib"

func Parse(pat string) (*lib.RE, error) {
	return lib.Parse(pat)
}

func Search(pat string, candidate string) (*lib.REsult, error) {
	re, err := Parse(pat)
	if err != nil {
		return nil, err
	}
	return re.Search(candidate), nil
}
