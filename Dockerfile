FROM quay.io/rebuy/rebuy-go-sdk:v3.7.0 as builder

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
