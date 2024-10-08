FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo ./cmd/coachbot


FROM scratch
WORKDIR /app
COPY --from=builder /app/coachbot .
EXPOSE 80
CMD ["./coachbot"]
