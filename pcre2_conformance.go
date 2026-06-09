// Package pcrec contains the project's public API. This file holds the
// internals of the PCRE2 conformance harness — kept out of *_test.go so the
// bless tool under cmd/pcre2-bless can share the same classification logic.
package pcrec

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jettero/pcrec/internal/pcre2dsl"
	"github.com/jettero/pcrec/lib"
)

// DefaultPerSubjectTimeout caps how long the engine may chew on a single
// (pattern, subject) before we declare it timed out. The pcrec engine is a
// backtracking matcher; pathological patterns × inputs can blow up
// exponentially. A wall-clock cap keeps the conformance run finite.
const DefaultPerSubjectTimeout = 500 * time.Millisecond

// Bucket classifies one (pattern, subject) outcome.
type Bucket int

const (
	BucketPass            Bucket = iota // engine result matches expected
	BucketSkipModifier                  // pattern has modifiers we don't claim to support yet
	BucketSkipUnparseable               // DSL couldn't fully parse this case (multi-line, alt delim, etc.)
	BucketSkipDecode                    // subject uses an escape form we haven't taught the decoder
	BucketParseError                    // engine refused the pattern
	BucketEnginePanic                   // engine panicked on this case
	BucketMatchMismatch                 // matched-vs-not-matched disagrees with expected
	BucketGroupMismatch                 // matched but capture groups disagree
	BucketSpanMismatch                  // matched but the matched text doesn't line up
	BucketTimeout                       // engine didn't return within the per-subject deadline
)

func (b Bucket) String() string {
	switch b {
	case BucketPass:
		return "PASS"
	case BucketSkipModifier:
		return "SKIP_MODIFIER"
	case BucketSkipUnparseable:
		return "SKIP_UNPARSEABLE"
	case BucketSkipDecode:
		return "SKIP_DECODE"
	case BucketParseError:
		return "PARSE_ERROR"
	case BucketEnginePanic:
		return "PANIC"
	case BucketMatchMismatch:
		return "MATCH_MISMATCH"
	case BucketGroupMismatch:
		return "GROUP_MISMATCH"
	case BucketSpanMismatch:
		return "SPAN_MISMATCH"
	case BucketTimeout:
		return "TIMEOUT"
	}
	return fmt.Sprintf("Bucket(%d)", int(b))
}

func (b Bucket) IsPass() bool { return b == BucketPass }
func (b Bucket) IsSkip() bool {
	return b == BucketSkipModifier || b == BucketSkipUnparseable || b == BucketSkipDecode
}
func (b Bucket) IsFail() bool {
	return !b.IsPass() && !b.IsSkip()
}

// Outcome is one (pattern, subject) result, ready for bucketing & logging.
type Outcome struct {
	ID     string // stable identifier: e.g. "testinput1:42"
	Bucket Bucket
	Reason string // human-readable detail; mostly populated on failures
}

// LoadConformancePatterns reads the testoutputN file (which embeds both
// patterns and expected results) into the per-pattern struct.
func LoadConformancePatterns(file string) ([]pcre2dsl.Pattern, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file, err)
	}
	defer f.Close()
	return pcre2dsl.ParseOutput(deriveInputName(file), f)
}

// deriveInputName turns "…/testoutput1" into "testinput1" so IDs reference
// the file pcre2 users will actually grep.
func deriveInputName(file string) string {
	base := file
	if i := strings.LastIndexByte(file, '/'); i >= 0 {
		base = file[i+1:]
	}
	return strings.Replace(base, "testoutput", "testinput", 1)
}

// EvaluateAll runs the engine against every (pattern, subject) and returns
// outcomes in stable order. inputName is used as the file portion of each ID
// (e.g. "testinput1"). A per-subject deadline of DefaultPerSubjectTimeout is
// applied so runaway backtracking can't lock up the run.
//
// If progress is non-nil, it is called once per pattern with (patternIndex,
// totalPatterns) — handy for the bless tool to print a live counter.
func EvaluateAll(inputName string, pats []pcre2dsl.Pattern, progress func(i, n int)) []Outcome {
	var out []Outcome
	for i, p := range pats {
		if progress != nil {
			progress(i, len(pats))
		}
		for _, s := range p.Subjects {
			out = append(out, Evaluate(inputName, p, s, DefaultPerSubjectTimeout))
		}
	}
	return out
}

// Evaluate runs one (pattern, subject) with a hard deadline and returns the
// classified outcome. Engine panics are caught; engine hangs are bounded by
// `timeout` (set <=0 to disable, useful only for tiny unit tests).
func Evaluate(inputName string, p pcre2dsl.Pattern, s pcre2dsl.Subject, timeout time.Duration) (oc Outcome) {
	oc.ID = fmt.Sprintf("%s:%d", inputName, s.LineNo)

	if p.SkipReason != "" {
		oc.Bucket = BucketSkipUnparseable
		oc.Reason = p.SkipReason
		return
	}
	if p.Modifiers != "" {
		oc.Bucket = BucketSkipModifier
		oc.Reason = fmt.Sprintf("modifiers %q not implemented", p.Modifiers)
		return
	}

	subject, ok := pcre2dsl.DecodeSubject(s.Raw)
	if !ok {
		oc.Bucket = BucketSkipDecode
		oc.Reason = "subject uses an unsupported escape form"
		return
	}

	// Wrap Parse+Search in a goroutine bounded by `timeout`. Parse can also
	// blow up (unrecognized constructs may loop the grammar), so it goes
	// inside the deadline too. Panics inside the goroutine are caught and
	// reported back via the result channel.
	resCh := make(chan engineResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resCh <- engineResult{Bucket: BucketEnginePanic, Reason: fmt.Sprintf("panic: %v", r)}
			}
		}()
		re, err := Parse(p.Pattern)
		if err != nil {
			resCh <- engineResult{Bucket: BucketParseError, Reason: fmt.Sprintf("Parse(%q): %v", p.Pattern, err)}
			return
		}
		resCh <- evaluateMatch(re, subject, s.Expect)
	}()

	if timeout > 0 {
		select {
		case r := <-resCh:
			oc.Bucket = r.Bucket
			oc.Reason = r.Reason
		case <-time.After(timeout):
			oc.Bucket = BucketTimeout
			oc.Reason = fmt.Sprintf("engine did not return within %v", timeout)
			// Note: the goroutine leaks until the engine eventually returns
			// (or never). For a one-shot bless / test-binary run this is
			// fine — the process exits and the OS reclaims.
		}
	} else {
		r := <-resCh
		oc.Bucket = r.Bucket
		oc.Reason = r.Reason
	}
	return
}

// engineResult is the channel-borne payload from the worker goroutine.
type engineResult struct {
	Bucket Bucket
	Reason string
}

// evaluateMatch compares the engine's match result against the expected
// groups. No timeout/panic handling here — caller wraps it appropriately.
func evaluateMatch(re *lib.RE, subject string, expect pcre2dsl.Expect) engineResult {
	res := re.Search(subject)
	gotMatched := res != nil && res.Matched

	if gotMatched != expect.Matched {
		return engineResult{BucketMatchMismatch, fmt.Sprintf("got matched=%v, want %v", gotMatched, expect.Matched)}
	}
	if !gotMatched {
		return engineResult{Bucket: BucketPass}
	}

	// Both matched: compare whole match text + captures.
	runes := []rune(subject)
	if res.Start < 0 || res.End > len(runes) || res.Start > res.End {
		return engineResult{BucketSpanMismatch, fmt.Sprintf("match span [%d,%d] out of range [0,%d]", res.Start, res.End, len(runes))}
	}
	whole := string(runes[res.Start:res.End])
	if len(expect.Groups) == 0 {
		// No expected group-0 line? Shouldn't happen for a positive match.
		return engineResult{Bucket: BucketPass}
	}
	expWhole := decodeOutGroup(expect.Groups[0])
	if whole != expWhole {
		return engineResult{BucketSpanMismatch, fmt.Sprintf("matched %q, want %q", whole, expWhole)}
	}

	// Compare capture groups 1..N. Pcre2 prints groups it captured; absence
	// of a group line means the group was unset.
	for i := 1; i < len(expect.Groups); i++ {
		var got string
		var gotSet bool
		if i-1 < len(res.Groups) && res.Groups[i-1] != nil {
			got = string(res.Groups[i-1])
			gotSet = true
		}
		want := expect.Groups[i]
		wantSet := want != nil
		if gotSet != wantSet {
			return engineResult{BucketGroupMismatch, fmt.Sprintf("group %d: got set=%v want set=%v", i, gotSet, wantSet)}
		}
		if wantSet {
			wantText := decodeOutGroup(want)
			if got != wantText {
				return engineResult{BucketGroupMismatch, fmt.Sprintf("group %d: got %q, want %q", i, got, wantText)}
			}
		}
	}
	return engineResult{Bucket: BucketPass}
}

func decodeOutGroup(p *string) string {
	if p == nil {
		return ""
	}
	d, ok := pcre2dsl.DecodeOutputText(*p)
	if !ok {
		return *p
	}
	return d
}

// LoadKnownFailures reads a file of "ID  # comment" lines into a set keyed
// by ID. Missing file → empty set (no error). Blank lines and full-line "#"
// comments are ignored.
func LoadKnownFailures(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	defer f.Close()
	return parseKnownFailures(f)
}

func parseKnownFailures(r io.Reader) (map[string]bool, error) {
	out := map[string]bool{}
	scn := bufio.NewScanner(r)
	for scn.Scan() {
		line := scn.Text()
		// strip trailing comment
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		id := strings.TrimSpace(line)
		if id == "" {
			continue
		}
		out[id] = true
	}
	return out, scn.Err()
}

// Stats counts outcomes per bucket.
type Stats struct {
	Counts         map[Bucket]int
	KnownFail      int // outcomes that failed AND were in the known-failures set
	UnexpectedPass int // outcomes that passed BUT were in the known-failures set
	UnexpectedFail int // outcomes that failed AND were NOT in known-failures
	Total          int
}

func NewStats() *Stats {
	return &Stats{Counts: map[Bucket]int{}}
}

func (s *Stats) Add(o Outcome, known map[string]bool) {
	s.Total++
	s.Counts[o.Bucket]++
	if o.Bucket.IsFail() {
		if known[o.ID] {
			s.KnownFail++
		} else {
			s.UnexpectedFail++
		}
		return
	}
	if o.Bucket.IsPass() && known[o.ID] {
		s.UnexpectedPass++
	}
}

func (s *Stats) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "PCRE2 conformance summary (%d outcomes total):\n", s.Total)
	for _, bk := range allBuckets {
		if n := s.Counts[bk]; n > 0 {
			fmt.Fprintf(&b, "  %-18s %d\n", bk.String(), n)
		}
	}
	fmt.Fprintf(&b, "  ----\n")
	fmt.Fprintf(&b, "  known-fail (ok):   %d\n", s.KnownFail)
	fmt.Fprintf(&b, "  unexpected pass:   %d\n", s.UnexpectedPass)
	fmt.Fprintf(&b, "  unexpected fail:   %d\n", s.UnexpectedFail)
	// pass rate among non-skipped outcomes
	nonSkip := s.Total
	for _, bk := range []Bucket{BucketSkipModifier, BucketSkipUnparseable, BucketSkipDecode} {
		nonSkip -= s.Counts[bk]
	}
	if nonSkip > 0 {
		pass := s.Counts[BucketPass]
		fmt.Fprintf(&b, "  pass rate (non-skipped): %d/%d = %.1f%%\n",
			pass, nonSkip, 100.0*float64(pass)/float64(nonSkip))
	}
	return b.String()
}

var allBuckets = []Bucket{
	BucketPass,
	BucketSkipModifier,
	BucketSkipUnparseable,
	BucketSkipDecode,
	BucketParseError,
	BucketEnginePanic,
	BucketMatchMismatch,
	BucketGroupMismatch,
	BucketSpanMismatch,
	BucketTimeout,
}
