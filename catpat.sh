#!/usr/bin/env bash

x=$#

for i in "$@"; do

    NFA_TRACE=1 ./expr.sh -D $i \
        && dot pat.dot -Lg -Tpng > pat.png \
        && imgcat -H 90% pat.png
    
    if [ $x -gt 1 ]
    then read -ep "(pause)"
    fi

    x=$(( x - 1 ))
done
