FROM alpine:3.12
RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.0735-r0

# entrypoint script
COPY docker/populator/claim/entrypoint.sh /usr/sbin/entrypoint.sh
RUN chmod +x /usr/sbin/entrypoint.sh
RUN mkdir -p /entrypoint.d
COPY bin/pv-populator /usr/sbin/pv-populator

ENTRYPOINT [ "entrypoint.sh" ]
CMD [ "pv-populator" ]
