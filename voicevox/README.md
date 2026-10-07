# VOICEVOX VPS

低メモリVPS（512MiB〜1GB）を想定した VOICEVOX Engine の systemd / swap / nftables 設定です。

`VOICEVOX_BIN` は実際に配置した Engine のパスへ変更してください。

起動オプションは Engine の版によって差があるため、CPU スレッド数等は `VOICEVOX_EXTRA_ARGS` で、使用中の Engine がサポートする引数だけを指定してください。
