package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jettero/pcrec"
	"github.com/jettero/pcrec/lib"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/pflag"
	"io"
	"os"
)

func ProcessArgs() []string {
	var trace *bool = pflag.BoolP("verbose", "v", false, "print verbose debug messages during the RE parse")
	var ppres *bool = pflag.BoolP("debug", "d", false, "print match result objects for each match (only really useful for debugging)")
	var halp *bool = pflag.BoolP("help", "h", false, "show the help screen text")

	pflag.Parse()

	if *halp {
		b := bytes.NewBufferString("\nUSAGE: pcrec-grep [--options] pattern [filename …]]\n")
		pflag.CommandLine.SetOutput(b)
		pflag.PrintDefaults()
		fmt.Println(b.String())
		os.Exit(0)
	}

	if *ppres {
		os.Setenv("PCREC_PP_RES", "yes")
	}

	if *trace {
		os.Setenv("PCREC_TRACE", "yes")
	}

	return pflag.Args()
}

func main() {
	args := ProcessArgs()
	pat := args[0]
	args = args[1:]

	if len(args) < 1 {
		args = []string{"-"}
	}

	nfa, err := pcrec.Parse(pat)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	if lib.TruthyEnv("PCREC_TRACE") {
		fmt.Print("---=: NFA:\n", nfa.Describe(1), "\n")
	}

	for _, fname := range args {
		var fh *bufio.Reader
		if fname == "-" {
			fh = bufio.NewReader(os.Stdin)
		} else {
			fh_, err := os.Open(fname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
				return
			}
			fh = bufio.NewReader(fh_)
		}
		for {
			line, err := fh.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "Error reading: %v\n", err)
				return
			}
			if lib.TruthyEnv("PCREC_PP_RES") {
				fmt.Print("---=: line: ", line)
				res := nfa.Search(line)
				pp.Println(res)
				fmt.Println("")
			} else {
				if res := nfa.Search(line); res.Matched {
					fmt.Print(line)
				}
			}
		}
	}
}
