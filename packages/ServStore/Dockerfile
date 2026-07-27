FROM golang:alpine AS builder
WORKDIR /src
COPY . .
RUN sed -i 's/go 1\.2[5-9]\.[0-9]*/go 1.24/g' go.mod vendor/modules.txt 2>/dev/null || true && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /servstored ./cmd/servstored && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /servstore ./cmd/servstore

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /servstored /usr/local/bin/servstored
COPY --from=builder /servstore /usr/local/bin/servstore

EXPOSE 9000 9001
VOLUME /data

ENTRYPOINT ["servstored"]
CMD ["-port", ":9000", "-admin-port", ":9001"]
