package lib

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const INDENT string = "  "

func PrintableizeRunes(rz []rune, max int) string {
	var ret []string
	for i, r := range rz {
		if max > 0 && i >= max {
			ret = append(ret, " …")
			break
		}
		ret = append(ret, Printableize(r, false))
	}
	return strings.Join(ret, "")
}

func Printableize(r rune, loudWhitespace bool) string {
	switch r {
	case '\t':
		return `\t`
	case '\r':
		return `\r`
	case '\n':
		return `\n`
	case '"':
		return "\\\""
	case ' ':
		if loudWhitespace {
			return `«space»`
		}
		return " "
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

func Qstr(min int, max int, greedy bool) string {
	if min != max {
		qstr := fmt.Sprintf("{%d,%s}", min, NegIsInfinite(max))
		if !greedy {
			qstr += "?"
		}
		return qstr
	}
	if min == 1 {
		return ""
	}
	return fmt.Sprintf("{%d}", min)
}

func (g *Group) Describe(indent int) string {
	istr := strings.Repeat(INDENT, indent)
	jstr := istr + INDENT
	var ret []string
	for _, orItem := range g.States {
		var orStr []string
		for _, sItem := range orItem {
			orStr = append(orStr, sItem.Describe(indent+1))
		}
		ret = append(ret, strings.Join(orStr, "\n"))
	}
	qstr := Qstr(g.Min, g.Max, g.Greedy)
	if g.Capture {
		qstr += ":cap"
	}
	return fmt.Sprintf("%sG%s => {\n%s\n%s}",
		istr, qstr,
		strings.Join(ret, fmt.Sprintf("\n%s|\n", jstr)),
		istr)
}

func (r *RE) Describe(indent int) string {
	var ret []string
	for _, s := range r.States {
		ret = append(ret, s.Describe(indent+1))
	}
	return fmt.Sprintf("RE:\n%s", strings.Join(ret, "\n"))
}

func (m *Matcher) Describe() string {
	if m.Any {
		if m.Inverse {
			// the parser shouldn't actually produce this ... right?
			return "[«nil»]"
		}
		return "[«any»]"
	}
	var ret string = "["
	if m.Inverse {
		ret += "^"
	}
	ret += Printableize(m.First, true)
	if m.First < m.Last {
		ret += "-" + Printableize(m.Last, true)
	}
	ret += "]"
	return ret
}

func (r *REsult) Describe(indent int) string {
	istr := strings.Repeat(INDENT, indent)
	jstr := istr + INDENT

	ret := fmt.Sprintf("%sREsult[matched:%v]", istr, r.Matched)
	if len(r.Groups) > 0 {
		for i, item := range r.Groups {
			var sitem string = "nil"
			if item != nil {
				sitem = fmt.Sprintf("\"%s\"", string(*item))
			}
			ret += fmt.Sprintf("\n%sGroup(%d): %v", jstr, i+1, sitem)
		}
	}

	return ret + "\n"
}

func (s *State) short() string {
	return GetTag(s)
}

func (g *Group) medium() string {
	return fmt.Sprintf("%s: %s", GetTag(g), g.short())
}

func (s *State) medium() string {
	ret := []string{}
	for _, m := range s.Match {
		ret = append(ret, m.Describe())
	}
	var junk string
	if s.And {
		junk = " && "
	} else {
		junk = " || "
	}
	sstr := strings.Join(ret, junk)
	if len(ret) > 1 {
		sstr = fmt.Sprintf("(?:%s)", sstr)
	}
	qstr := Qstr(s.Min, s.Max, s.Greedy)
	return fmt.Sprintf("%s: %s%s", GetTag(s), sstr, qstr)
}

func (s *State) Describe(indent int) string {
	istr := strings.Repeat(INDENT, indent)
	ret := []string{}
	for _, m := range s.Match {
		ret = append(ret, m.Describe())
	}
	var junk string
	if s.And {
		junk = " && "
	} else {
		junk = " || "
	}
	sstr := strings.Join(ret, junk)
	qstr := Qstr(s.Min, s.Max, s.Greedy)
	return fmt.Sprintf("%sS%s => %s", istr, qstr, sstr)
}

func (g *Group) short() string {
	var istr []string
	for _, items := range g.States {
		var iistr []string
		for _, item := range items {
			iistr = append(iistr, item.short())
		}
		istr = append(istr, strings.Join(iistr, "."))
	}
	flags := ""
	if !g.Capture {
		flags += "?:"
	}
	return fmt.Sprintf("(%s%s)%s", flags, strings.Join(uniqueStrings(istr), "|"), Qstr(g.Min, g.Max, g.Greedy))
}

func (n *NFA) asDotNodes(oo *numberedItems) (ret []string) {
	if !oo.onlyOnce(n) {
		return
	}
	nt := GetTag(n)
	ret = append(ret, fmt.Sprintf("%s [label=\"%s\"]", nt, n.Whence.medium()))
	for _, ni := range n.children {
		if ni == nil {
			continue
		}
		for _, line := range ni.asDotNodes(oo) {
			ret = append(ret, line)
		}
	}
	for _, nfaSlice := range n.Transitions {
		for _, ni := range nfaSlice {
			if ni == nil {
				continue
			}
			for _, line := range ni.asDotNodes(oo) {
				ret = append(ret, line)
			}
		}
	}
	return
}

func (n *NFA) asDotTransitions(oo *numberedItems) (ret []string) {
	nt := GetTag(n)
	if !oo.onlyOnce(n) {
		return
	}
	for s, nfaSlice := range n.Transitions {
		if s == nil {
			for _, nfa := range nfaSlice {
				ret = append(ret, fmt.Sprintf("%s -> %s [label=\"ε\"]", nt, GetTag(nfa)))
				for _, line := range nfa.asDotTransitions(oo) {
					ret = append(ret, line)
				}
			}
			continue
		}
		for _, m := range s.Match {
			// XXX: this is slightly spurious since it's going to be wrong
			// for the case of [\D\W] or similar.
			if s.Max != 0 {
				for _, nfa := range nfaSlice {
					if nfa == nil {
						ret = append(ret, fmt.Sprintf("%s -> F [label=\"%s\"]",
							nt, m.Describe()))
					} else {
						ret = append(ret, fmt.Sprintf("%s -> %s [label=\"%s\"]", nt, GetTag(nfa), m.Describe()))
						for _, line := range nfa.asDotTransitions(oo) {
							ret = append(ret, line)
						}
					}
				}
			}
		}
	}
	return
}

func (n *NFA) AsDot() string {
	lines := []string{"digraph G {"}
	t := n.asDotNodes(makeNumberedItems())
	sort.Strings(t)
	for _, i := range t {
		lines = append(lines, i)
	}

	lines = append(lines, "F [label=\"F: Accept\"]")
	lines = append(lines, "")

	t = n.asDotTransitions(makeNumberedItems())
	sort.Strings(t)
	for _, i := range t {
		lines = append(lines, i)
	}

	return fmt.Sprintf("%s\n}", strings.Join(lines, "\n  "))
}
