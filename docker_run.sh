#!/bin/sh

case $1 in
    tracker)
        bin/tracker $*
    ;;
    storage)
        bin/storage $*
    ;;
    dashboard)
        bin/dashboard $*
    ;;
    *)
        echo "unknown command: $1"
    ;;
esac
exit 0