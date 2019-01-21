#!/bin/bash

export envTrackers="localhost:8015"

log() {
echo "`date '+%Y-%m-%d %H:%M:%S'`" $*
}

trackers=(${envTrackers//,/ })

if [ ${#trackers[@]} == 0 ];then
    echo "no tracker configured"
    exit 1
fi

# curl -X POST http://${trackers[0]}/nginx
serverF=""


for ele in ${trackers[@]}
do
    log "request tracker: "$ele
    temp=$(curl -s $ele/servers)
    log "response is: <$temp>"
    code=$?
    if [ $code != 0 ]; then
        log "error request url($code):${ele}/servers"
    elif [[ $temp =~ ^\[.*]$ ]]; then
        serverF=${serverF}" -F 'servers=$temp'"
        availableS=$ele
    fi
done

if [ "$availableS"x == ""x ];then
    log "no available tracker server!"
    exit 2
fi

serverF="curl -sX POST http://${availableS}/nginx -F template=@nginx-structed-template.conf ${serverF}"

log $serverF

sh -c "$serverF"