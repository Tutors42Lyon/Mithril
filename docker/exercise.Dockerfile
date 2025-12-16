FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o exercise ./cmd/exercise

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/exercise .
COPY --from=builder /app/exercises ./exercises

CMD ["./exercise"]

