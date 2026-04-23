FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cullarr .

FROM alpine:3
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/cullarr /usr/local/bin/cullarr
ENTRYPOINT ["cullarr"]
