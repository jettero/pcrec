package lib

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type NFA struct {
	States []*State `@@*`
}

type MinMax struct {
	min    int
	max    int
	greedy bool
}

func (mm *MinMax) Capture(values []string) error {
	v := []rune(values[0])

	switch v[0] {
	case '?':
		mm.min = 0
		mm.max = 1
	case '*':
		mm.min = 0
		mm.max = -1
	case '+':
		mm.min = 1
		mm.max = -1
	default:
		mm.min = 1 // this never gets reached when Quant doesn't match
		mm.max = 1
	}
	if len(v) > 1 && v[1] == '?' {
		mm.greedy = false
	} else {
		mm.greedy = true
	}
	return nil
}

type RuneRange struct {
	B *string `@Rune`
	E *string `( "-" @Rune )?`
}

type Matcher struct {
	Any      bool         `  @DotOp`
	Range    []*RuneRange `| "[" @@ @@* "]" | @@`
	Groupies []*State     `| "(" @@ ( "|" @@ )* ")"`
}

type State struct {
	Match []*Matcher `@@+`
	Quant *MinMax    `@QuantOp?`
}

var (
	Lexer = lexer.MustSimple([]lexer.SimpleRule{
		{"DotOp", `\.`},
		{"Digits", `\d+`},
		{"QuantOp", `[*+?]\??|{\d+(,\d*)?}|{,\d+}`},
		{"Rune", `[^[\]*+?\\()|]`},
		{"AnyRune", `.`},
	})
	Parser = participle.MustBuild[NFA](
		participle.Lexer(Lexer),
		participle.UseLookahead(99999),
	)
)
