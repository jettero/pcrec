package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	reLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"AnyChar", `.`},
		{"Quant", `[*?]`},
	})
	parser = participle.MustBuild[Thing](
		participle.Lexer(reLexer),
		participle.UseLookahead(2),
	)
)
