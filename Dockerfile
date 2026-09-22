# compila dentro de uma imagem com go, assim nao precisa do go instalado na maquina.
FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -o finops .

# imagem final so com o binario e o front, sem o go.
FROM alpine:3.22

WORKDIR /app

COPY --from=build /app/finops .
COPY public ./public

EXPOSE 8080

CMD ["./finops"]
