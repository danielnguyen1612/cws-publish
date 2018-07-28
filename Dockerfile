FROM alpine:3.8

COPY bin/cws-publish /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/cws-publish"]