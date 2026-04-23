FROM golang:1.26-alpine AS builder
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cullarr ./cmd/cullarr

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/cullarr /usr/local/bin/cullarr
ENTRYPOINT ["/usr/local/bin/cullarr"]
