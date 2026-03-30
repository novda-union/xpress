FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/server cmd/server/main.go && \
    go build -o bin/migrate cmd/migrate/main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/bin/server .
COPY --from=builder /app/bin/migrate .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
