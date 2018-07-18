FROM golang:alpine as builder

RUN apk add git
WORKDIR /tmp
RUN git clone https://github.com/hetianyi/godfs.git && \
    cd godfs && ./install.sh
CMD ["/tmp/godfs/bin/storage"]
