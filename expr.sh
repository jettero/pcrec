#!/usr/bin/env bash

if [[ "$*" =~ -[a-zA-Z]*D || "$*" =~ --dot ]]; then
    if go run ./cmd/expr/main.go "$@" > pat.dot
    then cat pat.dot
    else rm -v pat.dot; exit 1
    fi

else go run ./cmd/expr/main.go "$@"
fi
