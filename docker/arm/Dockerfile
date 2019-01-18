
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

FROM hehety/golang:arm32v6-godfs
RUN mkdir /godfs && cd /tmp && rm -rf godfs && git clone https://github.com/hetianyi/godfs.git && \
    cd godfs && mkdir data && cp conf/storage.db data/ && \
    ./make.sh && mv bin /godfs && mv conf /godfs && mv data /godfs && mv docker_run.sh /godfs && \
    cd / && tar czvf godfs.tar.gz godfs


FROM arm32v6/alpine
COPY --from=0 /godfs.tar.gz /
RUN tar -xzf /godfs.tar.gz -C / && rm -f /godfs.tar.gz && \
    # faster in china
    echo "http://mirrors.aliyun.com/alpine/v3.8/main/" > /etc/apk/repositories && \
    echo "http://mirrors.aliyun.com/alpine/v3.8/community/" >> /etc/apk/repositories && \
    apk add tzdata && apk add libc6-compat
WORKDIR /godfs/bin
ENV  PATH="/godfs/bin:${PATH}"
CMD ["sh", "docker_run.sh"]