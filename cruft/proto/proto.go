package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/k0kubun/pp/v3"
)

type RE struct {
	States []*State `@@*`
}

type MinMax struct {
	min int
	max int
}

func (mm *MinMax) Capture(values []string) error {
	switch values[0] {
	case "?":
		mm.min = 0
		mm.max = 1
	case "*":
		mm.min = 0
		mm.max = -1
	case "+":
		mm.min = 1
		mm.max = -1
	default:
		mm.min = 1 // this never gets reached when Quant doesn't match
		mm.max = 1
	}
	return nil
}

type RuneRange struct {
	B *string `@Rune`
	E *string `( "-" @Rune )?`
}

type Matcher struct {
	Any   bool         `  @DotOp`
	Range []*RuneRange `| "[" @@ @@* "]" | @@`
}

type State struct {
	Match *Matcher `@@`
	Quant *MinMax  `@QuantOp?`
}

var (
	reLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"DotOp", `\.`},
		{"QuantOp", `[*+?]`},
		{"Rune", `[^[\]*+?\\]`},
		{"AnyRune", `.`},
	})
	parser = participle.MustBuild[RE](
		participle.Lexer(reLexer),
		participle.UseLookahead(99999),
	)
)

func main() {
	for _, arg := range os.Args[1:] {
		fmt.Printf("-------------=: parsing, \"%s\"\n", arg)
		ast, err := parser.ParseString("-", arg)
		if err != nil {
			fmt.Printf("ERROR: %+v\n", err)
		} else {
			pp.Print(ast) // why doesn't this print its own newlin? pfft
			fmt.Println()
		}
		fmt.Println()
	}
}
