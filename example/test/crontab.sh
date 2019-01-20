#!/bin/bash
export envTrackers="localhost:8015"
trackers=(${envTrackers//,/ })

if [ ${#trackers[@]} == 0 ];then
    echo "no tracker configured"
    exit 1
fi

serverF="curl -X POST http://${trackers[0]}/nginx \
         -F template=@nginx-structed-template.conf"

for ele in ${trackers[@]}
do
    echo "querying from "$ele
    temp=$(curl -s $ele/servers)
    code=$?
    if [ $code != 0 ]; then
        echo "error request url($code):${ele}/servers"
        exit $code
    fi
    serverF=${serverF}" -F 'servers=$temp'"
done


echo $serverF