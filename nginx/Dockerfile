FROM alpine:3.15 as base-image

ENV NGINX_VERSION nginx-1.21.4
ENV FFMPEG_VERSION snapshot
ENV NGINX_RTMP_MODULE_VERSION 1.2.2
ENV TZ=America/New_York
ENV LD_LIBRARY_PATH=/usr/local/lib

FROM base-image as builder
RUN apk add --no-cache \
    git wget cmake openssl-dev build-base librtmp yasm ca-certificates openssl pcre-dev zlib zlib-dev

# Install nginx.
WORKDIR /tmp/build/nginx
RUN wget -O "${NGINX_VERSION}.tar.gz" "https://nginx.org/download/${NGINX_VERSION}.tar.gz" && \
    tar -zxf "${NGINX_VERSION}.tar.gz" && \
    rm "${NGINX_VERSION}.tar.gz"

# Download and decompress RTMP module
WORKDIR /tmp/build/nginx-rtmp-module
RUN wget -O "nginx-rtmp-module-${NGINX_RTMP_MODULE_VERSION}.tar.gz" "https://github.com/arut/nginx-rtmp-module/archive/v${NGINX_RTMP_MODULE_VERSION}.tar.gz" && \
    tar -zxf "nginx-rtmp-module-${NGINX_RTMP_MODULE_VERSION}.tar.gz" && \
    rm "nginx-rtmp-module-${NGINX_RTMP_MODULE_VERSION}.tar.gz"

WORKDIR /tmp/build/nginx/"${NGINX_VERSION}"
RUN ./configure \
    --sbin-path=/usr/local/sbin/nginx \
    --conf-path=/etc/nginx/nginx.conf \
    --error-log-path=/var/log/nginx/error.log \
    --pid-path=/var/run/nginx/nginx.pid \
    --lock-path=/var/lock/nginx/nginx.lock \
    --http-log-path=/var/log/nginx/access.log \
    --http-client-body-temp-path=/tmp/nginx-client-body \
    --with-http_ssl_module \
    --with-threads \
    --with-ipv6 \
    --add-module="/tmp/build/nginx-rtmp-module/nginx-rtmp-module-${NGINX_RTMP_MODULE_VERSION}" \
    --with-cc-opt="-Wimplicit-fallthrough=0" \
    --with-debug
RUN make -j "$(getconf _NPROCESSORS_ONLN)"

FROM base-image

RUN apk add --no-cache \
    make ca-certificates openssl ffmpeg

COPY --from=builder /tmp/build/ /tmp/build/

WORKDIR /tmp/build/nginx/"${NGINX_VERSION}"
RUN make install -j "$(getconf _NPROCESSORS_ONLN)"
WORKDIR /

RUN mkdir /var/lock/nginx
RUN rm -rf /tmp/build

# Forward logs to Docker
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log

COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 1935

CMD ["nginx", "-g", "daemon off;"]
