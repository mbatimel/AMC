FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/access ./access
COPY back/auth ./auth
COPY back/admin/go.mod back/admin/go.sum ./admin/
WORKDIR /src/admin
RUN go mod download

COPY back/admin/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/admin ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/admin /usr/local/bin/admin

USER app
EXPOSE 8084

ENTRYPOINT ["/usr/local/bin/admin"]
