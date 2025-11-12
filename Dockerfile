# Stage 1: Go builder
FROM golang:1.25.3-bookworm AS builder
WORKDIR /server
COPY . /server
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/autodoc main.go

# Stage 2: Node.js builder (optional, to install deps)
FROM node:20-bookworm AS node-builder
WORKDIR /app
COPY node_js/package*.json ./
RUN npm ci --production
COPY node_js/ ./

# Stage 3: Final image
FROM ubuntu:22.04
RUN apt-get update && \
    apt-get install -y jq curl ca-certificates && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs

# Copy Go binary
COPY --from=builder /bin/autodoc /bin/autodoc
# Copy Node.js app + deps
COPY --from=node-builder /app /node_js
COPY ./html-swagger.sh /html-swagger.sh
WORKDIR /
RUN chmod +x /node_js/deref.js

# Default command: Go binary
CMD ["/bin/autodoc"]
