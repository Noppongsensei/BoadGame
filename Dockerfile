FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install git (required for go get)
RUN apk add --no-cache git

# Make module download more reliable in CI/VPS environments
ENV GOPROXY=https://proxy.golang.org,direct

# Copy source code
COPY . .

# Initialize go modules, fetch all dependencies and build with flags to ignore unused variables
RUN (go mod download || (sleep 2 && go mod download) || (sleep 5 && go mod download)) && \
    go build -mod=mod -gcflags="all=-e" -ldflags="-s -w" -o avalon-server ./cmd/avalon

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Check if binary exists and then copy it
RUN mkdir -p /app

# Copy binary from builder stage
COPY --from=builder /app/avalon-server /app/

# Verify binary exists
RUN ls -la /app

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/avalon-server"]
