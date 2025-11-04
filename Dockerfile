FROM golang:1.22 AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=""

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-X 'github.com/religiosa1/tgnotifier/internal/cmd.version=${VERSION}'" \
    -o /go/bin/app ./cmd/tgnotifier
RUN chmod +x /go/bin/app

FROM gcr.io/distroless/static-debian12

ARG VERSION=""
ARG BUILD_DATE=""
ARG VCS_REF=""

LABEL org.opencontainers.image.title="tgnotifier"
LABEL org.opencontainers.image.description="Telegram notification service with HTTP API"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.source="https://github.com/religiosa1/tgnotifier"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="religiosa1"

WORKDIR /app

COPY --from=build /go/bin/app ./tgnotifier

ENV BOT_ADDR=0.0.0.0:6000
ENV BOT_CONFIG_PATH=/config.yml

EXPOSE 6000

# Run as non-root user for security (distroless nonroot user)
USER 65532:65532

CMD ["./tgnotifier"]