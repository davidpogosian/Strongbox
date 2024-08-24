# Stage 1: Build the Go binary
FROM golang:1.22.5-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o strongbox

# Stage 2: A minimal image with just the Go binary
FROM alpine:3.18

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /app/strongbox .

# Copy the .env file
COPY .env .

# Copy the web/templates directory
COPY web/template/ /app/web/template/

COPY web/static /app/web/static

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./strongbox"]
