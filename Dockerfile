# --- ETAPA 1: Construcción (Builder) ---
# Usamos la imagen oficial de Go basada en Alpine (muy ligera)
FROM golang:1.23-alpine AS builder
# Establecemos el directorio de trabajo dentro del contenedor
WORKDIR /app

# 1. Copiamos primero los archivos de dependencias (para aprovechar la caché de Docker)
COPY go.mod go.sum ./
RUN go mod download

# 2. Copiamos el resto del código fuente
COPY . .

# 3. Compilamos el binario
# IMPORTANTE: CGO_ENABLED=0 es necesario para que funcione con 'glebarez/sqlite'
# y para que el binario funcione en la imagen ligera 'alpine'
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ms3-service main.go

# --- ETAPA 2: Ejecución (Runtime) ---
# Usamos una imagen vacía y ligera de Alpine Linux
FROM alpine:latest  

WORKDIR /root/

# Copiamos SOLO el binario compilado desde la etapa anterior
COPY --from=builder /app/ms3-service .

# (Opcional) Si tu ms3 escucha en algún puerto HTTP, exponlo aquí
# EXPOSE 3003 

# Comando para iniciar el microservicio
CMD ["./ms3-service"]