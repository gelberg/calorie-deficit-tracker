# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
COPY ./ ./
COPY *.json ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /google-fit-producer

CMD [ "/google-fit-producer" ]