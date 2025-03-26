FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o firewall-updater main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/firewall-updater .
CMD ["./firewall-updater"]