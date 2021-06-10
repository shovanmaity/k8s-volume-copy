FROM alpine:3.12
RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.0735-r0
RUN apk add --no-cache rsync==3.1.3-r3
RUN apk add --no-cache openssh-client==8.3_p1-r2

# entrypoint script
ADD package/client/entrypoint.sh /usr/sbin/entrypoint.sh
RUN chmod +x /usr/sbin/entrypoint.sh
RUN mkdir -p /entrypoint.d

ENTRYPOINT [ "entrypoint.sh" ]
