FROM alpine:latest

RUN apk add --no-cache dnsmasq
#RUN adduser -D -s /bin/sh dnsmasq
RUN mkdir -p /var/lib/dnsmasq

EXPOSE 53/udp 53/tcp

USER dnsmasq
CMD ["dnsmasq", "--no-daemon", "--user=dnsmasq", "--group=dnsmasq"]