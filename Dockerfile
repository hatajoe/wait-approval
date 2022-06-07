FROM golang:1.18 AS builder
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o app .

FROM alpine:3.14
LABEL org.opencontainers.image.source https://github.com/hatajoe/wait-approval
RUN apk update && apk add ca-certificates
COPY --from=builder /app/app /app/app
CMD ["/app/app"]
