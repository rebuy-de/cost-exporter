FROM golang:1.12-alpine as builder

RUN apk add --no-cache git make

# Configure Go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

# Install Go Tools
RUN go get -u golang.org/x/lint/golint
RUN go get -u github.com/golang/dep/cmd/dep

COPY . /go/src/github.com/rebuy-de/cost-exporter
WORKDIR /go/src/github.com/rebuy-de/cost-exporter
RUN CGO_ENABLED=0 make install

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/cost-exporter /usr/local/bin/
COPY run.sh /run.sh

RUN chmod +x /run.sh

RUN adduser -D cost-exporter
USER cost-exporter
ENTRYPOINT ["/run.sh"]
