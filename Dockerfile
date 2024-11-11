FROM golang:1.22 AS builder
WORKDIR /app
COPY src/ ./
RUN go mod download
RUN go build -a -installsuffix cgo -o app .

FROM archlinux:latest
RUN pacman -Syyu --needed --noconfirm
RUN pacman-key --init
RUN pacman -S --noconfirm ansible openssh
WORKDIR /root/
COPY --from=builder /app/app ./
CMD ["./app"]
