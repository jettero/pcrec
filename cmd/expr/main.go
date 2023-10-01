package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/jettero/pcrec"
	"github.com/jettero/pcrec/lib"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/pflag"
)

func ProcessArgs() []string {
	var trace *bool = pflag.BoolP("verbose", "v", false, "print verbose debug messages during the RE parse")
	var ppp *bool = pflag.BoolP("pp", "p", false, "print the analysis with k0kubun/pp rather than the internal formatter")
	var halp *bool = pflag.BoolP("help", "h", false, "show the help screen text")

	pflag.Parse()

	if *halp {
		b := bytes.NewBufferString("\nUSAGE: pcrec-expr [--options] [pattern [pattern …]]\n")
		pflag.CommandLine.SetOutput(b)
		pflag.PrintDefaults()
		fmt.Println(b.String())
		os.Exit(0)
	}

	if *trace {
		os.Setenv("PCREC_TRACE", "yes")
	}

	if *ppp {
		os.Setenv("PCREC_PP_NFA", "yes")
	}

	return pflag.Args()
}

func main() {
	for _, arg := range ProcessArgs() {
		nfa, err := pcrec.Parse(arg)

		fmt.Printf("\nNFA Description for \"%s\":\n", arg)
		if lib.TruthyEnv("PCREC_PP_NFA") {
			pp.Println(nfa)
		} else {
			fmt.Print(nfa.Describe(1))
		}

		if err != nil {
			fmt.Printf("RE-ARG: %s\n", arg)
			fmt.Printf("%+v\n", err)
		}
		fmt.Println()
	}
}
