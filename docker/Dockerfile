# GoDFS dockerfile
# author : hetianyi(https://github.com/hetianyi)
# date   : 2019/11/19

FROM golang:1.12-alpine3.9
RUN apk add git && \
    cd / && \
    git clone https://github.com/hetianyi/godfs.git && \
    cd godfs && \
    git checkout 2.0.0-dev && \
    ./make.sh


FROM alpine:latest
COPY --from=0 /godfs/bin/godfs /usr/bin
RUN apk add tzdata && apk add libc6-compat
WORKDIR /
