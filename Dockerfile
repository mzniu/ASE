# Multi-stage build for ASE API server (go.mod toolchain may auto-download via GOTOOLCHAIN).
FROM golang:1.24-bookworm AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# base-debian12 includes CA roots for HTTPS (e.g. OpenSearch / Tavily).
FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /server /server
USER nonroot:nonroot
EXPOSE 18080
ENTRYPOINT ["/server"]
