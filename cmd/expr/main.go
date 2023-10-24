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
	var dot *bool = pflag.BoolP("dot", "D", false, "output graphviz dot format")
	var halp *bool = pflag.BoolP("help", "h", false, "show the help screen text")

	pflag.Parse()

	if *halp {
		b := bytes.NewBufferString("\nUSAGE: pcrec-expr [--options] [pattern [pattern …]]\n")
		pflag.CommandLine.SetOutput(b)
		pflag.PrintDefaults()
		fmt.Println(b.String())
		os.Exit(0)
	}

	if *dot {
		os.Setenv("PCREC_GV_DOT", "yes")
	}

	if *trace {
		os.Setenv("PCREC_TRACE", "yes")
	}

	if *ppp {
		os.Setenv("PCREC_PP_RE", "yes")
	}

	return pflag.Args()
}

func main() {
	for _, arg := range ProcessArgs() {
		re, err := pcrec.Parse(arg)

		if lib.TruthyEnv("PCREC_GV_DOT") {
			if err == nil {
				if nfa := lib.BuildNFA(re); nfa != nil {
					fmt.Println(nfa.AsDot())
				}
			} else {
				fmt.Println("digraph graphname {\n  /*")
				fmt.Println("  **", err)
				fmt.Println("  */\n}")
			}

		} else {
			if lib.TruthyEnv("PCREC_PP_RE") {
				pp.Println(re)
			} else {
				fmt.Print(re.Describe(1))
			}

			if err != nil {
				fmt.Printf("RE-ARG: %s\n", arg)
				fmt.Printf("%+v\n", err)
			}
			fmt.Println()
		}
	}
}
