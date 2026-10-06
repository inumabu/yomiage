FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/yomiage .

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/yomiage /app/yomiage
ENV VOICEVOX_URL=http://voicevox:50021 \
    YOMIAGE_SETTINGS_FILE=/data/settings.json
VOLUME ["/data"]
ENTRYPOINT ["/app/yomiage"]
