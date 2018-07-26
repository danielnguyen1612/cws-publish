FROM golang:1.9

COPY bin/cws-publish /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/cws-publish"]