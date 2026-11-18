# ==========================================
# Stage 1: Build the Go Backend (+ embedded static UI)
# ==========================================
FROM golang:1.26-alpine AS backend-builder

WORKDIR /build

# Copy go.mod and go.sum and fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and static frontend assets (embedded via //go:embed)
COPY main.go ./
COPY internal/ ./internal/
COPY static/ ./static/

# Build a static Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gym-rabbit main.go

# ==========================================
# Stage 2: Runner Container
# ==========================================
FROM alpine:3.20

# Install ca-certificates and timezone data (useful for logs/scheduling)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the static binary (includes embedded static UI)
COPY --from=backend-builder /build/gym-rabbit /app/gym-rabbit

# Define volume for persistent database storage
VOLUME ["/app/data"]

# Expose port (default is 8080)
EXPOSE 8080

# Run the backend binary
CMD ["./gym-rabbit"]
