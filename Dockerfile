FROM golang:1.21-alpine AS builder

RUN apk add --no-cache build-base sqlite-dev
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o kvstore

FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs
WORKDIR /app

COPY --from=builder /app/kvstore .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

VOLUME /app/data
EXPOSE 8080

CMD ["/app/kvstore"]