FROM hehety/golang:1.12.13-alpine3.9-arm32v7
ENV GOPROXY https://goproxy.io

RUN cd /root && \
    git clone https://github.com/hetianyi/godfs && \
    cd godfs && \
    git checkout 2.0.0-dev && \
    sh make.sh


FROM arm32v7/alpine:3.10.3
COPY --from=0 /root/godfs/bin/godfs /usr/bin
RUN apk add tzdata && apk add libc6-compat
WORKDIR /
