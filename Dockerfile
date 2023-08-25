FROM golang:1.21-alpine as builder

RUN apk add --no-cache git openssl

ENV CGO_ENABLED=0
RUN go install golang.org/x/lint/golint@latest

COPY . /build
RUN cd /build && ./buildutil

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime && \
    echo "Europe/Berlin" > /etc/timezone && \
    apk del tzdata

COPY --from=builder /build/dist/cost-exporter /usr/local/bin/
COPY run.sh /run.sh

RUN adduser -D cost-exporter
USER cost-exporter

ENTRYPOINT ["/run.sh"]
