FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod main.go ./
RUN go mod tidy
RUN go build -o sample-service main.go
CMD ["./sample-service"]