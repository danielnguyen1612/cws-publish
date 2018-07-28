FROM debian:stable

COPY bin/cws-publish /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/cws-publish"]