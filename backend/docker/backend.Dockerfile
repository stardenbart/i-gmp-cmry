# syntax=docker/dockerfile:1
FROM docker.io/library/golang:1.25-alpine3.22 AS builder

WORKDIR /app

# Install dependensi
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy source code
COPY . .

# Build aplikasi
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ─────────────────────────────────────────────
# Runner stage (image kecil tanpa toolchain go)
FROM docker.io/library/alpine:3.22

WORKDIR /app

# Set timezone
RUN apk add --no-cache ca-certificates su-exec tzdata \
    && addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app

# Copy binary & config file dari builder
COPY --from=builder /out/server ./main
COPY --from=builder /out/migrate ./migrate
# Copy direktori migrasi dan foto upload jika dibutuhkan (opsional)
COPY --from=builder /app/migrations ./migrations
# Copy templates Excel untuk fitur export laporan
COPY --from=builder /app/templates ./templates
RUN mkdir -p uploads logs

COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod 0755 /usr/local/bin/entrypoint.sh \
    && chown -R app:app /app

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["./main"]
