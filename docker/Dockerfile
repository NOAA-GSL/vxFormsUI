# Start from the official Golang image for building
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
RUN go build -o vxformsui

# Start a minimal image for running
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy the built binary and static/templates
COPY --from=builder /app/vxformsui .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose the port (default 8080)
EXPOSE 8080

# Run the binary
CMD ["./vxformsui"]
