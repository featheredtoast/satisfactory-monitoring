FROM golang AS builder
WORKDIR /usr/src
COPY ./src/app .
ENV GOOS=linux GARCH=amd64 CGO_ENABLED=0
RUN /usr/local/go/bin/go build -o /usr/local/bin/alertmanager-config main.go

FROM scratch
ENV OUTPUT_PATH=/etc/alertmanager/alertmanager.yml \
INPUT_PATH=/var/lib/alertmanager-config/templates/alertmanager.yml.tmpl

COPY --from=builder /usr/local/bin/alertmanager-config /usr/local/bin/alertmanager-config
COPY ./src/templates/alertmanager.yml.tmpl /var/lib/alertmanager-config/templates/alertmanager.yml.tmpl
ENTRYPOINT ["/usr/local/bin/alertmanager-config"]