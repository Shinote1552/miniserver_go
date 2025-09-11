FROM golang:1.23.0-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o urlshortener ./cmd/server

FROM alpine:3.19.1

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /app/tmp /data && \
    chmod 755 /app && \
    chmod 755 /app/tmp && \
    chmod 755 /data

COPY --from=builder /app/urlshortener .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN chown -R appuser:appgroup /app/tmp /data

USER appuser

EXPOSE 8080

CMD ["./urlshortener"]