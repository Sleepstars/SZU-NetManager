# syntax=docker/dockerfile:1

# Build args for multi-arch download
ARG TARGETARCH
ARG SZU_LOGIN_VERSION=v0.1.1-alpha

FROM node:23-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json* web/pnpm-lock.yaml* ./
RUN npm ci || npm i
COPY web/ .
RUN npm run build

FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates curl
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /out/netmanager ./cmd/netmanager
ARG TARGETARCH
ARG SZU_LOGIN_VERSION
RUN set -eux; \
    mkdir -p /out; \
    case "$TARGETARCH" in \
      amd64) BIN="szu-login-linux-amd64" ;; \
      arm64) BIN="szu-login-linux-arm64" ;; \
      *) echo "Unsupported arch: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    URL="https://github.com/Sleepstars/SZU-login/releases/download/${SZU_LOGIN_VERSION}/${BIN}"; \
    curl -fL "$URL" -o /out/srun-login; \
    chmod +x /out/srun-login

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/netmanager /usr/local/bin/netmanager
COPY --from=build /out/srun-login /usr/local/bin/srun-login
COPY --from=web /web/dist /app/web
ENV NM_WEB_DIR=/app/web
ENV NM_SZU_LOGIN=/usr/local/bin/srun-login
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/netmanager"]
