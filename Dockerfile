FROM golang:alpine as builder
RUN apk add git
WORKDIR /tmp
RUN git clone https://github.com/hetianyi/godfs.git && \
    cd godfs && ./install.sh
CMD ["/tmp/godfs/bin/storage"]


FROM alpine:latest
WORKDIR /tmp
COPY --from=0 /tmp/godfs/bin/storage .
COPY --from=0 /tmp/godfs/conf/storage.conf .
CMD ["/tmp/storage", "-c", "storage.conf"]