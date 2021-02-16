#!/bin/sh

{

git log --format="%ai %h" --reverse \
    | awk '
    {
        # find the first diff date, print prev
        if ($1 != prev) {print prevhash }
            prev = $1
            prevhash = $4
    }
    '

    git log -n1 --format="%h"
} \
    | while read h; do
        if [ ".$prev" == "." ]; then
            prev=$h
            continue
        fi

        git log -n1 $h --format="%ai" \
            | awk '{print "-   " $1 ":"}'
        git diff $prev $h --shortstat -- . ':(exclude)*.pb.go' \
            | awk '{ print "    LOC: +" $4 " -" $6 ""} '
        echo

        echo '    ```'
        git log  $prev..$h --reverse --format="%s" \
            | awk '{ gsub("day-.: ", "", $0); print "    " $0 }'
        echo '    ```'
        echo
        prev=$h
done > docs/log.md
