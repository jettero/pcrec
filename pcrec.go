package pcrec

import "github.com/jettero/pcrec/lib"

func Match(pat string, subject string, flags int) (*obj.Match, error) {
	r, err := lib.Compile(pat, flags)
	if err != nil {
		return nil, err
	}
	return r.Match(subject)
}
