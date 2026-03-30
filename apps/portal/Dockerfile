FROM golang:1.24 AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/portal ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/portal /app/portal
COPY --from=build /src/web /app/web

ENV PORT=8080
ENV EXECUTOR_MODE=kubernetes

EXPOSE 8080

ENTRYPOINT ["/app/portal"]
