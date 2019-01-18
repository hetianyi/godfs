FROM golang:alpine
RUN echo "http://mirrors.aliyun.com/alpine/v3.8/main/" > /etc/apk/repositories && \
    echo "http://mirrors.aliyun.com/alpine/v3.8/community/" >> /etc/apk/repositories && \
    # faster in china
    echo "192.30.253.112 github.com" >> /etc/hosts && echo "151.101.113.194 github.global.ssl.fastly.net" >> /etc/hosts && \
    apk add git gcc g++ cmake make automake && \
    go get github.com/mattn/go-sqlite3 && \
    go get github.com/jinzhu/gorm && \
    go get github.com/json-iterator/go && \
    go get github.com/urfave/cli