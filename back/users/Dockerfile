FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/access ./access
COPY back/users/go.mod back/users/go.sum ./users/
WORKDIR /src/users
RUN go mod download

COPY back/users/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/users ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/users /usr/local/bin/users

USER app
EXPOSE 8083

ENTRYPOINT ["/usr/local/bin/users"]
