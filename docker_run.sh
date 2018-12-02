#!/bin/sh

case $1 in
    tracker)
        bin/tracker $*
    ;;
    storage)
        bin/storage $*
    ;;
    bashboard)
        bin/bashboard $*
    ;;
    *)
        echo "unknown command: $1"
esac
exit 0