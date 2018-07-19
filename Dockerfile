
# reference:
# https://docs.docker.com/develop/develop-images/multistage-build/#before-multi-stage-builds
# NOTICE:Multi-stage builds are a new feature requiring Docker 17.05 or higher on the daemon and client.
# docker install : https://get.docker.com
# author : hetianyi(https://github.com/hetianyi)
# date   : 2018/07/18

FROM golang:alpine as builder
RUN apk add git gcc && \
    go get github.com/mattn/go-sqlite3 && \
    go install github.com/mattn/go-sqlite3
WORKDIR /tmp
RUN rm -rf godfs && git clone https://github.com/hetianyi/godfs.git && \
    cd godfs && ./install.sh
CMD ["/tmp/godfs/bin/storage"]


FROM alpine:latest
WORKDIR /tmp
COPY --from=0 /tmp/godfs/bin/storage .
COPY --from=0 /tmp/godfs/conf/storage.conf .
CMD ["/tmp/storage", "-c", "storage.conf"]