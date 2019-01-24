#!/bin/bash

cd /cronjob

log() {
echo "`date '+%Y-%m-%d %H:%M:%S'`" $*
}

trackersArr=(${trackers//,/ })

if [ ${#trackersArr[@]} == 0 ];then
    echo "no tracker configured"
    exit 1
fi

# curl -X POST http://${trackersArr[0]}/nginx
serverF=""


for ele in ${trackersArr[@]}
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

serverF="curl -o nginx.conf -sX POST http://${availableS}/nginx -F template=@nginx-structed-template.conf ${serverF}"

sh -c "$serverF"

code=$?
if [ $code != 0 ]; then
    log "error do request($code):$serverF"
    exit $code
fi

tempFileName=$(echo "`date '+%Y%m%d%H%M%S'`")
[ -f nginx.conf ] && mv /usr/local/nginx/conf/nginx.conf /usr/local/nginx/conf/nginx.conf.bak.$tempFileName && mv nginx.conf /usr/local/nginx/conf

nginx -t
code=$?
if [ $code != 0 ]; then
    log rollback config file
    [ -f /usr/local/nginx/conf/nginx.conf.bak.$tempFileName ] \
    && rm -rf /usr/local/nginx/conf/nginx.conf && mv /usr/local/nginx/conf/nginx.conf.bak.$tempFileName /usr/local/nginx/conf/nginx.conf
fi

# clean
log clean...
rm -f nginx.conf
rm -f /usr/local/nginx/conf/nginx.conf.bak.$tempFileName

nginx -s reload