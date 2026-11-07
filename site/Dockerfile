# syntax=docker/dockerfile:1
# Documentation site (trading.vukem.dev): Astro static build served by nginx. Build context is the
# REPOSITORY ROOT: the site renders ../CHANGELOG.md and its content is generated from ../docs.
#
#   docker build -f site/Dockerfile -t vukem-trading-site .
#
# On Railway set RAILWAY_DOCKERFILE_PATH=site/Dockerfile for the site service.

FROM node:24.22.0-alpine3.24@sha256:44fce696f38ba8c987c4db89954051000bbd05ffd8d8659ba15ed4a8602e937f AS build
# Bun installs the dependencies and runs sync:docs; Astro runs on Node.
COPY --from=oven/bun:1.4.4-alpine@sha256:4cf82f5686b196504fe119db3fcd9185b191f82ece82c40680e597228ec6070a /usr/local/bin/bun /usr/local/bin/bun
WORKDIR /src/site
COPY site/package.json site/bun.lock ./
RUN bun install --frozen-lockfile
COPY site/ ./
COPY docs /src/docs
COPY CHANGELOG.md CONTRIBUTING.md SECURITY.md CODE_OF_CONDUCT.md /src/
RUN bun run sync:docs && bun run build

FROM nginx:1.31.0-alpine@sha256:2f07d83bf561b506400dc183b1b2003803e39efbd22451f848adaba14d28c7c7
# The official image renders /etc/nginx/templates/*.template with the container's environment on start.
ENV PORT=3000
COPY site/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /src/site/out /usr/share/nginx/html
EXPOSE 3000
