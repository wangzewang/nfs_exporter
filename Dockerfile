FROM alpine:latest

RUN apk update \
    && apk add -u coreutils bash\
    && rm -rf /var/lib/apt/lists/*

COPY nfs_exporter /bin

RUN chmod +x /bin/nfs_exporter

EXPOSE 9102

ENTRYPOINT ["/bin/nfs_exporter"]
