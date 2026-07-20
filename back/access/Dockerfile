FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/access/go.mod back/access/go.sum ./
RUN go mod download

COPY back/access/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/access ./cmd/access-internal

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/access /usr/local/bin/access

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/access"]
