package lib

import (
	"fmt"
	"strings"
)

//go:generate vartan-go --package lib ../vartan/regexp.json

// Parse compiles a PCRE-syntax pattern (currently a subset) into an *RE.
func Parse(pat string) (*RE, error) {
	toks, err := NewTokenStream(strings.NewReader(pat))
	if err != nil {
		return nil, fmt.Errorf("parse %q: tokenize: %w", pat, err)
	}
	gram := NewGrammar()
	tb := NewDefaultSyntaxTreeBuilder()
	p, err := NewParser(toks, gram, SemanticAction(NewASTActionSet(gram, tb)))
	if err != nil {
		return nil, fmt.Errorf("parse %q: new parser: %w", pat, err)
	}
	if err := p.Parse(); err != nil {
		return nil, fmt.Errorf("parse %q: %w", pat, err)
	}
	if errs := p.SyntaxErrors(); len(errs) > 0 {
		return nil, formatSyntaxErrors(pat, gram, errs)
	}
	tree := tb.Tree()
	if tree == nil {
		return nil, fmt.Errorf("parse %q: empty parse tree", pat)
	}
	op, n, err := compileRoot(tree)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", pat, err)
	}
	return &RE{Root: op, NumCaptures: n}, nil
}

func formatSyntaxErrors(pat string, gram Grammar, errs []*SyntaxError) error {
	var sb strings.Builder
	for i, e := range errs {
		if i > 0 {
			sb.WriteString("; ")
		}
		var lex string
		if e.Token != nil {
			lex = string(e.Token.Lexeme())
		}
		fmt.Fprintf(&sb, "syntax error at col %d: %s near %q", e.Col+1, e.Message, lex)
	}
	return fmt.Errorf("parse %q: %s", pat, sb.String())
}
