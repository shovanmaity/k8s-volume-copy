FROM alpine:3.12
RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.0735-r0

# entrypoint script
COPY docker/populator/rsync/entrypoint.sh /usr/sbin/entrypoint.sh
RUN chmod +x /usr/sbin/entrypoint.sh
RUN mkdir -p /entrypoint.d
COPY bin/rsyncp /usr/sbin/rsyncp

ENTRYPOINT [ "entrypoint.sh" ]
CMD [ "rsyncp" ]
