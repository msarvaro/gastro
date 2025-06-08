# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY backend/ ./backend/
WORKDIR /app/backend
RUN go mod tidy
RUN go build -o server ./cmd/server

# Run stage
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/backend/server .
COPY --from=builder /app/backend/configs ./configs
COPY --from=builder /app/backend/migrations ./migrations
EXPOSE 8080
CMD ["./server"]