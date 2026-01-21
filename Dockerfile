# Build stage
FROM golang:1.25-alpine AS builder

# Install tzdata for timezone support
RUN apk --no-cache add tzdata

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o qis-note-notifier .

# Final stage
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/qis-note-notifier /qis-note-notifier

USER 1000

ENTRYPOINT ["/qis-note-notifier"]