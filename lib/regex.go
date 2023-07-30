package lib

type Regexp struct {
	pat string
	pos int
}

type MatchBytes struct {
	limit [2]byte
	chars [][2]byte
}
