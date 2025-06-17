ARG GO_VERSION=1.24
FROM --platform=linux/amd64 golang:${GO_VERSION}-bookworm AS builder_amd64
ENV CGO_ENABLED=1
RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      build-essential \
      libsqlite3-dev
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /tallyho .

RUN go build -v -o /tallyho .


FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-bookworm AS builder_buildplatform

ENV CGO_ENABLED=1
RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      build-essential \
      libsqlite3-dev
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${BUILDPLATFORM} go build -a -installsuffix cgo -o /tallyho .


FROM --platform=linux/amd64 golang:${GO_VERSION}-bookworm AS runtime_amd64
RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      libsqlite3-0
WORKDIR /app
COPY --from=builder_amd64 /tallyho ./
COPY --from=builder_amd64 /etc/passwd /etc/passwd
COPY --from=builder_amd64 /etc/group /etc/group

ENV MEDIA_DIR=/media
ENV WEB_DIR=/web
ENV DB=file::memory

#USER myuser:myuser
CMD ./tallyho \
    --media-dir ${MEDIA_DIR} \
    --web ${WEB_DIR} \
    --db ${DB}


FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-bookworm AS runtime_buildplatform
RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      libsqlite3-0
WORKDIR /app
COPY --from=builder_buildplatform /tallyho ./
COPY --from=builder_buildplatform /etc/passwd /etc/passwd
COPY --from=builder_buildplatform /etc/group /etc/group

ENV MEDIA_DIR=/media
ENV WEB_DIR=/web
ENV DB=file::memory

#USER myuser:myuser
CMD ./tallyho \
    --media-dir ${MEDIA_DIR} \
    --web ${WEB_DIR} \
    --db ${DB}
