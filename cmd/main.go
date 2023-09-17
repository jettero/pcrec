package main

import (
	"fmt"
	"github.com/jettero/pcrec"
	"github.com/k0kubun/pp/v3"
	"os"
)

func main() {
	for _, arg := range os.Args[1:] {
		nfa, err := pcrec.ParseString(arg)
		pp.Print(nfa) // why doesn't this print its own newlin? pfft
		fmt.Println()
		if err != nil {
			fmt.Printf("RE-ARG: %s\n", arg)
			fmt.Printf("%+v\n", err)
		}
		fmt.Println()
	}
}
