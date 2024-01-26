FROM golang:1.20 AS BUILDER

ARG version=unknown

WORKDIR $GOPATH/src/github.com/rverst/dyndns

ENV CGO_ENABLED 0

COPY . .

RUN go build -ldflags="-X 'github.com/rverst/dyndns/main.Version=$version'" -o /ddns

FROM alpine

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

WORKDIR /sls
COPY --from=builder /ddns /app/ddns

EXPOSE 80/tcp
VOLUME /data
ENV DYNDNS_PORT 80
ENV DYNDNS_BIND 0.0.0.0
ENV DYNDNS_CONFIG /config.yaml
ENV DYNDNS_DATA /data
ENV DYNDNS_LOGLVL info

CMD ["/app/ddns"]


