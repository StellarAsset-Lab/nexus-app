# syntax=docker/dockerfile:1
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY services/indexer/go.mod ./services/indexer/
COPY internal ./internal
COPY services/indexer ./services/indexer
WORKDIR /src/services/indexer
RUN go build -o /out/indexer ./cmd/indexer

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/indexer /usr/local/bin/indexer
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/indexer"]
