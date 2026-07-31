FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o pranor-notify

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/pranor-notify .
EXPOSE 8094
ENTRYPOINT ["./pranormail"]
