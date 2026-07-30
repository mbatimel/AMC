FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/access ./access
COPY back/products/go.mod back/products/go.sum ./products/
WORKDIR /src/products
RUN go mod download

COPY back/products/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/products ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/products /usr/local/bin/products

USER app
EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/products"]
