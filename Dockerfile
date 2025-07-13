FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
COPY docs ./docs
RUN mkdir -p ./swagger && cp ./docs/swagger.json ./swagger/doc.json
RUN go build -o books-api ./cmd/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/books-api .
CMD ["./books-api"] 