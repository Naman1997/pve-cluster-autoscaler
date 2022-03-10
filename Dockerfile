FROM golang:1.17 as builder
WORKDIR /app
COPY src/ ./
RUN go mod download
RUN go build -a -installsuffix cgo -o app .

FROM debian:buster-slim
ENV DEBIAN_FRONTEND=noninteractive
RUN set -x && apt-get update && apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
RUN apt-get update
RUN apt-get -y install python3 python3-nacl python3-pip libffi-dev openssh-client
RUN pip3 install ansible
WORKDIR /root/
COPY --from=builder /app/app ./
CMD ["./app"]