# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Download Go dependencies first (Docker cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Statically compile the binary (CGO_ENABLED=0 is ideal for Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o velov-radar .

# Final lightweight image stage
FROM alpine:latest

# Install SSL certificates (needed for Telegram API over HTTPS)
# and tzdata for proper timezone handling
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/velov-radar .

# Create an empty users.json by default (will be overridden by volume mount)
RUN touch users.json

# Startup command
CMD ["./velov-radar"]

