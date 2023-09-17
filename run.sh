#!/usr/bin/env bash

export PCREC_TRACE=1

go run ./cmd/main.go "$@"
