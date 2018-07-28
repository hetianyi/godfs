
# reference:
# https://docs.docker.com/develop/develop-images/multistage-build/#before-multi-stage-builds
# NOTICE:Multi-stage builds are a new feature requiring Docker 17.05 or higher on the daemon and client.
# docker install : https://get.docker.com
# author : hetianyi(https://github.com/hetianyi)
# date   : 2018/07/18

FROM arm32v6/golang:alpine as builder
RUN echo "http://mirrors.aliyun.com/alpine/v3.8/main/" > /etc/apk/repositories && \
    apk add git gcc g++ make automake && \
    go get github.com/mattn/go-sqlite3 && \
    go install github.com/mattn/go-sqlite3
WORKDIR /tmp
RUN rm -rf godfs && git clone https://github.com/hetianyi/godfs.git && \
    cd godfs && mkdir data && cp conf/storage.db data/ && \
    rm -rf dev_tool && ./docker_build.sh && rm -rf docker && rm -rf src && cd .. && \
    tar czvf godfs.tar.gz godfs


FROM alpine:latest
COPY --from=0 /tmp/godfs.tar.gz /
RUN tar -xzf /godfs.tar.gz -C /
WORKDIR /godfs
ENTRYPOINT ["sh", "run.sh"]