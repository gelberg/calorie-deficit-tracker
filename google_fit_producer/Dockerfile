# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app
COPY google_fit_producer/go.mod google_fit_producer/go.sum ./
COPY common ../common

RUN go mod download
COPY google_fit_producer/ ./
COPY *.json ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /google-fit-producer

CMD [ "/google-fit-producer" ]