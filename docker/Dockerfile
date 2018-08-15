
# reference:
# https://docs.docker.com/develop/develop-images/multistage-build/#before-multi-stage-builds
# NOTICE:Multi-stage builds are a new feature requiring Docker 17.05 or higher on the daemon and client.
# docker install : https://get.docker.com
# start script:
#   start tracker:(single local machine may cause storage connect failed with tracker,
#   and ensure that the firewall opens port like 1022 and 8001)
# docker run -d -p 1022:1022 --name tracker [--net host] godfs:latest tracker
#   start storage:
# docker run -d -p 1024:1024 -p 8001:8001 -e trackers="192.x.x.x:1022" --name storage godfs:latest storage
# author : hetianyi(https://github.com/hetianyi)
# date   : 2018/07/18

FROM golang:alpine
RUN echo "http://mirrors.aliyun.com/alpine/v3.8/main/" > /etc/apk/repositories && \
    echo "http://mirrors.aliyun.com/alpine/v3.8/community/" >> /etc/apk/repositories && \
    echo "192.30.253.112 github.com" >> /etc/hosts && echo "151.101.113.194 github.global.ssl.fastly.net" >> /etc/hosts && \
    apk add git gcc g++ cmake make automake && \
    go get github.com/mattn/go-sqlite3
RUN cd /tmp && rm -rf godfs && git clone https://github.com/hetianyi/godfs.git && \
    cd godfs && mkdir data && cp conf/storage.db data/ && \
    rm -rf dev_tool && ./docker_build.sh && rm -rf docker && rm -rf src && cd .. && \
    tar czvf godfs.tar.gz godfs


FROM alpine:latest
COPY --from=0 /tmp/godfs.tar.gz /
RUN tar -xzf /godfs.tar.gz -C / && rm -f /godfs.tar.gz && \
    echo "http://mirrors.aliyun.com/alpine/v3.8/main/" > /etc/apk/repositories && \
    echo "http://mirrors.aliyun.com/alpine/v3.8/community/" >> /etc/apk/repositories && \
    apk add tzdata
WORKDIR /godfs
ENTRYPOINT ["sh", "docker_run.sh"]