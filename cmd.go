package main

import (
	"fmt"
	"github.com/jettero/pcrec/lib"
	"github.com/k0kubun/pp/v3"
	"os"
)

func main() {
	for _, arg := range os.Args[1:] {
		fmt.Printf("-------------=: parsing, \"%s\"\n", arg)
		ast, err := lib.Parser.ParseString("-", arg)
		if err != nil {
			fmt.Printf("ERROR: %+v\n", err)
		} else {
			pp.Print(ast) // why doesn't this print its own newlin? pfft
			fmt.Println()
		}
		fmt.Println()
	}
}
