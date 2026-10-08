# VOICEVOX（VPS）

低メモリVPS（512MiB〜1GB）を想定したVOICEVOX Engineのsystemd・swap・nftables設定です。

`VOICEVOX_BIN`は実際に配置したEngineのパスへ変更してください。

起動オプションはEngineの版によって異なるため、CPUスレッド数などは`VOICEVOX_EXTRA_ARGS`で、使用中のEngineが対応する引数だけを指定してください。
