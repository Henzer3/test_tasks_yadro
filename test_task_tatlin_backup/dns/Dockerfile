FROM golang:1.25 AS builder

WORKDIR /app

COPY go.work ./
COPY dns/go.mod dns/go.sum ./dns/
COPY pkg/go.mod pkg/go.sum ./pkg/

COPY . .

WORKDIR /app/dns
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/dns ./cmd/app

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /bin/dns /app/dns

CMD ["./dns"]