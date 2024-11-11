FROM golang AS builder
WORKDIR /usr/src/
ENV GOOS=linux GARCH=amd64 CGO_ENABLED=0
RUN /usr/bin/git clone --depth 1 -b featheredtoast-main https://github.com/featheredtoast/FicsitRemoteMonitoringCompanion.git . &&\
  cd Companion &&\
  /usr/local/go/bin/go build -o bin/companion main.go

FROM scratch
COPY --from=builder /usr/src/Companion/bin/companion /usr/local/bin/companion
EXPOSE 9000
ENV FRM_LOG_STDOUT=1
ENTRYPOINT ["/usr/local/bin/companion"]
