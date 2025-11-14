FROM golang:1.24.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server


FROM ubuntu:latest
WORKDIR /
COPY --from=builder /app/server .
CMD ["/server"]