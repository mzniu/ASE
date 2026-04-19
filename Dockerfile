# ASE API + Chromium（默认镜像：baidu/bing/google 等 chromedp Provider）。
# 更小、无浏览器的镜像见 Dockerfile.distroless。
FROM golang:1.24-bookworm AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	chromium \
	dumb-init \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=build /server /server
RUN useradd -r -u 65532 -s /bin/false -m -d /var/lib/ase ase \
	&& chown ase:ase /server
USER ase
WORKDIR /var/lib/ase
EXPOSE 18080
ENTRYPOINT ["dumb-init", "--", "/server"]
