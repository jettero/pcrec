package pcre2dsl

import (
	"os"
	"strings"
	"testing"
)

func parseFile(path string) ([]Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseOutput(path, f)
}

const sample = `# header comment
#forbid_utf

/the quick brown fox/
    the quick brown fox
 0: the quick brown fox
\= Expect no match
    The quick brown FOX
No match

/(a)(b)/i
    xaby
 0: ab
 1: a
 2: b

/abcd\t\n/
    abcd\t\n
 0: abcd\x09\x0a
`

func TestParseOutput_Sample(t *testing.T) {
	pats, err := ParseOutput("sample", strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	if len(pats) != 3 {
		t.Fatalf("got %d patterns, want 3: %+v", len(pats), pats)
	}

	p0 := pats[0]
	if p0.Pattern != "the quick brown fox" || p0.Modifiers != "" {
		t.Errorf("p0 pattern/mods wrong: %q %q", p0.Pattern, p0.Modifiers)
	}
	if len(p0.Subjects) != 2 {
		t.Fatalf("p0 subjects = %d, want 2: %+v", len(p0.Subjects), p0.Subjects)
	}
	if p0.Subjects[0].Raw != "the quick brown fox" {
		t.Errorf("p0.s0 raw = %q", p0.Subjects[0].Raw)
	}
	if !p0.Subjects[0].Expect.Matched {
		t.Errorf("p0.s0 should have matched")
	}
	if g := p0.Subjects[0].Expect.Groups; len(g) != 1 || g[0] == nil || *g[0] != "the quick brown fox" {
		t.Errorf("p0.s0 group0 wrong: %v", g)
	}
	if p0.Subjects[1].Expect.Matched {
		t.Errorf("p0.s1 should NOT have matched")
	}

	p1 := pats[1]
	if p1.Modifiers != "i" {
		t.Errorf("p1.modifiers = %q, want i", p1.Modifiers)
	}
	if len(p1.Subjects) != 1 {
		t.Fatalf("p1 subjects = %d, want 1", len(p1.Subjects))
	}
	g := p1.Subjects[0].Expect.Groups
	if len(g) != 3 {
		t.Fatalf("p1.s0 groups = %d, want 3: %v", len(g), g)
	}
	if g[0] == nil || *g[0] != "ab" {
		t.Errorf("p1.s0 g0 = %v, want ab", deref(g[0]))
	}
	if g[1] == nil || *g[1] != "a" {
		t.Errorf("p1.s0 g1 = %v, want a", deref(g[1]))
	}
	if g[2] == nil || *g[2] != "b" {
		t.Errorf("p1.s0 g2 = %v, want b", deref(g[2]))
	}
}

func TestDecodeSubject(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"hello", "hello", true},
		{"a\\tb", "a\tb", true},
		{"a\\nb", "a\nb", true},
		{"a\\x41b", "aAb", true},
		{"a\\\\b", "a\\b", true},
		{"a\\$b", "a$b", true},
		{"a\\x{41}b", "", false}, // unsupported brace form
		{"a\\071", "", false},    // unsupported octal
	}
	for _, c := range cases {
		got, ok := DecodeSubject(c.in)
		if ok != c.ok {
			t.Errorf("DecodeSubject(%q) ok=%v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && got != c.want {
			t.Errorf("DecodeSubject(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDecodeOutputText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abcd\\x09\\x0a", "abcd\t\n"},
		{"plain", "plain"},
		{"slash\\\\back", "slash\\back"},
	}
	for _, c := range cases {
		got, ok := DecodeOutputText(c.in)
		if !ok {
			t.Errorf("DecodeOutputText(%q) failed", c.in)
			continue
		}
		if got != c.want {
			t.Errorf("DecodeOutputText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParsePatternLine_EscapedSlash(t *testing.T) {
	pat, mods, ok := parsePatternLine(`/a\/b/i`)
	if !ok {
		t.Fatal("expected ok")
	}
	if pat != `a\/b` || mods != "i" {
		t.Errorf("got pat=%q mods=%q", pat, mods)
	}
}

func deref(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

// TestParseRealTestoutput1 is a smoke test against the vendored upstream
// corpus. Skips cleanly if the file isn't where it should be (e.g., running
// the package tests from an extracted tree without contrib/).
func TestParseRealTestoutput1(t *testing.T) {
	const path = "../../contrib/pcre2/testdata/testoutput1"
	pats, err := parseFile(path)
	if err != nil {
		t.Skipf("skipping integration smoke (no corpus at %s): %v", path, err)
	}
	if len(pats) < 100 {
		t.Errorf("only got %d patterns from testoutput1; expected >100", len(pats))
	}
	var withSubj, withGroups int
	for _, p := range pats {
		if len(p.Subjects) > 0 {
			withSubj++
		}
		for _, s := range p.Subjects {
			if len(s.Expect.Groups) > 1 {
				withGroups++
			}
		}
	}
	t.Logf("parsed %d patterns; %d have subjects; %d subjects had >1 group",
		len(pats), withSubj, withGroups)
}
