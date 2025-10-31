FROM golang:1.25.3-bookworm AS builder
COPY . /server
WORKDIR /server
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/autodoc main.go

FROM ubuntu:22.04
COPY ./html-swagger.sh /html-swagger.sh

