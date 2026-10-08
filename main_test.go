package main

import (
	"context"
	"testing"
)

func newTestBot(t *testing.T) *bot {
	t.Helper()
	return &bot{
		speaker:      3,
		players:      make(map[string]*player),
		settings:     make(map[string]guildSettings),
		settingsPath: t.TempDir() + "/settings.json",
	}
}

func TestQueueMessageAndClearQueue(t *testing.T) {
	b := newTestBot(t)
	guildID := "guild-1"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.players[guildID] = &player{
		guildID: guildID,
		queue:   make(chan string, 3),
		ctx:     ctx,
		cancel:  cancel,
	}
	b.players[guildID].enqueue("一つ目")
	b.players[guildID].enqueue("二つ目")

	if got := b.queueMessage(guildID); got != "📚 読み上げキュー（2件）:\n1. 一つ目\n2. 二つ目" {
		t.Fatalf("queueMessage() = %q", got)
	}
	if got := b.clearQueue(guildID); got != "待機中の読み上げを 2 件取り消しました。" {
		t.Fatalf("clearQueue() = %q", got)
	}
	if got := len(b.players[guildID].queue); got != 0 {
		t.Fatalf("queue length after clear = %d, want 0", got)
	}
}

func TestQueueOperationsWithoutPlayer(t *testing.T) {
	b := newTestBot(t)
	if got := b.queueMessage("missing"); got != "現在、このサーバーでは読み上げをしていません。" {
		t.Fatalf("queueMessage() = %q", got)
	}
	if got := b.clearQueue("missing"); got != "現在、このサーバーでは読み上げをしていません。" {
		t.Fatalf("clearQueue() = %q", got)
	}
}

func TestSetSpeakerAndStatus(t *testing.T) {
	b := newTestBot(t)
	guildID := "guild-1"
	if got := b.setSpeaker(guildID, 0); got != "話者IDは1以上で指定してください。" {
			t.Fatalf("setSpeaker()の結果が不正です = %q", got)
	}
	b.setSpeaker(guildID, 8)
	if got := b.guildSpeaker(guildID); got != 8 {
		t.Fatalf("guildSpeaker() = %d, want 8", got)
	}
	if got := b.statusMessage(guildID); got != "🎧 現在、このサーバーでは読み上げを停止中です。話者ID: 8 / 音量: 1.00 / 速度: 1.00" {
		t.Fatalf("statusMessage() = %q", got)
	}
}

func TestSettingsPersistAndExclusions(t *testing.T) {
	b := newTestBot(t)
	if got := b.setVolume(guildIDForTest, 1.5); got != "音量を 1.50 に設定しました。" {
		t.Fatalf("setVolume() = %q", got)
	}
	if got := b.setSpeed(guildIDForTest, 1.25); got != "読み上げ速度を 1.25 に設定しました。" {
		t.Fatalf("setSpeed() = %q", got)
	}
	if got := b.configureChannel(guildIDForTest, []string{"allow", "channel-1"}); got == "" || b.isExcluded(guildIDForTest, "channel-2", "user-1") == false {
		t.Fatal("allow-list channel should exclude other channels")
	}
	b.configureUser(guildIDForTest, []string{"block", "user-1"})
	if !b.isExcluded(guildIDForTest, "channel-1", "user-1") {
		t.Fatal("blocked user should be excluded")
	}
	if err := b.saveSettings(); err != nil {
		t.Fatalf("saveSettings()でエラー = %v", err)
	}

	loaded := newTestBot(t)
	loaded.settingsPath = b.settingsPath
	if err := loaded.loadSettings(); err != nil {
		t.Fatalf("loadSettings()でエラー = %v", err)
	}
	setting := loaded.guildSettings(guildIDForTest)
	if setting.Volume != 1.5 || setting.Speed != 1.25 || !loaded.isExcluded(guildIDForTest, "channel-1", "user-1") {
			t.Fatalf("保存した設定を復元できませんでした: %+v", setting)
	}
}

const guildIDForTest = "guild-settings"

func TestVolumeAndSpeedValidation(t *testing.T) {
	b := newTestBot(t)
	if got := b.setVolume("guild", 2.1); got != "音量は 0.0〜2.0 の範囲で指定してください。" {
			t.Fatalf("音量が不正です = %q", got)
	}
	if got := b.setSpeed("guild", 0.4); got != "速度は 0.5〜2.0 の範囲で指定してください。" {
			t.Fatalf("速度が不正です = %q", got)
	}
}

func TestPauseAndResume(t *testing.T) {
	b := newTestBot(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.players["guild"] = &player{queue: make(chan string, 1), ctx: ctx, cancel: cancel}
	if got := b.pause("guild"); got == "" {
		t.Fatal("pause returned an empty message")
	}
	b.players["guild"].mu.Lock()
	paused := b.players["guild"].paused
	b.players["guild"].mu.Unlock()
	if !paused {
		t.Fatal("player was not paused")
	}
	b.resume("guild")
	b.players["guild"].mu.Lock()
	paused = b.players["guild"].paused
	b.players["guild"].mu.Unlock()
	if paused {
		t.Fatal("player was not resumed")
	}
}
