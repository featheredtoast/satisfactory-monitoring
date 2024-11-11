FROM --platform=$BUILDPLATFORM golang AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /usr/src
COPY ./src/app .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 /usr/local/go/bin/go build -o /usr/local/bin/alertmanager-config main.go

FROM scratch
ENV OUTPUT_PATH=/etc/alertmanager/alertmanager.yml \
INPUT_PATH=/var/lib/alertmanager-config/templates/alertmanager.yml.tmpl

COPY --from=builder /usr/local/bin/alertmanager-config /usr/local/bin/alertmanager-config
COPY ./src/templates/alertmanager.yml.tmpl /var/lib/alertmanager-config/templates/alertmanager.yml.tmpl
ENTRYPOINT ["/usr/local/bin/alertmanager-config"]
