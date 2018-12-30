#!/bin/bash

if [ "$1"x == ""x ]; then
    echo "usage:"
    echo "  start-example.sh <hostname or ip>"
    echo "  hostname or ip represents a hostname or an ip address that can be reached by client"
    exit 0
fi

export PROXY_HOST=$1
docker stack deploy -c docker-compose.yml --prune godfs


