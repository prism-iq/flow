# Go server Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build without CGo (pure Go fallback)
RUN CGO_ENABLED=0 GOOS=linux go build -o /flow-server ./cmd/server

# Runtime image
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl

# Copy binary
COPY --from=builder /flow-server /app/flow-server

EXPOSE 8090

CMD ["/app/flow-server"]
