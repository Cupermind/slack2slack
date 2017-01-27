FROM alpine:3.4
MAINTAINER Alexander Svyrydov

ADD slack2slack /

RUN apk update || true && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/*

EXPOSE 8888
WORKDIR /
CMD /slack2slack
