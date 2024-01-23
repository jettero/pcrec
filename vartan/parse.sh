#!/usr/bin/env bash

set -e
make regexp.json

set -x
echo -n "$*" | vartan parse regexp.json
