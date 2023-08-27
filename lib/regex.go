package lib

type NFA struct {
	States []*State
}

type State struct {
}

func Parse(s string) (*NFA, error) {
	return new(NFA), nil // everything compiles! gratz
}
