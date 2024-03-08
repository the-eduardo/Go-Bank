# Build stage
FROM golang:1.22.1-alpine3.19 AS builder
LABEL authors="the-eduardo"
WORKDIR /app
COPY . .
RUN go build -o main .

# Final stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main /app/main
COPY app.env /app/app.env

EXPOSE 8080
CMD ["/app/main"]