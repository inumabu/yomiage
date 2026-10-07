# 🐳 DAVE対応Linux container for Docker Desktop on Windows and amd64 VPS
FROM golang:1.24-bookworm AS build

ARG UPSTREAM_COMMIT=3b810b45153f200d8a6ec8e190dee7a52d090d6e
ARG DAVE_REF=v0.29.1-dave.26
WORKDIR /work

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       git patch ca-certificates cmake ninja-build g++ make pkg-config \
       autoconf automake libtool zip unzip tar libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# The DAVE fork contains a libdave git submodule, so this must be a recursive clone.
RUN git clone --depth 1 --branch "$DAVE_REF" --recurse-submodules \
      https://github.com/aleph-garden/discordgo.git /work/discordgo \
    && cd /work/discordgo/dave/libdave/cpp \
    && ./vcpkg/bootstrap-vcpkg.sh -disableMetrics \
    && make BUILD_TYPE=Release

RUN git clone --filter=blob:none --no-checkout https://github.com/inumabu/yomiage.git /work/yomiage \
    && cd /work/yomiage \
    && git fetch --depth 1 origin "$UPSTREAM_COMMIT" \
    && git checkout --detach "$UPSTREAM_COMMIT" \
    && test "$(git rev-parse HEAD)" = "$UPSTREAM_COMMIT"

COPY patches/0001-keiryou-cache.patch /work/yomiage-keiryou.patch
COPY patches/0002-dave.patch /work/yomiage-dave.patch
RUN cd /work/yomiage \
    && git apply --check /work/yomiage-keiryou.patch \
    && git apply --check /work/yomiage-dave.patch \
    && git apply /work/yomiage-keiryou.patch \
    && git apply /work/yomiage-dave.patch

RUN cd /work/yomiage \
    && set -eux; \
    LIBDAVE_BUILD=/work/discordgo/dave/libdave/cpp/build; \
    VCPKG_LIBDIR="$(find "$LIBDAVE_BUILD/vcpkg_installed" -maxdepth 2 -type d -name lib | head -1)"; \
    VCPKG_LIBS="$(find "$VCPKG_LIBDIR" -maxdepth 1 -name '*.a' | sort | tr '\n' ' ')"; \
    export CGO_ENABLED=1; \
    export CGO_CFLAGS="-I/work/discordgo/dave/libdave/cpp/includes"; \
    export CGO_LDFLAGS="-L$VCPKG_LIBDIR $LIBDAVE_BUILD/libdave.a -Wl,--start-group $VCPKG_LIBS -Wl,--end-group -lstdc++ -lm -ldl -lpthread"; \
    go test -vet=off ./...; \
    go build -trimpath -ldflags='-s -w' -o /out/yomiage-keiryou .

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg ca-certificates libstdc++6 \
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
