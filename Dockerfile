FROM golang:alpine AS builder
WORKDIR /build
ENV CGO_ENABLED=0
ENV GOOS=linux
ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o watch_bot .

FROM alpine:latest
WORKDIR /build
COPY --from=builder /build/watch_bot /build/watch_bot
ENTRYPOINT ["./watch_bot"]