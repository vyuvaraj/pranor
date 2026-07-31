FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o pranor-flow

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/pranor-flow .
EXPOSE 8096
ENTRYPOINT ["./pranorflow"]
