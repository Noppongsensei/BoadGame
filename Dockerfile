FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Go application
RUN go build -o avalon-server ./cmd/avalon

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/avalon-server .

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./avalon-server"]
