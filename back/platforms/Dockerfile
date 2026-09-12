FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/access ./access
COPY back/products ./products
COPY back/platforms/go.mod back/platforms/go.sum ./platforms/
WORKDIR /src/platforms
RUN go mod download

COPY back/platforms/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/platforms ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/platforms /usr/local/bin/platforms

USER app
EXPOSE 8086

ENTRYPOINT ["/usr/local/bin/platforms"]
