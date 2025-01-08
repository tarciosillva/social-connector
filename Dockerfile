FROM golang:1.23.4-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o social-connector .

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/social-connector .

EXPOSE 8001

CMD ["./social-connector"]
