FROM golang:1.24.3-alpine AS builder

ARG SERVICE

WORKDIR /app

COPY go.work go.work.sum* ./
COPY pkg/ ./pkg/
COPY scripts/ ./scripts/
COPY services/ ./services/

RUN go build -o /bin/${SERVICE} ./services/${SERVICE}/cmd/server

FROM alpine:latest

ARG SERVICE

COPY --from=builder /bin/${SERVICE} /bin/server

CMD ["/bin/server"]
