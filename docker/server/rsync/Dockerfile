FROM alpine:3.12
LABEL org.opencontainers.image.source https://github.com/shovanmaity/k8s-volume-copy
RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.0735-r0
RUN apk add --no-cache rsync==3.1.3-r3
COPY docker/server/rsync/rsyncd.conf /etc/rsyncd.conf

# entrypoint script
COPY docker/server/rsync/entrypoint.sh /usr/sbin/entrypoint.sh
RUN chmod +x /usr/sbin/entrypoint.sh
RUN mkdir -p /entrypoint.d
VOLUME /data
ENTRYPOINT [ "entrypoint.sh" ] 

EXPOSE 873
CMD [ "/usr/bin/rsync", "--no-detach", "--daemon", "--log-file=/dev/stdout" ]
