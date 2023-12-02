package lib

type State struct {
	// Mostly just placeholders for connecting edges?
	// Surely we here track our descent though, right?
	//     Yeah, our counters
	//     and quantities go here for sure
}

type Transition struct {
	Inverse bool
	First   rune
	Last    rune
	// What do we consume?
	// Do we capture it?
	// Perhaps we here record what positions in the string we've eaten?
	//     How do we reverse it when we backtrack?
	//     Is this a pushdown stack of captures?
}
