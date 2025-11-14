FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED go build -o /app/server ./cmd/server/main.go

WORKDIR /

COPY  --from=builder /app/server .

CMD ["/server"]