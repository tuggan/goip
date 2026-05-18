FROM golang:1.25-alpine3.21 AS build

RUN apk add --update git make

WORKDIR /src
COPY . .

RUN make

FROM alpine:3.21

LABEL maintainer="dennis@vestern.se"

ENV GOIP_ROOT=/srv/goip/
ENV GOIP_CONFIG_ROOT=/etc/goip/
ENV GOIP_USER=goip
ENV GOIP_GROUP=${GOIP_USER}

RUN mkdir -p ${GOIP_ROOT} ${GOIP_CONFIG_ROOT}

WORKDIR ${GOIP_ROOT}

RUN addgroup -S ${GOIP_GROUP} && adduser -S -G ${GOIP_GROUP} -H -D -g "" ${GOIP_USER}

COPY --from=build /src/build/goip /usr/local/bin/
COPY html /srv/goip/html/
COPY config/goip.toml /etc/goip/

RUN chmod 755 /usr/local/bin/goip

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1

USER ${GOIP_USER}

ENTRYPOINT ["goip"]
