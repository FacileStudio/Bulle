FROM oven/bun:1 AS client-build
WORKDIR /client
COPY apps/client/package.json apps/client/bun.lock* ./
RUN bun install --frozen-lockfile
COPY apps/client/ .
RUN bun run build

FROM golang:1.26-alpine AS api-build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo

COPY go.work go.work.sum ./
COPY core/go.mod core/go.sum ./core/
COPY apps/cli/go.mod apps/cli/go.sum ./apps/cli/
COPY apps/api/go.mod apps/api/go.sum ./apps/api/

RUN cd core && go mod download && cd ../apps/api && go mod download

COPY core ./core
COPY apps/api ./apps/api

WORKDIR /repo/apps/api

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o bin/api .

FROM ubuntu:24.04

ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
    openssh-client \
    ca-certificates \
    curl \
    tar \
    gzip \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd -g 65532 nonroot \
    && useradd -u 65532 -g nonroot -m -d /home/nonroot -s /bin/sh nonroot

RUN ARCH="${TARGETARCH:-amd64}" \
    && curl -fsSL "https://github.com/FacileStudio/courrier-cli/releases/download/v0.6.0/courrier-cli_0.6.0_linux_${ARCH}.tar.gz" | tar -xz -C /usr/local/bin courrier \
    && curl -fsSL "https://github.com/FacileStudio/agenda-cli/releases/download/v0.5.0/agenda_0.5.0_linux_${ARCH}.tar.gz" | tar -xz -C /usr/local/bin agenda \
    && curl -fsSL "https://github.com/FacileStudio/opus-cli/releases/download/v0.7.1/opus_0.7.1_linux_${ARCH}.tar.gz" | tar -xz -C /usr/local/bin opus \
    && curl -fsSL "https://github.com/FacileStudio/sablier-cli/releases/download/v0.1.0/sablier_0.1.0_linux_${ARCH}.tar.gz" | tar -xz -C /usr/local/bin sablier \
    && chmod +x /usr/local/bin/courrier /usr/local/bin/agenda /usr/local/bin/opus /usr/local/bin/sablier

COPY --from=api-build /repo/apps/api/bin/api /api
COPY --from=client-build /client/build /client

ENV CLIENT_DIR=/client

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
