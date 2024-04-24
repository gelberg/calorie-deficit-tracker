# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app
COPY fatsecret_producer/go.mod fatsecret_producer/go.sum ./
COPY common ../common

RUN go mod download
COPY fatsecret_producer/ ./
COPY *.json ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /fatsecret-producer

CMD [ "/fatsecret-producer" ]