# ---------- Estágio de build ----------
FROM golang:1.26-alpine AS build

WORKDIR /src

# As dependências vêm primeiro para aproveitar o cache de camadas do Docker.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Binário estático: permite uma imagem final mínima, sem toolchain Go.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/api ./cmd/api

# ---------- Imagem final ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget && \
    adduser -D -u 10001 api

COPY --from=build /bin/api /usr/local/bin/api

# Sem isto o Gin imprime avisos de modo debug a cada start nos logs do Pod.
ENV GIN_MODE=release

USER api
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
