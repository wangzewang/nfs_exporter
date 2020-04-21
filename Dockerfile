FROM alpine:latest

# install nfs-utils
RUN apk update && apk add --update nfs-utils coreutils && rm -rf /var/cache/apk/*

RUN rm /sbin/halt /sbin/poweroff /sbin/reboot

#RUN unset http_proxy

# start nfs_exporter
ADD nfs_exporter /usr/local/bin/nfs_exporter

RUN chmod 775 /usr/local/bin/nfs_exporter

EXPOSE 9102

ENTRYPOINT  [ "/usr/local/bin/nfs_exporter" ]
