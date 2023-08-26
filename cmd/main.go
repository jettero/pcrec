package main

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/jettero/pcrec/lib"
	"github.com/k0kubun/pp/v3"
	"os"
	"strconv"
	"strings"
)

func truthy(s string) bool {
	s = strings.ToLower(s)
	if s == "true" || s == "yes" {
		return true
	}
	if num, err := strconv.Atoi(s); err == nil {
		return num != 0
	}
	return false
}

func main() {
	fmt.Printf("ENBF:\n%s\n\n", lib.Parser)

	for _, arg := range os.Args[1:] {
		fmt.Printf("-------------=: parsing, \"%s\"\n", arg)
		var trace strings.Builder
		ast, err := lib.Parser.ParseString("-", arg, participle.Trace(&trace))
		if truthy(os.Getenv("PCREC_TRACE")) {
			fmt.Printf("TRACE:\n%s", trace.String())
		}
		if err != nil {
			if truthy(os.Getenv("PCREC_TRACE_ERROR")) {
				fmt.Printf("TRACE:\n%s", trace.String())
			}
			fmt.Printf("RE-ARG: %s\n", arg)
			fmt.Printf("ERROR:  %+v\n", err)
		} else {
			pp.Print(ast) // why doesn't this print its own newlin? pfft
			fmt.Println()
		}
		fmt.Println()
	}
}
