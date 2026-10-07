# 🐳 Linux container for Docker Desktop on Windows (WSL2 backend)
FROM golang:1.22-bookworm AS build

ARG UPSTREAM_COMMIT=3b810b45153f200d8a6ec8e190dee7a52d090d6e
WORKDIR /work

RUN apt-get update \
    && apt-get install -y --no-install-recommends git patch ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --filter=blob:none --no-checkout https://github.com/inumabu/yomiage.git /work/yomiage \
    && cd /work/yomiage \
    && git fetch --depth 1 origin "$UPSTREAM_COMMIT" \
    && git checkout --detach "$UPSTREAM_COMMIT" \
    && test "$(git rev-parse HEAD)" = "$UPSTREAM_COMMIT"

COPY patches/0001-keiryou-cache.patch /work/yomiage-keiryou.patch
RUN cd /work/yomiage \
    && git apply --check /work/yomiage-keiryou.patch \
    && git apply /work/yomiage-keiryou.patch \
    && go test ./... \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/yomiage-keiryou .

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --create-home --home-dir /var/lib/yomiage-keiryou --shell /usr/sbin/nologin yomiage-keiryou \
    && install -d -o yomiage-keiryou -g yomiage-keiryou -m 0750 /var/lib/yomiage-keiryou/cache/tts /var/lib/yomiage-keiryou/tmp

COPY --from=build /out/yomiage-keiryou /usr/local/bin/yomiage-keiryou
RUN chmod 0755 /usr/local/bin/yomiage-keiryou

USER yomiage-keiryou
ENV YOMIAGE_SETTINGS_FILE=/var/lib/yomiage-keiryou/settings.json \
    YOMIAGE_CACHE_DIR=/var/lib/yomiage-keiryou/cache/tts \
    YOMIAGE_CACHE_TTL=168h \
    YOMIAGE_CACHE_MAX_BYTES=268435456 \
    YOMIAGE_MAX_AUDIO_BYTES=33554432 \
    GOMEMLIMIT=128MiB \
    GOGC=50 \
    TMPDIR=/var/lib/yomiage-keiryou/tmp

ENTRYPOINT ["/usr/local/bin/yomiage-keiryou"]
