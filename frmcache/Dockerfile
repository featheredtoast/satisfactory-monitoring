FROM golang AS builder
WORKDIR /usr/src/
COPY ./src/app .
ENV GOOS=linux GARCH=amd64 CGO_ENABLED=0
RUN /usr/local/go/bin/go build -o /usr/local/bin/frmcache ./...

FROM scratch
COPY --from=builder /usr/local/bin/frmcache /usr/local/bin/frmcache
COPY ./src/db/ /var/lib/frmcache/
ENTRYPOINT ["/usr/local/bin/frmcache"]
