FROM golang:1.25 AS builder

WORKDIR /app

COPY server/go.mod server/go.sum ./server/

COPY . .

WORKDIR /app/server
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/log ./cmd/app

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /bin/log /app/log

CMD ["./log"]