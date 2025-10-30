FROM golang:1.24.4-bookworm AS builder
COPY . /server
WORKDIR /server
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/autodoc main.go

FROM node:latest

RUN npm -g install @scalar/cli
COPY --from=builder /bin/autodoc /bin/autodoc

CMD ["/bin/autodoc"]
