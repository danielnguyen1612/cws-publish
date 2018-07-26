FROM golang:1.9

COPY /go/bin/cws-publish /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/cws-publish"]