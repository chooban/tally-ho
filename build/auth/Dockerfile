ARG GO_VERSION=1.24

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-bookworm AS builder_buildplatform

ARG TAG=master
ENV CGO_ENABLED=1
RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      build-essential \
      libsqlite3-dev \
      libsqlite3-0 \
      ca-certificates

COPY localdev/caddy/caddy/pki/authorities/tallyho/root.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /usr/src/app

# I could just clone the repo, but...
RUN mkdir relme && \
    cd relme && \
    git init && \
    git config core.sparseCheckout true && \
    git remote add origin -f https://github.com/hawx/relme-auth.git && \
    echo "web" >> .git/info/sparse-checkout && \
    git pull origin master

#RUN git clone --depth 1 --branch ${TAG} https://github.com/hawx/relme-auth.git
#WORKDIR /usr/src/app/relme-auth

RUN go install hawx.me/code/relme-auth@latest

CMD ["relme-auth", "--cookie-secret", "dGhpcyBpcyBhIHNlY3JldA==", "--true", "--web-path", "./relme/web/", "--config", "/config/conf.toml", "--base-url", "https://auth.tallyho.test", "--db", "/data/database.sqlite"]
