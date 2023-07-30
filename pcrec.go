package pcrec

import rl "github.com/jettero/pcrec/lib"

func Match(pat string, subject string, flags int) (*rl.Match, error) {
	re, err := rl.Compile(pat, flags)
	if err != nil {
		return nil, err
	}
	return re.Match(subject), nil
}
