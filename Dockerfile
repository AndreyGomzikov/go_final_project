FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o todo-app .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/todo-app .
COPY --from=builder /app/web ./web

VOLUME /data

CMD ["./todo-app"]
