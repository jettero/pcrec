package main

import (
	"fmt"
	"github.com/jettero/pcrec"
	"github.com/k0kubun/pp/v3"
)

func main() {
	for _, arg := range os.Args[1:] {
		fmt.Printf("-------------=: parsing, \"%s\"\n", arg)
		var trace strings.Builder
		nfa, err := pcrec.Parse(arg)
		if err != nil {
			fmt.Printf("RE-ARG: %s\n", arg)
			fmt.Printf("ERROR:  %+v\n", err)
		} else {
			pp.Print(nfa) // why doesn't this print its own newlin? pfft
			fmt.Println()
		}
		fmt.Println()
	}
}
