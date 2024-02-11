# Start from the official golang image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Change directory to where your main.go is located
WORKDIR /app/cmd

# Build the Go app
RUN go build -o /app/main .

# Set back the working directory
WORKDIR /app

# Add wait-for-it script
ADD https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh

# Command to run the executable
CMD ["./wait-for-it.sh", "mongo:27017", "--", "./main"]
