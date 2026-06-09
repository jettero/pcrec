package pcrec_test

import (
	"testing"

	"github.com/jettero/pcrec"
)

// TestPCRE2Conformance runs the vendored upstream PCRE2 corpus (testinput1
// only, for now) through pcrec and records pass/fail/skip per pattern×subject.
//
// Day-1 expectation: most things fail because anchors, escapes, lookaround,
// lazy quantifiers, backreferences and Unicode features are not implemented
// in the engine yet. The harness is therefore failure-list-driven: every
// known failure is recorded in testdata/pcre2_known_failures.txt, and the
// test passes if outcomes match that list. Regressions (an unexpected fail,
// or an unexpected pass that needs the list shortened) fail the test.
//
// To refresh the list after engine progress, run: go run ./cmd/pcre2-bless
func TestPCRE2Conformance(t *testing.T) {
	const outputFile = "contrib/pcre2/testdata/testoutput1"
	const knownFile = "testdata/pcre2_known_failures.txt"
	const inputName = "testinput1"

	pats, err := pcrec.LoadConformancePatterns(outputFile)
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	known, err := pcrec.LoadKnownFailures(knownFile)
	if err != nil {
		t.Fatalf("load known-failures: %v", err)
	}
	stats := pcrec.NewStats()

	outcomes := pcrec.EvaluateAll(inputName, pats, nil)
	for _, oc := range outcomes {
		stats.Add(oc, known)
		switch {
		case oc.Bucket.IsFail() && !known[oc.ID]:
			t.Errorf("UNEXPECTED_FAIL %s [%s]: %s", oc.ID, oc.Bucket, oc.Reason)
		case oc.Bucket.IsPass() && known[oc.ID]:
			t.Errorf("UNEXPECTED_PASS %s — remove from %s", oc.ID, knownFile)
		}
	}
	t.Log("\n" + stats.Summary())
}
