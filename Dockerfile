# Build stage
FROM golang:1.22.1-alpine3.19 AS builder
LABEL authors="the-eduardo"
WORKDIR /app
COPY . .
RUN go build -o main .
RUN apk add --no-cache curl

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz

# Final stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main /app/main
COPY --from=builder /app/migrate ./migrate
COPY db/migration ./db/migration
COPY app.env ./app.env
COPY wait-for.sh ./wait-for.sh
COPY start.sh ./start.sh


EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]