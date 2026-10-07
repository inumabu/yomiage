# Bot VPS

512MiB を意識して Docker を使わず、Go の単体バイナリ + ffmpeg + systemd で動かします。

TTS WAV は Go heap に保持せず、ローカルキャッシュへ直接ストリームします。VOICEVOX 合成は全 guild 共通で 1 本に直列化します。
