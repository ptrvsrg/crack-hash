FROM golang:1.24.2-alpine3.21 AS deps

WORKDIR /app

COPY ./commonlib ./commonlib
COPY ./manager ./manager
COPY ./worker ./worker

WORKDIR /app/manager

ENV GO111MODULE=on

RUN go generate ./...
RUN go mod download

FROM deps AS build

WORKDIR /app/manager

ENV CGO_ENABLED=0
ARG ARTIFACT_VERSION

RUN go build \
    -o ./bin/manager \
    -installsuffix "static" \
    -tags "" \
    -ldflags " \
    	-X github.com/ptrvsrg/crack-hash/manager/internal/version.AppVersion=${ARTIFACT_VERSION:-0.0.0} \
    	-X github.com/ptrvsrg/crack-hash/manager/internal/version.GoVersion=$(go version | cut -d " " -f 3) \
    	-X github.com/ptrvsrg/crack-hash/manager/internal/version.Platform=$(go env GOOS)/$(go env GOARCH)" \
    ./cmd/cli

FROM alpine:3.22 AS runtime

WORKDIR /app

COPY --from=build /app/manager/bin /app
COPY --from=build /app/manager/config/config.default.yaml config/config.yaml

RUN apk update \
    && apk add --no-cache ca-certificates tzdata \
    && echo 'Etc/UTC' > /etc/timezone \
    && adduser --disabled-password --home /app --gecos '' gouser \
    && chown -R gouser /app

ENV TZ=Etc/UTC
ENV LANG=en_US.utf8
ENV LC_ALL=en_US.UTF-8

USER gouser

ENTRYPOINT [ "/app/manager", "server" ]