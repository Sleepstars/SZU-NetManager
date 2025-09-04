# syntax=docker/dockerfile:1

ARG SZU_LOGIN_URL=https://github.com/Sleepstars/SZU-login/releases/latest/download/srun-login-linux-amd64

FROM node:20-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json* web/pnpm-lock.yaml* ./
RUN npm ci || npm i
COPY web/ .
RUN npm run build

FROM golang:1.21-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates curl
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /out/netmanager ./cmd/netmanager
RUN curl -fL "$SZU_LOGIN_URL" -o /out/srun-login && chmod +x /out/srun-login || (echo "Failed to download SZU-login" && exit 1)

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

