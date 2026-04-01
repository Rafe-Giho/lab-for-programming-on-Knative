# syntax=docker/dockerfile:1.7

FROM golang:1.24 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/portal ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/portal /app/portal
COPY web /app/web

ENV PORT=8080
ENV EXECUTOR_MODE=kubernetes    

EXPOSE 8080

ENTRYPOINT ["/app/portal"]
