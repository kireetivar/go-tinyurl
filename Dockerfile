FROM golang:1.24.2 AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server


FROM ubuntu:latest
WORKDIR /
COPY --from=builder /app/server .
CMD ["/server"]