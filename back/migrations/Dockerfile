FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/migrations/go.mod back/migrations/go.sum ./
RUN go mod download

COPY back/migrations/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/migrations ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/migrations /usr/local/bin/migrations

USER app

ENTRYPOINT ["/usr/local/bin/migrations"]
