# Build stage
FROM golang:1.25.9 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
COPY . .
RUN go mod tidy
# Build statically with stripped symbols
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o app ./cmd/api

# Runtime
FROM alpine:3.18 AS runtime
RUN apk add --no-cache ca-certificates curl
RUN adduser -D -u 1000 app
WORKDIR /app
# Copy binary and migrations from builder
COPY --from=builder /app/app .
COPY --from=builder /app/internal/infrastructure/database/migrations ./internal/infrastructure/database/migrations
USER 1000
ENV GIN_MODE=release
CMD ["./app"]
