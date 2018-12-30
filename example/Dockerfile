# build nginx docker images with stream module
FROM alpine
MAINTAINER "hehety<hehety@outlook.com>"
RUN apk add libxslt-dev && apk add libxml2 && \
        apk add gd-dev && \
        apk add geoip-dev && \
        apk add libatomic_ops-dev && \
        apk add pcre-dev && \
        apk add zlib-dev && \
        apk add build-base && \
        apk add libaio-dev && \
        apk add openssl-dev && \
cd /tmp && \
wget http://nginx.org/download/nginx-1.15.8.tar.gz && \
tar -xzf nginx-1.15.8.tar.gz && \
cd nginx-1.15.8 && \
./configure --prefix=/usr/local/nginx \
--with-select_module \
--with-poll_module \
--with-threads \
--with-http_ssl_module \
--with-http_v2_module \
--with-http_realip_module \
--with-http_addition_module \
--with-http_xslt_module \
--with-http_image_filter_module \
--with-http_geoip_module \
--with-http_sub_module \
--with-http_dav_module \
--with-http_gunzip_module \
--with-http_gzip_static_module \
--with-http_auth_request_module \
--with-http_random_index_module \
--with-http_secure_link_module \
--with-http_degradation_module \
--with-http_slice_module \
--with-http_stub_status_module \
--with-stream \
--with-stream_ssl_module \
--with-stream_realip_module \
--with-stream_geoip_module \
--with-stream_ssl_preread_module \
--with-compat \
--with-libatomic \
--with-debug && \
make && make install

FROM alpine
COPY --from=0 /usr/local/nginx /usr/local/nginx
RUN apk add libxslt libxml2 gd geoip libatomic_ops pcre zlib openssl
WORKDIR /usr/local/nginx
ENV PATH /usr/local/nginx/sbin:$PATH
CMD ["nginx", "-g", "daemon off;"]