ARG NODE_VERSION=23.6.0-alpine3.21
ARG NGINX_VERSION=1.28.0-alpine3.21

FROM node:${NODE_VERSION} AS deps

WORKDIR /app

COPY package.json yarn.lock ./

RUN yarn install --frozen-lockfile


FROM deps AS builder

WORKDIR /app

COPY . .

RUN NODE_ENV=production yarn build


FROM nginx:${NGINX_VERSION} AS runtime

COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html

RUN apk update \
    && apk add --no-cache --upgrade \
        bash \
        ca-certificates \
        curl \
        tzdata \
    && update-ca-certificates \
    && echo 'Etc/UTC' > /etc/timezone

ENV TZ     :/etc/localtime
ENV LANG   en_US.utf8
ENV LC_ALL en_US.UTF-8

ENTRYPOINT [ "nginx", "-g", "daemon off;" ]