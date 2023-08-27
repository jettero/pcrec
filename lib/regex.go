package lib

type NFA struct {
	States []*State
}

type Matcher struct {
	min *rune
	max *rune
}

type State struct {
	Match []*Matcher
	min   *int
	max   *int
}

const (
	start int = iota
	slashed
	class
)

func Parse(s string) (*NFA, error) {
	var mode int
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case mode == start:
			switch {
			case r == '\\':
				mode = slashed
			case r == '.':
			case r == '[':
			case r == ']':
			case r == '(':
			case r == ')':
			case r == '{':
			case r == '}':
			case r == '?':
			case r == '*':
			case r == '+':
			}
		case mode == slashed:
			switch {
			case r == 'u':
			}
		}
		i++
	}
	return new(NFA), nil // everything compiles! gratz
}
