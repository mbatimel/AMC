FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/integrations/go.mod back/integrations/go.sum ./integrations/
COPY back/orders/go.mod back/orders/go.sum ./orders/
COPY back/access/go.mod back/access/go.sum ./access/
WORKDIR /src/integrations
RUN go mod download

COPY back/integrations/ ./
COPY back/orders/ ../orders/
COPY back/access/ ../access/

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/onec-sync ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/onec-sync /usr/local/bin/onec-sync

USER app
EXPOSE 9096

ENTRYPOINT ["/usr/local/bin/onec-sync"]
