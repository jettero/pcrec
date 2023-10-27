#!/usr/bin/env bash

if [[ "$*" =~ -[a-zA-Z]*D || "$*" =~ --dot ]]; then
    t="$(mktemp /tmp/pat-XXX.dot)"
    go run ./cmd/expr/main.go "$@" | tee $t
    ls -1tr /tmp/pat-*.dot | head -n -5 | xargs -r rm -v

else go run ./cmd/expr/main.go "$@"
fi
