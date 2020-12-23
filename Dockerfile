### Because of some quirks with alpine the binary has to be built inside alpine
FROM golang:1.15-alpine3.12 AS build

RUN apk add --update git make

RUN mkdir -p /go/src/github.com/tuggan/goip

COPY . /go/src/github.com/tuggan/goip/

WORKDIR /go/src/github.com/tuggan/goip

RUN make



### Bulid the actual container image
FROM alpine:3.12

LABEL maintainer="dennisvesterlund@gmail.com"

ENV GOIP_ROOT=/srv/goip/
ENV GOIP_CONFIG_ROOT=/etc/goip/
ENV GOIP_USER=goip
ENV GOIP_GROUP=${GOIP_USER}

RUN mkdir -p ${GOIP_ROOT} ${GOIP_CONFIG_ROOT} /srv/goip

WORKDIR ${GOIP_ROOT}

RUN addgroup -S ${GOIP_GROUP} && adduser -S -G ${GOIP_GROUP} -H -D -g "" ${GOIP_USER}

COPY --from=build /go/src/github.com/tuggan/goip/goip /usr/bin/
COPY html /srv/goip/html/
COPY goip.toml /etc/goip/

RUN chmod 755 /usr/local/bin/goip

EXPOSE 3000

USER ${GOIP_USER}

ENTRYPOINT ["goip"]
