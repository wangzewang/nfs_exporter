FROM golang:latest

COPY nfs_exporter /bin/

RUN chmod 775 /bin/nfs_exporter

EXPOSE 9102

ENTRYPOINT ["/bin/nfs_exporter"]
