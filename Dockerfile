# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

# Final stage
FROM ubuntu:latest
WORKDIR /
COPY --from=builder /app/server .
CMD ["/server"]