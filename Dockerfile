FROM golang:1.11.5-alpine3.9 as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app/
RUN apk add --no-cache git gcc libc-dev && \
    go build -o /app/wg-dynamic .


FROM alpine:3.9
COPY --from=builder /app/wg-dynamic /app/
WORKDIR /app
RUN apk add --no-cache \
        ca-certificates \
        libmnl iproute2 iptables

COPY --from=r.j3ss.co/wireguard:install /usr/bin/wg /usr/bin/wg
COPY --from=r.j3ss.co/wireguard:install /usr/share/man/man8/wg.8 /usr/share/man/man8/wg.8

ENTRYPOINT ["/app/wg-dynamic"]
CMD ["--help"]
