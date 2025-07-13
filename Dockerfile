FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o books-api ./cmd/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/books-api .
CMD ["./books-api"] 