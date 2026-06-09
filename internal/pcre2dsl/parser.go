// Package pcre2dsl parses the pcre2test corpus format (the plain-text DSL
// the upstream pcre2 test harness reads). The parser intentionally targets
// `testoutputN` files, because those are a strict superset of the matching
// `testinputN` files — every input line is reproduced verbatim with the
// expected result lines interleaved. So one parser yields both patterns and
// expectations.
//
// Scope is deliberately narrow: enough to drive a conformance run of
// testinput1 against pcrec. Exotic constructs (\x{…}, \NNN, replication
// shorthand, # save/load, alternate pattern delimiters, multi-line patterns)
// are NOT supported here — the parser flags those cases as skippable so the
// driver can bucket them appropriately rather than silently mis-running.
package pcre2dsl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Pattern is one /…/modifiers block with all its subjects+expectations.
type Pattern struct {
	File      string
	LineNo    int    // 1-based, of the line that opens the pattern
	Pattern   string // regex body, sans delimiters
	Modifiers string // raw modifier suffix (e.g. "i", "im", "jit=3,no_utf_check")
	Subjects  []Subject
	// SkipReason is non-empty if the parser couldn't fully understand the
	// pattern block (e.g., alternate delimiter, multi-line pattern). Driver
	// should bucket such patterns as SKIP_UNPARSEABLE.
	SkipReason string
}

// Subject is one input string and its expected result.
type Subject struct {
	LineNo int
	Raw    string // subject as it appears, sans leading indent
	Expect Expect
}

// Expect is the structural expectation derived from the testoutput result lines.
type Expect struct {
	Matched bool
	// Groups[0] is group 0 (the whole match). Groups[i] for i>=1 is capture
	// group i. A missing/unset group is represented by an empty *string (nil).
	Groups []*string
}

// ParseOutput consumes a testoutputN reader and returns the patterns. Lines
// it cannot classify are silently skipped — driving the engine off of what
// IS understood is the priority; what isn't understood becomes SKIP_*.
func ParseOutput(file string, r io.Reader) ([]Pattern, error) {
	scn := bufio.NewScanner(r)
	scn.Buffer(make([]byte, 1<<20), 1<<22)

	var out []Pattern
	var cur *Pattern
	var curSub *Subject // last subject whose result lines we're still accumulating
	lineNo := 0

	for scn.Scan() {
		lineNo++
		line := scn.Text()

		if isBlank(line) || isComment(line) || isDirective(line) {
			curSub = nil // result block (if any) ends here
			continue
		}
		if isExpectHint(line) {
			// \= lines apply to following subjects but for *structural*
			// comparison the testoutput's actual result lines carry the
			// answer — so these need no semantic handling.
			curSub = nil
			continue
		}
		if r, ok := parseResultLine(line); ok {
			if curSub != nil {
				applyResult(&curSub.Expect, r)
			}
			// orphan result lines (no preceding subject) are dropped
			continue
		}
		if pat, mods, ok := parsePatternLine(line); ok {
			// flush previous pattern
			if cur != nil {
				out = append(out, *cur)
			}
			cur = &Pattern{
				File:      file,
				LineNo:    lineNo,
				Pattern:   pat,
				Modifiers: mods,
			}
			curSub = nil
			continue
		}
		if isMaybePatternOpener(line) && cur != nil {
			// pattern that doesn't close on its first line — flag the
			// outgoing one as multi-line if we just opened it, then move on
			cur.SkipReason = "multi-line or non-/ delimiter pattern"
			out = append(out, *cur)
			cur = nil
			curSub = nil
			continue
		}
		if isSubjectLine(line) && cur != nil {
			sub := Subject{LineNo: lineNo, Raw: strings.TrimLeft(line, " \t")}
			cur.Subjects = append(cur.Subjects, sub)
			curSub = &cur.Subjects[len(cur.Subjects)-1]
			continue
		}
		// unrecognized — ignore
		_ = curSub
	}
	if cur != nil {
		out = append(out, *cur)
	}
	if err := scn.Err(); err != nil {
		return out, fmt.Errorf("scan %s: %w", file, err)
	}
	return out, nil
}

func isBlank(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return false
		}
	}
	return true
}

func isComment(s string) bool {
	// Plain # at column 0, NOT a directive line like "#forbid_utf".
	// Directives are also '#'-prefixed; we treat both the same way (skip).
	return strings.HasPrefix(s, "# ") || s == "#"
}

func isDirective(s string) bool {
	if !strings.HasPrefix(s, "#") || strings.HasPrefix(s, "# ") {
		return false
	}
	// "#forbid_utf", "#newline_default lf any anycrlf", "#if !ebcdic",
	// "#endif", "#perltest", "#pattern …", "#subject …"
	return true
}

func isExpectHint(s string) bool {
	return strings.HasPrefix(strings.TrimLeft(s, " \t"), `\=`)
}

// parsePatternLine handles the common case: a line that starts with '/' and
// closes with '/' followed by zero or more modifier characters. Anything more
// exotic returns ok=false.
func parsePatternLine(line string) (pat, mods string, ok bool) {
	if len(line) == 0 || line[0] != '/' {
		return "", "", false
	}
	// Find the unescaped closing '/'. PCRE patterns can contain '\/' which we
	// must treat as a literal slash inside the pattern, not the terminator.
	i := 1
	for i < len(line) {
		if line[i] == '\\' && i+1 < len(line) {
			i += 2
			continue
		}
		if line[i] == '/' {
			break
		}
		i++
	}
	if i >= len(line) {
		// pattern doesn't close on this line — caller will mark multi-line
		return "", "", false
	}
	return line[1:i], line[i+1:], true
}

func isMaybePatternOpener(line string) bool {
	return len(line) > 0 && line[0] == '/'
}

func isSubjectLine(s string) bool {
	// Subjects are indented at least 4 spaces in the upstream files.
	// The tab variant is also valid.
	if len(s) < 1 {
		return false
	}
	if !(s[0] == ' ' || s[0] == '\t') {
		return false
	}
	t := strings.TrimLeft(s, " \t")
	if t == "" {
		return false
	}
	// Distinguish from result lines, which begin with " N:" or "No match"
	// but those have already been matched earlier in the dispatch chain.
	return true
}

// resultLine is the parsed form of one ` N: text` or `No match` line.
type resultLine struct {
	noMatch bool
	group   int    // 0..N
	text    string // raw, still escaped (e.g., "\x09")
}

// parseResultLine handles:
//
//	"No match"
//	" 0: <text>"
//	"10: <text>"
//
// pcre2test emits other line shapes too (callout traces, named-group dumps,
// PCRE2_INFO output) that we don't try to interpret here — they show up only
// for patterns with extra modifiers which day-one driver will already SKIP.
func parseResultLine(line string) (resultLine, bool) {
	if line == "No match" {
		return resultLine{noMatch: true}, true
	}
	// look for "NN: " — number is right-padded in a 2-char field for N<10.
	t := strings.TrimLeft(line, " \t")
	if t == line && !strings.HasPrefix(line, " ") {
		// no leading whitespace AND doesn't fit other patterns
		return resultLine{}, false
	}
	colon := strings.Index(t, ":")
	if colon <= 0 {
		return resultLine{}, false
	}
	num := t[:colon]
	if !allDigits(num) {
		return resultLine{}, false
	}
	n := 0
	for _, c := range num {
		n = n*10 + int(c-'0')
	}
	rest := t[colon+1:]
	rest = strings.TrimPrefix(rest, " ")
	return resultLine{group: n, text: rest}, true
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func applyResult(e *Expect, r resultLine) {
	if r.noMatch {
		e.Matched = false
		return
	}
	e.Matched = true
	for len(e.Groups) <= r.group {
		e.Groups = append(e.Groups, nil)
	}
	t := r.text
	e.Groups[r.group] = &t
}

// DecodeSubject interprets the standard pcre2test escapes used in subject
// lines: \a \b \e \f \n \r \t \0 \\ and \xHH. Anything else (\NNN, \x{…},
// \[abc]{N}, \cX, \N{…}) returns ok=false — caller should SKIP that subject.
func DecodeSubject(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		if i+1 >= len(s) {
			return "", false
		}
		nc := s[i+1]
		switch nc {
		case 'a':
			b.WriteByte(0x07)
		case 'b':
			b.WriteByte(0x08)
		case 'e':
			b.WriteByte(0x1b)
		case 'f':
			b.WriteByte(0x0c)
		case 'n':
			b.WriteByte(0x0a)
		case 'r':
			b.WriteByte(0x0d)
		case 't':
			b.WriteByte(0x09)
		case '0':
			// Bare \0 is NUL; \0NN or \0NNN is octal (unsupported here).
			if i+2 < len(s) && s[i+2] >= '0' && s[i+2] <= '9' {
				return "", false
			}
			b.WriteByte(0x00)
		case '\\':
			b.WriteByte('\\')
		case 'x':
			// \xHH (two hex digits). \x{…} not supported.
			if i+3 >= len(s) || s[i+2] == '{' {
				return "", false
			}
			h1, ok1 := hexDigit(s[i+2])
			h2, ok2 := hexDigit(s[i+3])
			if !ok1 || !ok2 {
				return "", false
			}
			b.WriteByte(byte(h1*16 + h2))
			i += 2
		case '$', '?', '"', '\'', '/', '.':
			// \$ and friends: per pcre2test docs, a backslash before a
			// non-special character is the literal character.
			b.WriteByte(nc)
		default:
			// \NNN (octal), \cX, \N{…}, \[abc]{N}, etc. — not handled here.
			return "", false
		}
		i++
	}
	return b.String(), true
}

// DecodeOutputText is the inverse for the form pcre2test emits: it re-escapes
// non-printable bytes as \xHH. We decode those (plus \\) back to the raw form
// so it can be compared against the decoded subject. Anything else is taken
// as literal (the only escapes pcre2test emits in this slot are \xHH and \\).
func DecodeOutputText(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		if i+1 >= len(s) {
			return "", false
		}
		nc := s[i+1]
		switch nc {
		case '\\':
			b.WriteByte('\\')
			i++
		case 'x':
			if i+3 >= len(s) {
				return "", false
			}
			h1, ok1 := hexDigit(s[i+2])
			h2, ok2 := hexDigit(s[i+3])
			if !ok1 || !ok2 {
				return "", false
			}
			b.WriteByte(byte(h1*16 + h2))
			i += 3
		default:
			// pcre2test shouldn't emit anything else, but be lenient
			b.WriteByte('\\')
			b.WriteByte(nc)
			i++
		}
	}
	return b.String(), true
}

func hexDigit(b byte) (int, bool) {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), true
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10, true
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10, true
	}
	return 0, false
}
