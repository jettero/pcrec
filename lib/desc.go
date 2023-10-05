package lib

import (
	"fmt"
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
