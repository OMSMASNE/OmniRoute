# Stage 1: Build the binary.
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o load_balancer cmd/load_balancer/main.go

# Stage 2: Final lightweight image.
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/load_balancer .
EXPOSE 8080
CMD ["./load_balancer"]
