FROM --platform=$BUILDPLATFORM golang AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /usr/src/
COPY ./src/app .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 /usr/local/go/bin/go build -o /usr/local/bin/frmcache ./...

FROM scratch
COPY --from=builder /usr/local/bin/frmcache /usr/local/bin/frmcache
COPY ./src/db/ /var/lib/frmcache/
ENTRYPOINT ["/usr/local/bin/frmcache"]
