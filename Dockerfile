# Use Node.js 24 as the base image
FROM node:24-slim AS builder

# Set the working directory
WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm ci

# Copy the rest of the application code
COPY . .

# Build the SvelteKit app
RUN npm run build

# --- Production Stage ---
FROM node:24-slim AS runner

WORKDIR /app

COPY --from=builder /app/package*.json ./
RUN npm ci --omit=dev

# Copy the built application from the builder stage
COPY --from=builder /app/build ./build

# The database file will be in the root of /app
# You should mount a volume here for persistence:
# docker run -v ./database.sqlite:/app/database.sqlite ...

# Expose the port the app runs on
EXPOSE 3000

# Set environment variables
ENV NODE_ENV=production
ENV PORT=3000

# Start the application
CMD ["node", "build"]
