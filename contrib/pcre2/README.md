# Vendored PCRE2 test corpus

The contents of `testdata/` and `LICENCE` in this directory are **copied
verbatim from the PCRE2 project** and used here solely as a conformance
corpus for the `pcrec` regex engine. None of the test data was authored by
the `pcrec` project.

Upstream: https://github.com/PCRE2Project/pcre2

Pinned SHA: `ff92e0b9cea5b5ae3af12ba930d03556684f098b`

## Licence

PCRE2 itself is BSD-licensed; the upstream `LICENCE.md` file (vendored
alongside this README, with the short `COPYING` summary) explicitly states:

> The data in the testdata directory is not copyrighted and is in the public
> domain.

So the test files in `testdata/` carry no redistribution obligations. The
PCRE2 `LICENCE` is included here as the canonical reference for that
statement and for the BSD terms of the rest of PCRE2. (The pcre2-test data
is itself public domain, but PCRE2 the library — which we do not bundle —
is BSD.)

## Refreshing

Run `./FETCH.sh` from this directory. The script downloads the upstream
archive at the pinned SHA, replaces `testdata/` and `LICENCE` with the
upstream copies, and re-stamps the SHA into this README. To move to a newer
upstream snapshot, edit `PCRE2_SHA` at the top of `FETCH.sh` and run it.

## What we actually consume

`pcre2test`, the upstream test harness, drives matching from these files in
a custom plain-text DSL: a pattern in delimiters, indented subject lines, a
paired `testoutputN` file that captures expected output.

`pcrec` does not link the C library or shell out to `pcre2test`. Instead the
Go test harness (`pcre2_conformance_test.go` at the project root) parses the
DSL itself, drives the Go engine, and compares against the expected matches
parsed structurally out of `testoutputN`.

Initial scope: only `testinput1` + `testoutput1` are wired into the driver.
The rest of the files are vendored so the corpus is honest and complete, and
so wiring additional sets (UTF, Unicode properties, etc.) is a one-line
change later.
