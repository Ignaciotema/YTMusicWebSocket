# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR .

# Instalar certificados (por si el programa hace HTTPS)
RUN apk add --no-cache ca-certificates

# Descargar dependencias primero para aprovechar la caché
COPY go.mod go.sum* ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar el binario
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ytmusic-websocket ./cmd

# ---------- Runtime stage ----------
FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/ytmusic-websocket .

# Puerto (cambialo si usás otro)
EXPOSE 8080

CMD ["./ytmusic-websocket"]