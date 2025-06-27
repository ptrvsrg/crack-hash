ARG GOLANG_VERSION=1.24.2-alpine3.21
ARG ALPINE_VERSION=3.21

FROM golang:${GOLANG_VERSION} AS deps

WORKDIR /app

ARG REGISTRY_HOST
ARG REGISTRY_USER
ARG REGISTRY_PASSWORD

COPY ./commonlib ./commonlib
COPY ./manager ./manager
COPY ./worker ./worker
COPY ./worker/.netrc.tmpl /root

WORKDIR /app/worker

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPRIVATE=github.com/*

RUN apk --no-cache update \
    && apk add --no-cache --upgrade \
        gcc \
        git \
        musl-dev \
        perl
RUN go generate ./...
RUN go mod download

FROM deps AS build

WORKDIR /app/worker

ENV CGO_ENABLED=0
ARG ARTIFACT_VERSION

RUN go build \
    -o ./bin/worker \
    -installsuffix "static" \
    -tags "" \
    -ldflags " \
    	-X github.com/ptrvsrg/crack-hash/worker/internal/version.AppVersion=${ARTIFACT_VERSION:-0.0.0} \
    	-X github.com/ptrvsrg/crack-hash/worker/internal/version.GoVersion=$(go version | cut -d " " -f 3) \
    	-X github.com/ptrvsrg/crack-hash/worker/internal/version.Platform=$(go env GOOS)/$(go env GOARCH)" \
    ./cmd/cli

FROM alpine:${ALPINE_VERSION} AS runtime

WORKDIR /app

COPY --from=build /app/worker/bin /app
COPY --from=build /app/worker/config/config.default.yaml config/config.yaml

RUN apk update \
    && apk add --no-cache --upgrade \
        bash \
        ca-certificates \
        curl \
        tzdata \
    && update-ca-certificates \
    && echo 'Etc/UTC' > /etc/timezone \
    && adduser --disabled-password --home /app --gecos '' gouser \
    && chown -R gouser /app

ENV TZ     :/etc/localtime
ENV LANG   en_US.utf8
ENV LC_ALL en_US.UTF-8

USER gouser

ENTRYPOINT [ "/app/worker", "server" ]