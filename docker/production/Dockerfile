FROM node:24-alpine AS assets
WORKDIR /build

RUN apk add --no-cache git \
    && git config --global --add safe.directory /build \
    && corepack enable
COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

COPY vite.config.ts tsconfig.json ./
COPY assets ./assets
COPY views ./views

ARG SENTRY_ORG
ARG SENTRY_PROJECT
ARG SENTRY_REPOSITORY
RUN --mount=type=bind,source=.git,target=/build/.git \
    --mount=type=secret,id=SENTRY_AUTH_TOKEN \
    SENTRY_UPLOAD_TOKEN="$(cat /run/secrets/SENTRY_AUTH_TOKEN 2>/dev/null)" \
    SENTRY_RELEASE="$(git rev-parse HEAD)" \
    yarn build \
    && find public/build -name '*.map' -delete

FROM golang:1.26-alpine AS build
WORKDIR /build

RUN apk add --no-cache git \
    && git config --global --add safe.directory /build

COPY go.mod go.sum ./
RUN go mod download

COPY app ./app
COPY locale ./locale
COPY migrations ./migrations
COPY views ./views
COPY public ./public
COPY --from=assets /build/public/build ./public/build

ENV CGO_ENABLED=0 GOOS=linux
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.git,target=/build/.git \
    go build -trimpath -buildvcs=true -ldflags="-s -w" -o /out/gokick ./app

FROM alpine:3.24

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 1000 gokick \
    && adduser -S -D -H -u 1000 -G gokick gokick

WORKDIR /app
COPY --from=build /out/gokick /app/gokick

USER gokick:gokick

EXPOSE 8020

HEALTHCHECK --start-period=30s --start-interval=1s --interval=30s --timeout=3s --retries=3 \
    CMD wget -q --spider "http://127.0.0.1:${PORT:-8020}/readyz" || exit 1

ENTRYPOINT ["/app/gokick"]
