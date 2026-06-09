package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/jettero/pcrec"
	"github.com/spf13/pflag"
)

func ProcessArgs() []string {
	var halp *bool = pflag.BoolP("help", "h", false, "show the help screen text")

	pflag.Parse()

	if *halp {
		b := bytes.NewBufferString("\nUSAGE: pcrec-expr [--options] [pattern [pattern …]]\n")
		pflag.CommandLine.SetOutput(b)
		pflag.PrintDefaults()
		fmt.Println(b.String())
		os.Exit(0)
	}

	return pflag.Args()
}

func main() {
	for _, arg := range ProcessArgs() {
		re, err := pcrec.Parse(arg)
		fmt.Printf("RE-ARG: %s\n", arg)
		if err != nil {
			fmt.Printf("%+v\n", err)
		} else {
			fmt.Println(re.Describe(1))
		}
		fmt.Println()
	}
}
