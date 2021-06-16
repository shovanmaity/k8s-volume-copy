FROM alpine:3.12
RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.0735-r0
RUN apk add --no-cache openssh-server==8.3_p1-r2

# entrypoint script
COPY docker/server/ssh/entrypoint.sh /usr/sbin/entrypoint.sh
RUN chmod +x /usr/sbin/entrypoint.sh
RUN mkdir -p /entrypoint.d
VOLUME /data
ENTRYPOINT [ "entrypoint.sh" ] 
