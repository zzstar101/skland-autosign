FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go ./go

WORKDIR /app/go

RUN go build -o /app/skland-attendance ./cmd/skland-attendance

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/skland-attendance /app/skland-attendance

ENV TOKENS=""
ENV NOTIFICATION_URLS=""
ENV MAX_RETRIES="3"

ENTRYPOINT ["/app/skland-attendance"]
CMD ["-mode=once"]
