# syntax=docker/dockerfile:1

FROM golang:alpine

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ozonshrt ./cmd/api

EXPOSE 8080


