# Gunakan image golang alpine untuk build
FROM golang:alpine AS builder

WORKDIR /app

# Install dependensi
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# ─────────────────────────────────────────────
# Runner stage (image kecil tanpa toolchain go)
FROM alpine:latest

WORKDIR /app

# Set timezone
RUN apk add --no-cache tzdata

# Copy binary & config file dari builder
COPY --from=builder /app/main .
# Copy direktori migrasi dan foto upload jika dibutuhkan (opsional)
COPY --from=builder /app/migrations ./migrations
# Copy templates Excel untuk fitur export laporan
COPY --from=builder /app/templates ./templates
RUN mkdir -p uploads logs

EXPOSE 8080

CMD ["./main"]
