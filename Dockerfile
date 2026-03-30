# Builder stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o html-to-image ./cmd/html-to-image/main.go

# Final stage
FROM alpine:latest

# Install Chromium and dependencies
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app
COPY --from=builder /app/html-to-image .

# Set chrome path if needed (chromedp usually finds it)
ENV CHROME_PATH=/usr/bin/chromium-browser

ENTRYPOINT ["./html-to-image"]
