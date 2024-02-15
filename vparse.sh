#!/usr/bin/env bash

HERE="$( dirname "$0" )"

set -e
cd "$HERE/vartan"

(set -x; make regexp.json)

echo -n "$*" | (set -x; vartan parse regexp.json)
