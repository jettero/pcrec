package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/k0kubun/pp/v3"
)

type Things struct {
	Matchers []*Matcher `@@*`
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

type Matcher struct {
	AnyChar *string `@AnyChar`
	String  *string `| @String`
	Quant   *MinMax `@Quant?`
}

var (
	reLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"String", `([\w\d]+|\.)+`},
		{"AnyChar", `\.`},
		{"Quant", `[*+?]`},
	})
	parser = participle.MustBuild[Things](
		participle.Lexer(reLexer),
		participle.UseLookahead(2),
	)
)

func main() {
	for _, arg := range os.Args[1:] {
		reader := strings.NewReader(arg)
		fmt.Printf("-------------=: parsing, \"%s\"\n", arg)
		ast, err := parser.Parse("-", reader)
		if err != nil {
			fmt.Printf("ERROR: %+v\n", err)
		} else {
			pp.Print(ast)
		}
		fmt.Println()
	}
}
