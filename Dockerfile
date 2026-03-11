FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# Build only the HTTP server (cmd/game is WebAssembly and must be built separately with GOOS=js GOARCH=wasm)
RUN go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./server"]
