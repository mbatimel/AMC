FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/access ./access
COPY back/objectstorage ./objectstorage
COPY back/auth/go.mod back/auth/go.sum ./auth/
WORKDIR /src/auth
RUN go mod download

COPY back/auth/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/auth ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/auth /usr/local/bin/auth

USER app
EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/auth"]
