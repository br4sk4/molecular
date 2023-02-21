ARG BASE_IMAGE=alpine

FROM golang:1.20.0-alpine3.16 AS go-builder

WORKDIR /usr/local/src/molecular

RUN apk add --no-cache --update alpine-sdk ca-certificates openssl

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT}

ARG GOPROXY

COPY . .

RUN go build -o ./.dist/molecular molecular


FROM node:19.6-alpine3.16 AS node-builder

WORKDIR /usr/local/src/molecular/

RUN apk add --no-cache --update alpine-sdk ca-certificates openssl

COPY . .

WORKDIR /usr/local/src/molecular/login

RUN yarn install
RUN yarn run build


FROM $BASE_IMAGE

# Dex connectors, such as GitHub and Google logins require root certificates.
# Proper installations should manage those certificates, but it's a bad user
# experience when this doesn't work out of the box.
#
# See https://go.dev/src/crypto/x509/root_linux.go for Go root CA bundle locations.
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy module files for CVE scanning / dependency analysis.
COPY --from=go-builder /usr/local/src/molecular/go.mod /usr/local/src/molecular/go.sum /usr/local/src/molecular/
COPY --from=node-builder /usr/local/src/molecular/login/package.json /usr/local/src/molecular/login/yarn.lock /usr/local/src/molecular/

# Copy distribution files
COPY --from=go-builder /usr/local/src/molecular/.dist/molecular /opt/molecular/molecular
COPY --from=node-builder /usr/local/src/molecular/.dist/frontend /opt/molecular/frontend

# Copy configuration files
COPY --from=go-builder /usr/local/src/molecular/.conf/config.docker.yaml /etc/molecular/config.yaml

RUN chown 1001:1001 /etc/molecular
RUN ln -s /opt/molecular/molecular /usr/local/bin/molecular

USER 1001:1001

CMD ["molecular", "--config", "/etc/molecular/config.yaml"]