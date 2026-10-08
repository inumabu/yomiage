# Bot（VPS）

512MiB を意識して Docker を使わず、Go の単体バイナリ + ffmpeg + systemd で動かします。

TTS WAV はGoのヒープに保持せず、ローカルキャッシュへ直接ストリームします。VOICEVOX合成は全Guild共通で1件ずつ直列化します。
