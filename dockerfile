# Usa una imagen oficial con Go para construir el binario
FROM golang:1.21-alpine AS builder

# Directorio de trabajo
WORKDIR /app

# Copia los archivos del proyecto
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Construye la aplicación en modo producción, generando un binario
RUN go build -o bot-telegram main.go

# Usa una imagen pequeña para ejecutar el binario
FROM alpine:latest

# Instala certificados SSL locales (recomendado para HTTPS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia el binario compilado desde la etapa builder
COPY --from=builder /app/bot-telegram .

# Expone el puerto que espera Gin (por defecto 8080)
EXPOSE 8080 

# Define la variable de entorno PORT que Railway usa para definir el puerto (opcional si usas otro)
ENV PORT 8080

# Comando para iniciar la app
CMD ["./bot-telegram"]
