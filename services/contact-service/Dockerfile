FROM golang:1.24.4-alpine AS builder

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bin/contact-service ./cmd/server

FROM alpine:latest

WORKDIR /

COPY --from=builder bin/contact-service ./cmd/server

CMD ["/bin/contact-service"]