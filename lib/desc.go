package lib

import (
	"fmt"
	"strings"
	"unicode"
)

const INDENT string = "  "

func Printableize(r rune) string {
	switch r {
	case '\t':
		return `\t`
	case '\r':
		return `\r`
	case '\n':
		return `\n`
	case ' ':
		return `«space»`
	}
	if unicode.IsPrint(r) {
		return fmt.Sprintf("%c", r)
	}
	ret := fmt.Sprintf("%02x", r)
	if len(ret) > 2 {
		return fmt.Sprintf("\\x{%s}", ret)
	}
	return fmt.Sprintf("\\x%s", ret)
}

func NegIsInfinite(i int) string {
	if i < 0 {
		return "∞"
	}
	return fmt.Sprintf("%d", i)
}

func (g *Group) Describe(indent int) string {
	istr := strings.Repeat(INDENT, indent)
	var ret []string
	for _, orItem := range g.States {
		var orStr []string
		for _, sItem := range orItem {
			orStr = append(orStr, sItem.Describe(indent+1))
		}
		ret = append(ret, strings.Join(orStr, "\n"))
	}
	ghead := fmt.Sprintf("%s<Group capture=%v greedy=%v min=%d max=%s>",
		istr, g.Capture, g.Greedy, g.Min, NegIsInfinite(g.Max))
	gstr := strings.Join(ret, fmt.Sprintf("\n%sor\n", istr))
	gfoot := fmt.Sprintf("%s</Group>", istr)
	return strings.Join([]string{ghead, gstr, gfoot}, "\n")
}

func (n *NFA) Describe(indent int) string {
	var ret []string
	for _, s := range n.States {
		ret = append(ret, s.Describe(indent+1))
	}
	return strings.Join(ret, "\n")
}

func (m *Matcher) Describe() string {
	if m.Any {
		if m.Inverse {
			// the parser shouldn't actually produce this ... right?
			return "M[«none»]"
		}
		return "M[«any»]"
	}
	var ret string = "M["
	if m.Inverse {
		ret += "^"
	}
	ret += Printableize(m.First)
	if m.First < m.Last {
		ret += "-" + Printableize(m.Last)
	}
	ret += "]"
	return ret
}

func (s *State) Describe(indent int) string {
	istr := strings.Repeat(INDENT, indent)
	ret := []string{}
	for _, m := range s.Match {
		ret = append(ret, m.Describe())
	}
	var junk string
	if s.And {
		junk = " and "
	} else {
		junk = " or "
	}
	shead := fmt.Sprintf("%s<State greedy=%v min=%d max=%s and=%v>",
		istr, s.Greedy, s.Min, NegIsInfinite(s.Max), s.And)
	sstr := strings.Join(ret, junk)
	sfoot := "</State>"
	return strings.Join([]string{shead, sstr, sfoot}, " ")
}
