package lib

import (
	"errors"
)

type Item interface {
}

type Group struct {
	Items []Item
	Capno int
}

type MatchClass struct {
	Ranges [][2]int
}

type RE struct {
	Pattern string
	Items   []Item
}

func (re RE) Match(subject string) (m *Match) {
	return new(Match)
}

func Compile(pat string, flags int) (*RE, error) {
	ret := new(RE)
	if ret != nil {
		return ret, nil
	}
	return ret, errors.New("failcopter")
}
