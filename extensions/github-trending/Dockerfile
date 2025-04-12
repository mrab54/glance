# --- Build Stage ---
# Use the alpine variant for a smaller base with potentially fewer vulnerabilities
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module files and download dependencies first
# (If you don't have go.mod/go.sum yet, you might need to run
# `go mod init github.com/youruser/github-trending` and `go mod tidy` first,
# or adjust the COPY commands)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY *.go ./

# Build the application statically
# Ensure CGO_ENABLED=0 for a static binary suitable for scratch
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o github-trending .
# Verify the binary was built
RUN ls -l github-trending

# --- Final Stage ---
# Use alpine instead of scratch to include necessary libraries like CA certificates
FROM alpine:latest

# Install ca-certificates (might be needed for HTTPS requests)
RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the static binary from the builder stage
COPY --from=builder /app/github-trending .
# Verify the binary was copied and check permissions
RUN ls -l /app/github-trending

# Expose port 8081
EXPOSE 8081

# Command to run the executable
CMD ["/app/github-trending"]
