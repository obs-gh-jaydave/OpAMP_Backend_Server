FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o opamp-server ./cmd/server.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/opamp-server .
CMD ["./opamp-server"]