# Start with a base image
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module dependency files
COPY go.mod .
COPY go.sum .

# Download Go module dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o main.exe ./cmd/main.go

# Start a new stage for the final lightweight container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app main.exe

# Set execute permission for the binary
RUN chmod +x ./main.exe

# Set the command to run the binary
CMD ["./main.exe"]
