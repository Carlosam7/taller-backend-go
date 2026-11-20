# Etapa 1: compilación
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copiamos primero los archivos de dependencias
COPY go.mod go.sum ./

RUN go mod download

# Copiamos el código fuente
COPY . .

# Compilamos la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/api ./cmd/api

# Etapa 2: ejecución
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

# Copiamos únicamente el ejecutable
COPY --from=builder /app/api ./api

EXPOSE 3000

CMD ["./api"]
