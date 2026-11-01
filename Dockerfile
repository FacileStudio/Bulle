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

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=api-build /repo/apps/api/bin/api /api
COPY --from=client-build /client/build /client

# The distroless base can carry its own WorkingDir (/home/nonroot on the
# :nonroot variant), which would make a relative ./client resolve there and
# the SPA silently not be served at all. Be explicit.
ENV CLIENT_DIR=/client

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
