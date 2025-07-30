FROM --platform=linux/amd64  node:lts

ARG TAG=v0.15.0
WORKDIR /usr/src/app
EXPOSE 8888
#RUN apk update && apk upgrade --no-cache && \
#    apk add --no-cache bash git openssh
#
RUN git clone --depth 1 --branch ${TAG} https://github.com/benjifs/sparkles.git

WORKDIR /usr/src/app/sparkles

RUN npm install -g netlify-cli@v22.3.0 && \
    npm install -g deno && \
    npm install

CMD ["netlify", "dev"]