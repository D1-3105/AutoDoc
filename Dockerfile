FROM golang:1.25.3-bookworm AS builder
COPY . /server
WORKDIR /server
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/autodoc main.go

FROM ubuntu:22.04
COPY ./html-swagger.sh /html-swagger.sh
COPY --from=builder /bin/autodoc /bin/autodoc
RUN apt-get update && \
    apt-get install -y jq curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

CMD ["/bin/autodoc"]

