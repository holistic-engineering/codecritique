# Stage 1: Build the application
FROM golang:1.22.5-alpine AS builder

# Install git and SSL certificates
RUN apk add --no-cache git ca-certificates

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o codecritique ./cmd/cli

# Stage 2: Create the final lightweight image
FROM alpine:3.19

# Install SSL certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/codecritique .

# Copy any additional configuration files if needed
# COPY --from=builder /app/config/config.toml .

# Expose any necessary ports
# EXPOSE 8080

# Command to run the executable
CMD ["./codecritique"]