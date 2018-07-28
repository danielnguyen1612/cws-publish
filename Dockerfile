FROM debian:stable

# Install SSL certificates
RUN apt-get update \
 && apt-get install -y --force-yes --no-install-recommends \
      apt-transport-https \
      curl \
      ca-certificates \
 && apt-get clean \
 && apt-get autoremove \
 && rm -rf /var/lib/apt/lists/*

COPY bin/cws-publish /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/cws-publish"]