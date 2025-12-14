FROM golang:1.25.0-alpine AS builder
    WORKDIR /app

    COPY ./ ./
    ENV CGO_ENABLED=0
    RUN go build -o service ./cmd/app/main.go

FROM scratch

    WORKDIR /app
    COPY --from=builder /app/service .

    CMD ["./service"]

