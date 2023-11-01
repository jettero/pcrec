#!/usr/bin/env bash

if [ "$1" = "-x" ]
then v=1; shift
else v=0
fi

if [[ "$*" =~ -[a-zA-Z]*D || "$*" =~ --dot ]]; then
    if go run ./cmd/expr/main.go "$@" > pat.dot
    then cat pat.dot ; if [ X$v = X1 ]; then xdot pat.dot; fi
    else rm -v pat.dot; exit 1
    fi

else go run ./cmd/expr/main.go "$@"
fi
