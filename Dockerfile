FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o coachbot ./cmd/coachbot

FROM alpine:latest
RUN apk update && \
    apk upgrade && \
    apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates
WORKDIR /app
COPY --from=builder /app/coachbot .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 80
CMD ["./coachbot"]
