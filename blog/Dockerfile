FROM golang:1.24.13-bookworm AS builder

COPY . /src
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler \
    && rm -rf /var/lib/apt/lists/

RUN  GOPROXY=https://goproxy.cn  make api  \
    && GOPROXY=https://goproxy.cn  make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

WORKDIR /app
COPY --from=builder /src/bin /app


EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
