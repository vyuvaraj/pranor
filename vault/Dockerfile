FROM golang:alpine AS builder
WORKDIR /src
COPY . .
RUN sed -i 's/go 1\.2[5-9]\.[0-9]*/go 1.24/g' go.mod vendor/modules.txt 2>/dev/null || true && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /pranor-vaultd ./cmd/pranor-vaultd && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /pranor-vault ./cmd/pranor-vault

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /pranor-vaultd /usr/local/bin/pranor-vaultd
COPY --from=builder /pranor-vault /usr/local/bin/pranor-vault

EXPOSE 9000 9001
VOLUME /data

ENTRYPOINT ["pranor-vaultd"]
CMD ["-port", ":9000", "-admin-port", ":9001"]
