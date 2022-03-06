FROM golang:1.17 as builder
WORKDIR /app
COPY src/ ./
RUN go mod download
RUN go build -a -installsuffix cgo -o app .

FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /app/app ./
CMD ["./app"]