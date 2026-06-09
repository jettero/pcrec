package lib

import (
	"fmt"
	"strings"
)

// Op is one node in the compiled regex tree. Match consumes input from
// runes[pos:] and on each successful prefix calls tail(newPos). If any path
// through tail returns true, Match returns true; otherwise Match returns
// false (the caller backtracks). This is the standard CPS-style backtracking
// matcher that lets PCRE features (lookaround, atomic groups, callouts) drop
// in cleanly later.
type Op interface {
	match(m *matcher, pos int, tail func(int) bool) bool
	describe(indent int) string
}

func indentS(n int) string { return strings.Repeat("  ", n) }

// --- Concat: A then B then C ---

type opConcat struct{ ops []Op }

func (c *opConcat) match(m *matcher, pos int, tail func(int) bool) bool {
	var rec func(i, p int) bool
	rec = func(i, p int) bool {
		if i >= len(c.ops) {
			return tail(p)
		}
		return c.ops[i].match(m, p, func(np int) bool { return rec(i+1, np) })
	}
	return rec(0, pos)
}

func (c *opConcat) describe(indent int) string {
	if len(c.ops) == 1 {
		return c.ops[0].describe(indent)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%sConcat", indentS(indent))
	for _, op := range c.ops {
		b.WriteString("\n")
		b.WriteString(op.describe(indent + 1))
	}
	return b.String()
}

// --- Alt: A | B | C (leftmost-first like PCRE, not leftmost-longest) ---

type opAlt struct{ branches []Op }

func (a *opAlt) match(m *matcher, pos int, tail func(int) bool) bool {
	for _, op := range a.branches {
		if op.match(m, pos, tail) {
			return true
		}
	}
	return false
}

func (a *opAlt) describe(indent int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%sAlt", indentS(indent))
	for _, op := range a.branches {
		b.WriteString("\n")
		b.WriteString(op.describe(indent + 1))
	}
	return b.String()
}

// --- Lit: single literal rune ---

type opLit struct{ r rune }

func (l *opLit) match(m *matcher, pos int, tail func(int) bool) bool {
	if pos < len(m.input) && m.input[pos] == l.r {
		return tail(pos + 1)
	}
	return false
}

func (l *opLit) describe(indent int) string {
	return fmt.Sprintf("%sLit %q", indentS(indent), l.r)
}

// --- Any: . (matches any one rune; for now, including \n — flag-driven later) ---

type opAny struct{}

func (opAny) match(m *matcher, pos int, tail func(int) bool) bool {
	if pos < len(m.input) {
		return tail(pos + 1)
	}
	return false
}

func (opAny) describe(indent int) string { return indentS(indent) + "Any" }

// --- Class: [abc] / [a-z] / [^...] ---

type runeRange struct{ first, last rune }

type opClass struct {
	ranges []runeRange
	invert bool
}

func (c *opClass) match(m *matcher, pos int, tail func(int) bool) bool {
	if pos >= len(m.input) {
		return false
	}
	r := m.input[pos]
	hit := false
	for _, rr := range c.ranges {
		if rr.first <= r && r <= rr.last {
			hit = true
			break
		}
	}
	if hit == c.invert {
		return false
	}
	return tail(pos + 1)
}

func (c *opClass) describe(indent int) string {
	var b strings.Builder
	if c.invert {
		fmt.Fprintf(&b, "%sClass^ [", indentS(indent))
	} else {
		fmt.Fprintf(&b, "%sClass [", indentS(indent))
	}
	for i, rr := range c.ranges {
		if i > 0 {
			b.WriteString(",")
		}
		if rr.first == rr.last {
			fmt.Fprintf(&b, "%q", rr.first)
		} else {
			fmt.Fprintf(&b, "%q-%q", rr.first, rr.last)
		}
	}
	b.WriteString("]")
	return b.String()
}

// --- Group: (...) capturing ---

type opGroup struct {
	id    int
	inner Op
}

func (g *opGroup) match(m *matcher, pos int, tail func(int) bool) bool {
	saved := m.captures[g.id]
	m.captures[g.id].Start = pos
	if g.inner.match(m, pos, func(np int) bool {
		m.captures[g.id].End = np
		return tail(np)
	}) {
		return true
	}
	m.captures[g.id] = saved
	return false
}

func (g *opGroup) describe(indent int) string {
	return fmt.Sprintf("%sGroup #%d\n%s", indentS(indent), g.id, g.inner.describe(indent+1))
}

// --- Quantifier: greedy {min,max}. max < 0 means unbounded.
// Lazy variants land in a later phase. ---

type opGreedyRep struct {
	inner    Op
	min, max int
}

func (g *opGreedyRep) match(m *matcher, pos int, tail func(int) bool) bool {
	var rec func(count, p int) bool
	rec = func(count, p int) bool {
		if g.max < 0 || count < g.max {
			if g.inner.match(m, p, func(np int) bool {
				if np == p {
					// zero-width body would loop forever — bail to tail if we've met min
					if count+1 >= g.min {
						return tail(np)
					}
					return false
				}
				return rec(count+1, np)
			}) {
				return true
			}
		}
		if count >= g.min {
			return tail(p)
		}
		return false
	}
	return rec(0, pos)
}

func (g *opGreedyRep) describe(indent int) string {
	mx := "∞"
	if g.max >= 0 {
		mx = fmt.Sprintf("%d", g.max)
	}
	return fmt.Sprintf("%sGreedyRep{%d,%s}\n%s", indentS(indent), g.min, mx, g.inner.describe(indent+1))
}
