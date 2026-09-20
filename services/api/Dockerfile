# syntax=docker/dockerfile:1
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY services/api/go.mod services/api/go.sum ./services/api/
COPY internal ./internal
COPY db ./db
COPY services/api ./services/api
WORKDIR /src/services/api
RUN go build -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/api /usr/local/bin/api
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/api"]
