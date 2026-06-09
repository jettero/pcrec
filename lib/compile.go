package lib

import (
	"fmt"
	"strconv"
	"strings"
)

// compileRoot walks a vartan-built *Node AST and produces an Op tree plus the
// number of capture groups encountered. The KindName values come from the
// non-terminal / terminal names in vartan/regexp.vartan.
func compileRoot(root *Node) (Op, int, error) {
	c := &compiler{}
	defer func() {
		if r := recover(); r != nil {
			// fall through; outer error is returned by closure capture below
		}
	}()
	var op Op
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("compile: %v", r)
			}
		}()
		op = c.node(root)
	}()
	return op, c.numGroups, err
}

type compiler struct {
	numGroups int
}

func (c *compiler) node(n *Node) Op {
	if n == nil {
		panic("nil node")
	}
	if n.Type == NodeTypeError {
		panic("error node in AST: " + n.KindName)
	}
	switch n.KindName {
	case "re":
		// alternation at the top level — one or more concat children
		return c.choice(n.Children, &opAlt{})
	case "concat":
		// concatenation — one or more atom children
		return c.choice(n.Children, &opConcat{})
	case "atom":
		return c.atom(n)
	case "a_grp":
		return c.group(n)
	case "a_set":
		return c.set(n)
	case "a_any":
		return opAny{}
	case "a_char":
		return &opLit{r: oneRune(n.Text)}
	}
	panic("unhandled kind: " + n.KindName)
}

// choice compiles a list of child nodes into either a single Op (if the list
// has one element — no wrapping needed) or wraps them in the given
// multi-child operator (opAlt for re, opConcat for concat).
func (c *compiler) choice(children []*Node, wrap Op) Op {
	if len(children) == 0 {
		return &opConcat{}
	}
	if len(children) == 1 {
		return c.node(children[0])
	}
	ops := make([]Op, len(children))
	for i, ch := range children {
		ops[i] = c.node(ch)
	}
	switch w := wrap.(type) {
	case *opAlt:
		w.branches = ops
		return w
	case *opConcat:
		w.ops = ops
		return w
	}
	panic("unsupported wrap type")
}

func (c *compiler) group(n *Node) Op {
	c.numGroups++
	id := c.numGroups
	// a_grp's children come from `#ast re...` — the inner re's children
	// inlined. Wrap them as an alt (matching the re semantics).
	inner := c.choice(n.Children, &opAlt{})
	return &opGroup{id: id, inner: inner}
}

// atom handles both unquantified atoms (single child) and quantified atoms
// (operand followed by one or more quantifier-indicator terminals).
//
// Per prod 8 (atom → atom qty, #ast atom... qty...) the AST flattens both
// sides, so a quantified atom's children look like [operand, q_min?, q_max?,
// q_01?, q_0m?, q_1m?] — the operand first, then 1-2 quantifier markers.
func (c *compiler) atom(n *Node) Op {
	if len(n.Children) == 0 {
		panic("empty atom")
	}
	operand := c.node(n.Children[0])
	rest := n.Children[1:]
	if len(rest) == 0 {
		return operand
	}
	min, max := parseQuantifier(rest)
	return &opGreedyRep{inner: operand, min: min, max: max}
}

// parseQuantifier turns a slice of q_* terminal nodes into (min, max).
// max < 0 means unbounded.
//
// NOTE — grammar ambiguity: `{n}` and `{n,}` both flatten to a single q_min
// child (see prods 25, 26 in vartan/regexp.vartan). For now we treat a lone
// q_min as `{n}`. Fixing that means amending the grammar so prod 26 preserves
// q_cnd in its #ast output — left for the phase that touches the grammar.
func parseQuantifier(markers []*Node) (int, int) {
	if len(markers) == 1 {
		m := markers[0]
		switch m.KindName {
		case "q_01":
			return 0, 1
		case "q_0m":
			return 0, -1
		case "q_1m":
			return 1, -1
		case "q_min":
			n := mustAtoi(m.Text)
			return n, n
		case "q_max":
			return 0, mustAtoi(m.Text)
		}
	}
	if len(markers) == 2 && markers[0].KindName == "q_min" && markers[1].KindName == "q_max" {
		return mustAtoi(markers[0].Text), mustAtoi(markers[1].Text)
	}
	panic("unrecognized quantifier shape: " + kindList(markers))
}

// set turns an a_set AST into an opClass. The a_set node's children are flat
// per prods 13-16 — possibly leading s_neg, then a sequence of s_min (single
// chars) and s_min/s_max pairs (ranges), possibly trailing s_cnd ("-]" form
// meaning the class includes a literal `-`).
func (c *compiler) set(n *Node) Op {
	var ranges []runeRange
	invert := false
	children := n.Children
	i := 0
	if i < len(children) && children[i].KindName == "s_neg" {
		invert = true
		i++
	}
	for i < len(children) {
		ch := children[i]
		if ch.KindName == "s_cnd" {
			ranges = append(ranges, runeRange{first: '-', last: '-'})
			i++
			continue
		}
		if ch.KindName != "s_min" {
			panic("unexpected child in a_set: " + ch.KindName)
		}
		r := oneRune(ch.Text)
		if i+1 < len(children) && children[i+1].KindName == "s_max" {
			ranges = append(ranges, runeRange{first: r, last: oneRune(children[i+1].Text)})
			i += 2
		} else {
			ranges = append(ranges, runeRange{first: r, last: r})
			i++
		}
	}
	return &opClass{ranges: ranges, invert: invert}
}

func oneRune(s string) rune {
	rs := []rune(s)
	if len(rs) != 1 {
		panic(fmt.Sprintf("expected single rune, got %q", s))
	}
	return rs[0]
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("not an int: %q", s))
	}
	return n
}

func kindList(ns []*Node) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = n.KindName
	}
	return strings.Join(parts, ",")
}
