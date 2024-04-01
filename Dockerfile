# Use the official Go image to create a build artifact.
# This is based on Debian and includes the Go toolchain.
FROM golang:1.21 as builder

# Create a directory for the application.
WORKDIR /app

# Copy the go.mod and go.sum to download all dependencies.
COPY go.mod ./
COPY go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed.
RUN go mod download

# Copy the source code into the container.
COPY . .

# Build the application as a static binary.
# -o specifies the output file name
# CGO_ENABLED=0 disables CGO which ensures a static binary is built
# -ldflags="-s -w" strips debugging information to reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pubg-leaderboard ./cmd/pubg-leaderboard/main.go

# Use a minimal image to run the application.
FROM alpine:latest  

# Install ca-certificates in case your application makes external network calls.
# If not needed, you can remove the ca-certificates line.
RUN apk --no-cache add ca-certificates

# Create a non-root user to run the application.
RUN adduser -D appuser

# Use the user just created.
USER appuser

# Copy the statically built binary from the builder stage.
COPY --from=builder /app/pubg-leaderboard .

# Expose the port the application listens on.
EXPOSE 8080

# Run the binary.
CMD ["./pubg-leaderboard"]
