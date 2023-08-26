package main

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/jettero/pcrec/lib"
	"github.com/k0kubun/pp/v3"
	"os"
	"strings"
)

func main() {
	fmt.Printf("ENBF:\n%s\n\n", lib.Parser)

	for _, arg := range os.Args[1:] {
		fmt.Printf("-------------=: parsing, \"%s\"\n", arg)
		var trace strings.Builder
		ast, err := lib.Parser.ParseString("-", arg, participle.Trace(&trace))
		if err != nil {
			fmt.Printf("TRACE:\n%s", trace.String())
			fmt.Printf("RE-ARG: %s\n", arg)
			fmt.Printf("ERROR:  %+v\n", err)
		} else {
			// fmt.Printf("TRACE:\n%s", trace.String())
			pp.Print(ast) // why doesn't this print its own newlin? pfft
			fmt.Println()
		}
		fmt.Println()
	}
}
