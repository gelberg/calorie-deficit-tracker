# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app
COPY calorie_deficit_calculator/go.mod calorie_deficit_calculator/go.sum ./
COPY common ../common

RUN go mod download
COPY calorie_deficit_calculator/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /calorie-deficit-calculator

CMD [ "/calorie-deficit-calculator" ]