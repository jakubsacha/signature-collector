# Use the official Go image as a builder
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk --no-cache add tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set timezone
ENV TZ=UTC

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
