# Stage 1: Build the binary.
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend cmd/backend/main.go

# Stage 2: Final lightweight image.
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/backend .
EXPOSE 8081
CMD ["./backend"]
