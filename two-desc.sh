#!/usr/bin/env bash

for item in "$@"; do
    echo;echo "----=: $item :=---=: $(date) :=----"
    go run ./cmd/expr/main.go "$item"
    echo
    go run ./cmd/expr/main.go -D "$item" | tee pat.dot
done
