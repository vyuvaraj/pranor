FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o pranor-auth

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/pranor-auth .
EXPOSE 8098
ENTRYPOINT ["./pranorauth"]
