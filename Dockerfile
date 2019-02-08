FROM golang:1.11-alpine as buildgo

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOMETALINTER=2.0.11

ARG VERSION

RUN \
    apk add --no-cache --update tzdata git bash curl && \
    cp /usr/share/zoneinfo/America/New_York /etc/localtime && \
    rm -rf /var/cache/apk/*


FROM alpine:3.8 as build-base

ENV \
TERM=xterm-color \
MYUSER=app \
MYUID=1001 \
DOCKER_GID=999

RUN apk add --no-cache --update su-exec tzdata curl ca-certificates && \
mkdir -p /home/$MYUSER && \
adduser -s /bin/sh -D -u $MYUID $MYUSER && chown -R $MYUSER:$MYUSER /home/$MYUSER && \
mkdir -p /srv && chown -R $MYUSER:$MYUSER /srv && \
rm -rf /var/cache/apk/*

FROM buildgo as build-backend

ADD . /go/src/mock.bill.com/

WORKDIR /go/src/mock.bill.com/
RUN go build -o mockserver -ldflags "-X main.version=$VERSION -s -w" *.go


FROM build-base
WORKDIR /srv

COPY --from=build-backend /go/src/mock.bill.com/mockserver /srv/

ADD start.sh /srv/start.sh
RUN chmod +x /srv/*.sh
RUN chown -R app:app /srv
RUN ln -s /srv/mockserver /usr/bin/mockserver

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/ping || exit 1

ENTRYPOINT ["/srv/start.sh"]