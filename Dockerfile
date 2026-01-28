# ===== build stage =====
FROM golang:1.25.4-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /app/bin/url-shortener ./cmd/url-shortener

# ===== runtime stage =====
FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/url-shortener /app/url-shortener
COPY --from=builder /app/config /app/config


EXPOSE 8080
ENTRYPOINT ["/app/url-shortener"]
