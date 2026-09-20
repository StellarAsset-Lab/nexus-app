# syntax=docker/dockerfile:1
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY services/worker/go.mod services/worker/go.sum ./services/worker/
COPY internal ./internal
COPY db ./db
COPY services/worker ./services/worker
WORKDIR /src/services/worker
RUN go build -o /out/worker ./cmd/worker

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/worker /usr/local/bin/worker
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/worker"]
