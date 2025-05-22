FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/main.go

FROM alpine:latest
WORKDIR /app

# Copy the binary and frontend files
COPY --from=builder /main /app/main
COPY frontend/ /app/frontend/

# Set default environment variables that can be overridden at runtime
ENV DB_HOST=dpg-d0nocoqdbo4c73agt8c0-a.oregon-postgres.render.com \
    DB_PORT=5432 \
    DB_USER=gastro_trp4_user \
    DB_PASSWORD=JY6zmcWkUWwwa7idXUoOxTjLKD2V6ZER \
    DB_NAME=gastro_trp4 \
    DB_SSL_MODE=require \
    SERVER_PORT=10000 \
    JWT_KEY=your-secret-key \
    PROJECT_ROOT=. \
    FRONTEND_PATH=frontend

# Expose the port
EXPOSE 10000

# Run the application
CMD ["/app/main"]