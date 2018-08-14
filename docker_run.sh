#!/bin/sh

case $1 in
    tracker)
        bin/tracker $*
    ;;
    storage)
        bin/storage $*
    ;;
    *)
        echo "unsupport command: $1"
esac
exit 0