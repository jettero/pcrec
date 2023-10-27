#!/bin/bash

tmux new-window bash -c 'for i in {1..60}; do echo; done
    imgcat <( ls -1tr /tmp/pat-*.dot | tail -n1 | xargs -r dot -Tpng )
    read X'
