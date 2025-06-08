FROM golang:1.23-alpine AS builder
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

# Environment variables (will be populated from .env file or runtime)
ENV DB_HOST=${DB_HOST} \
    DB_PORT=${DB_PORT} \
    DB_USER=${DB_USER} \
    DB_PASSWORD=${DB_PASSWORD} \
    DB_NAME=${DB_NAME} \
    DB_SSL_MODE=${DB_SSL_MODE} \
    SERVER_PORT=${SERVER_PORT} \
    JWT_KEY=${JWT_KEY} \
    PROJECT_ROOT=${PROJECT_ROOT} \
    FRONTEND_PATH=${FRONTEND_PATH} \
    GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID} \
    GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET} \
    GOOGLE_REDIRECT_URL=${GOOGLE_REDIRECT_URL} \
    SMTP_HOST=${SMTP_HOST} \
    SMTP_PORT=${SMTP_PORT} \
    SMTP_USERNAME=${SMTP_USERNAME} \
    SMTP_PASSWORD=${SMTP_PASSWORD} \
    SMTP_FROM=${SMTP_FROM}

# Expose the port
EXPOSE 10000

# Run the application
CMD ["/app/main"]