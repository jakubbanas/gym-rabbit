# ==========================================
# Stage 1: Build the React Frontend
# ==========================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend

# Install dependencies using npm ci for clean, deterministic builds
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

# Copy the remaining frontend code and build
COPY frontend/ ./
RUN npm run build

# ==========================================
# Stage 2: Build the Go Backend
# ==========================================
FROM golang:1.26-alpine AS backend-builder

WORKDIR /build

# Copy go.mod and go.sum and fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the backend source files
COPY main.go ./
COPY internal/ ./internal/

# Build a highly optimized, static Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gym-rabbit main.go

# ==========================================
# Stage 3: Runner Container
# ==========================================
FROM alpine:3.20

# Install ca-certificates and timezone data (useful for logs/scheduling)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the static backend binary
COPY --from=backend-builder /build/gym-rabbit /app/gym-rabbit

# Copy the built frontend static assets into the correct relative path
COPY --from=frontend-builder /build/frontend/dist /app/frontend/dist

# Define volume for persistent database storage
VOLUME ["/app/data"]

# Expose port (default is 8080)
EXPOSE 8080

# Run the backend binary
CMD ["./gym-rabbit"]
