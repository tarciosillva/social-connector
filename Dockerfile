FROM golang:1.23.4-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o social-connector .

FROM alpine:3.16 AS binary

WORKDIR /app

COPY --from=builder /app/social-connector .

EXPOSE 8001

RUN apk add --no-cache tzdata && \
    chmod +x /app/social-connector

CMD ["/app/social-connector"]
