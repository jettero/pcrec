#!/bin/bash
# Refetch the vendored PCRE2 test corpus.
#
# Bumps PCRE2_SHA at the top, re-downloads the archive tarball, and overwrites
# contrib/pcre2/testdata/ + contrib/pcre2/LICENCE with the upstream files
# untouched. After running, eyeball the diff before committing.
#
# The pinned SHA is written into README.md so the recorded provenance and the
# actual on-disk contents can't drift.

PCRE2_SHA="ff92e0b9cea5b5ae3af12ba930d03556684f098b"
PCRE2_REPO="PCRE2Project/pcre2"

trap "exit 1" INT

here=$(dirname "$0")
cd "$here" || exit 1

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

url="https://github.com/${PCRE2_REPO}/archive/${PCRE2_SHA}.tar.gz"
echo "fetching ${url}"
if ! curl -sSL "$url" -o "$tmp/pcre2.tar.gz"
then
    echo "fetch failed" >&2
    exit 1
fi

echo "extracting"
tar -xz -C "$tmp" -f "$tmp/pcre2.tar.gz" || exit 1

root="$tmp/pcre2-${PCRE2_SHA}"
if ! test -d "$root/testdata"
then
    echo "expected ${root}/testdata, not found" >&2
    exit 1
fi

# Wipe and repopulate the vendored slice, so removed-upstream files vanish here too.
rm -rf testdata
mkdir testdata
rsync -va "$root/testdata/" testdata/

cp "$root/LICENCE.md" LICENCE.md
cp "$root/COPYING" COPYING

# Stamp the SHA into README.md so the recorded provenance matches reality.
if test -f README.md
then
    sed -i -E "s/^Pinned SHA:.*/Pinned SHA: \`${PCRE2_SHA}\`/" README.md
fi

echo "done"
echo "  pinned SHA: ${PCRE2_SHA}"
echo "  testdata files: $(find testdata -type f | wc -l)"
echo "  total size:     $(du -sh testdata | cut -f1)"
