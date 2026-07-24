FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/access ./access
COPY back/orders/go.mod back/orders/go.sum ./orders/
WORKDIR /src/orders
RUN go mod download

COPY back/orders/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/orders ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/orders /usr/local/bin/orders

USER app
EXPOSE 8082

ENTRYPOINT ["/usr/local/bin/orders"]
