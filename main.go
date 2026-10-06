package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca/v2"
)

const (
	commandPrefix       = "!"
	joinCommandName     = "join"
	leaveCommandName    = "leave"
	sayCommandName      = "say"
	speakerCommandName  = "speaker"
	statusCommandName   = "status"
	queueCommandName    = "queue"
	clearCommandName    = "clear"
	skipCommandName     = "skip"
	volumeCommandName   = "volume"
	speedCommandName    = "speed"
	channelCommandName  = "channel"
	userCommandName     = "user"
	helpCommandName     = "help"
	speakersCommandName = "speakers"
	pauseCommandName    = "pause"
	resumeCommandName   = "resume"
	removeCommandName   = "remove"
	maxMessageRunes     = 120
	queueCapacity       = 30
	defaultVolume       = 1.0
	defaultSpeed        = 1.0
)

var applicationCommands = []*discordgo.ApplicationCommand{
	{
		Name:        joinCommandName,
		Description: "ボイスチャンネルに接続して読み上げを開始します。",
	},
	{
		Name:        leaveCommandName,
		Description: "ボイスチャンネルから切断して読み上げを終了します。",
	},
	{
		Name:        sayCommandName,
		Description: "現在のボイスチャンネルで指定したテキストを読み上げます。",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "読み上げる文言",
				Required:    true,
			},
		},
	},
	{
		Name:        speakerCommandName,
		Description: "このサーバーで使うVOICEVOXの話者IDを変更します。",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "VOICEVOXの話者ID",
				Required:    true,
			},
		},
	},
	{
		Name:        statusCommandName,
		Description: "読み上げBOTの現在状態を表示します。",
	},
	{
		Name:        queueCommandName,
		Description: "現在の読み上げキューの件数を表示します。",
	},
	{
		Name:        clearCommandName,
		Description: "待機中の読み上げをすべて取り消します。",
	},
	{
		Name:        skipCommandName,
		Description: "現在の読み上げをスキップします。",
	},
	{
		Name:        volumeCommandName,
		Description: "このサーバーの音量を変更します（0.0〜2.0）。",
		Options:     []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionNumber, Name: "value", Description: "音量", Required: true}},
	},
	{
		Name:        speedCommandName,
		Description: "このサーバーの読み上げ速度を変更します（0.5〜2.0）。",
		Options:     []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionNumber, Name: "value", Description: "速度", Required: true}},
	},
	{
		Name:        channelCommandName,
		Description: "読み上げ対象チャンネルを設定します（管理者専用）。",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "action", Description: "allow / block / remove", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "チャンネルID", Required: true},
		},
	},
	{
		Name:        userCommandName,
		Description: "読み上げ対象ユーザーを設定します（管理者専用）。",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "action", Description: "allow / block / remove", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "ユーザーID", Required: true},
		},
	},
	{Name: helpCommandName, Description: "利用できるコマンドを表示します。"},
	{Name: speakersCommandName, Description: "VOICEVOXの話者一覧を表示します。"},
	{Name: pauseCommandName, Description: "次の読み上げを一時停止します。"},
	{Name: resumeCommandName, Description: "読み上げを再開します。"},
	{
		Name:        removeCommandName,
		Description: "待機キューから指定番号の文を削除します（管理者専用）。",
		Options:     []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionInteger, Name: "index", Description: "キュー番号", Required: true}},
	},
}

type guildSettings struct {
	Speaker         int             `json:"speaker"`
	Volume          float64         `json:"volume"`
	VolumeSet       bool            `json:"volume_set,omitempty"`
	Speed           float64         `json:"speed"`
	SpeedSet        bool            `json:"speed_set,omitempty"`
	AllowedChannels map[string]bool `json:"allowed_channels,omitempty"`
	BlockedChannels map[string]bool `json:"blocked_channels,omitempty"`
	BlockedUsers    map[string]bool `json:"blocked_users,omitempty"`
}

type bot struct {
	session *discordgo.Session
	ttsURL  string
	speaker int
	client  *http.Client

	mu           sync.Mutex
	players      map[string]*player
	settings     map[string]guildSettings
	settingsPath string
}

type player struct {
	guildID         string
	voice           *discordgo.VoiceConnection
	notifyChannelID string
	queue           chan string
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.Mutex
	speechCancel    context.CancelFunc
	paused          bool
	queueMu         sync.Mutex
	queued          []string
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("🔐 DISCORD_TOKEN が必要です")
	}

	speaker := 3
	if value := os.Getenv("VOICEVOX_SPEAKER"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			log.Fatalf("⚠️ VOICEVOX_SPEAKER が不正です: %v", err)
		}
		speaker = parsed
	}

	ttsURL := strings.TrimRight(os.Getenv("VOICEVOX_URL"), "/")
	if ttsURL == "" {
		ttsURL = "http://localhost:50021"
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("❌ Discord セッションの作成に失敗しました: %v", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsMessageContent

	app := &bot{
		session:      session,
		ttsURL:       ttsURL,
		speaker:      speaker,
		client:       &http.Client{Timeout: 30 * time.Second},
		players:      make(map[string]*player),
		settings:     make(map[string]guildSettings),
		settingsPath: settingsFilePath(),
	}
	if err := app.loadSettings(); err != nil {
		log.Printf("⚠️ 設定の読み込みに失敗しました: %v", err)
	}
	session.AddHandler(app.onMessage)
	session.AddHandler(app.onInteraction)

	if err := session.Open(); err != nil {
		log.Fatalf("❌ Discord セッションの開始に失敗しました: %v", err)
	}
	defer session.Close()
	if err := app.registerCommands(); err != nil {
		log.Printf("⚠️ スラッシュコマンドの登録に失敗しました: %v", err)
	}
	log.Println("✅ 読み上げ BOT を起動しました")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	app.stopAll()
}

func (b *bot) onMessage(s *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author == nil || message.Author.Bot || message.GuildID == "" {
		return
	}

	content := strings.TrimSpace(message.Content)
	switch {
	case content == commandPrefix+joinCommandName:
		b.joinMessage(message)
		return
	case content == commandPrefix+leaveCommandName:
		b.reply(message.ChannelID, message.GuildID, b.leaveForCommand(message.GuildID), true)
		return
	case content == commandPrefix+sayCommandName:
		b.reply(message.ChannelID, message.GuildID, "読み上げる文章を指定してください。", false)
		return
	case strings.HasPrefix(content, commandPrefix+sayCommandName+" "):
		text := strings.TrimSpace(strings.TrimPrefix(content, commandPrefix+sayCommandName))
		b.reply(message.ChannelID, message.GuildID, b.enqueueText(message.GuildID, text, "メッセージコマンド"), true)
		return
	case strings.HasPrefix(content, commandPrefix+speakerCommandName+" "):
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		value := strings.TrimSpace(strings.TrimPrefix(content, commandPrefix+speakerCommandName))
		id, err := strconv.Atoi(value)
		if err != nil {
			b.reply(message.ChannelID, message.GuildID, "話者IDは数値で指定してください。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.setSpeaker(message.GuildID, id), true)
		return
	case content == commandPrefix+statusCommandName:
		b.reply(message.ChannelID, message.GuildID, b.statusMessage(message.GuildID), true)
		return
	case content == commandPrefix+speakerCommandName:
		b.reply(message.ChannelID, message.GuildID, "話者IDを指定してください。", false)
		return
	case content == commandPrefix+queueCommandName:
		b.reply(message.ChannelID, message.GuildID, b.queueMessage(message.GuildID), true)
		return
	case strings.HasPrefix(content, commandPrefix+removeCommandName+" "):
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		index, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(content, commandPrefix+removeCommandName)))
		if err != nil {
			b.reply(message.ChannelID, message.GuildID, "削除するキュー番号は数値で指定してください。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.removeQueue(message.GuildID, index), true)
		return
	case content == commandPrefix+clearCommandName:
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.clearQueue(message.GuildID), true)
		return
	case content == commandPrefix+skipCommandName:
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.skip(message.GuildID), true)
		return
	case content == commandPrefix+pauseCommandName:
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.pause(message.GuildID), true)
		return
	case content == commandPrefix+resumeCommandName:
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.resume(message.GuildID), true)
		return
	case content == commandPrefix+helpCommandName:
		b.reply(message.ChannelID, message.GuildID, helpMessage(), true)
		return
	case content == commandPrefix+speakersCommandName:
		b.reply(message.ChannelID, message.GuildID, b.speakerListMessage(), true)
		return
	case strings.HasPrefix(content, commandPrefix+volumeCommandName+" "):
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(content, commandPrefix+volumeCommandName)), 64)
		if err != nil {
			b.reply(message.ChannelID, message.GuildID, "音量は数値で指定してください。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.setVolume(message.GuildID, value), true)
		return
	case strings.HasPrefix(content, commandPrefix+speedCommandName+" "):
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(content, commandPrefix+speedCommandName)), 64)
		if err != nil {
			b.reply(message.ChannelID, message.GuildID, "速度は数値で指定してください。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.setSpeed(message.GuildID, value), true)
		return
	case strings.HasPrefix(content, commandPrefix+channelCommandName+" "):
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.configureChannel(message.GuildID, strings.Fields(strings.TrimPrefix(content, commandPrefix+channelCommandName))), true)
		return
	case strings.HasPrefix(content, commandPrefix+userCommandName+" "):
		if !b.isAdminMessage(message) {
			b.reply(message.ChannelID, message.GuildID, "この操作はサーバー管理者のみ実行できます。", false)
			return
		}
		b.reply(message.ChannelID, message.GuildID, b.configureUser(message.GuildID, strings.Fields(strings.TrimPrefix(content, commandPrefix+userCommandName))), true)
		return
	}

	if content == "" || strings.HasPrefix(content, commandPrefix) || utf8.RuneCountInString(content) > maxMessageRunes {
		return
	}

	b.mu.Lock()
	current := b.players[message.GuildID]
	b.mu.Unlock()
	if current == nil {
		return
	}
	if b.isExcluded(message.GuildID, message.ChannelID, message.Author.ID) {
		return
	}

	if !current.enqueue(content) {
		log.Printf("読み上げキューが満杯です guild=%s", message.GuildID)
	}
}

func (b *bot) onInteraction(s *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		if interaction.GuildID == "" {
			return
		}

		user := interaction.User
		if interaction.Member != nil {
			user = interaction.Member.User
		}
		if user == nil {
			return
		}

		data := interaction.ApplicationCommandData()
		switch data.Name {
		case joinCommandName:
			b.replyInteraction(interaction, b.joinForUser(interaction.GuildID, user.ID, interaction.ChannelID), true)
		case leaveCommandName:
			b.replyInteraction(interaction, b.leaveForCommand(interaction.GuildID), true)
		case sayCommandName:
			if len(data.Options) == 0 {
				b.replyInteraction(interaction, "読み上げる文章を指定してください。", false)
				return
			}
			text := data.Options[0].StringValue()
			b.replyInteraction(interaction, b.enqueueText(interaction.GuildID, text, "スラッシュコマンド"), true)
		case speakerCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			if len(data.Options) == 0 {
				b.replyInteraction(interaction, "話者IDを指定してください。", false)
				return
			}
			id := int(data.Options[0].IntValue())
			b.replyInteraction(interaction, b.setSpeaker(interaction.GuildID, id), true)
		case statusCommandName:
			b.replyInteraction(interaction, b.statusMessage(interaction.GuildID), true)
		case queueCommandName:
			b.replyInteraction(interaction, b.queueMessage(interaction.GuildID), true)
		case clearCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			b.replyInteraction(interaction, b.clearQueue(interaction.GuildID), true)
		case removeCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			if option := commandOption(data.Options, "index"); option != nil {
				b.replyInteraction(interaction, b.removeQueue(interaction.GuildID, int(option.IntValue())), true)
			} else {
				b.replyInteraction(interaction, "削除するキュー番号を指定してください。", false)
			}
		case skipCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			b.replyInteraction(interaction, b.skip(interaction.GuildID), true)
		case pauseCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			b.replyInteraction(interaction, b.pause(interaction.GuildID), true)
		case resumeCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			b.replyInteraction(interaction, b.resume(interaction.GuildID), true)
		case helpCommandName:
			b.replyInteraction(interaction, helpMessage(), true)
		case speakersCommandName:
			b.replyInteraction(interaction, b.speakerListMessage(), true)
		case volumeCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			if option := commandOption(data.Options, "value"); option != nil {
				b.replyInteraction(interaction, b.setVolume(interaction.GuildID, option.FloatValue()), true)
			} else {
				b.replyInteraction(interaction, "音量を指定してください。", false)
			}
		case speedCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			if option := commandOption(data.Options, "value"); option != nil {
				b.replyInteraction(interaction, b.setSpeed(interaction.GuildID, option.FloatValue()), true)
			} else {
				b.replyInteraction(interaction, "速度を指定してください。", false)
			}
		case channelCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			b.replyInteraction(interaction, b.configureChannel(interaction.GuildID, []string{commandOptionString(data.Options, "action"), commandOptionString(data.Options, "id")}), true)
		case userCommandName:
			if !b.isAdminInteraction(interaction) {
				b.replyInteraction(interaction, "この操作はサーバー管理者のみ実行できます。", false)
				return
			}
			b.replyInteraction(interaction, b.configureUser(interaction.GuildID, []string{commandOptionString(data.Options, "action"), commandOptionString(data.Options, "id")}), true)
		}
	case discordgo.InteractionMessageComponent:
		customID := interaction.MessageComponentData().CustomID
		user := interaction.User
		if interaction.Member != nil {
			user = interaction.Member.User
		}
		if user == nil {
			return
		}
		switch customID {
		case "join_button":
			b.replyInteraction(interaction, b.joinForUser(interaction.GuildID, user.ID, interaction.ChannelID), true)
		case "leave_button":
			b.replyInteraction(interaction, b.leaveForCommand(interaction.GuildID), true)
		case "status_button":
			b.replyInteraction(interaction, b.statusMessage(interaction.GuildID), true)
		case "queue_button":
			b.replyInteraction(interaction, b.queueMessage(interaction.GuildID), true)
		case "clear_button":
			b.replyInteraction(interaction, b.clearQueue(interaction.GuildID), true)
		}
	}
}

func (b *bot) replyInteraction(interaction *discordgo.InteractionCreate, content string, ok bool) {
	if err := b.session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{newEmbed("📣 読み上げBOT", content, ok, b.embedFieldsForGuild(interaction.GuildID))},
			Components: []discordgo.MessageComponent{newActionRow()},
		},
	}); err != nil {
		log.Printf("⚠️ Interaction への応答に失敗しました: %v", err)
	}
}

func (b *bot) joinMessage(message *discordgo.MessageCreate) {
	b.reply(message.ChannelID, message.GuildID, b.joinForUser(message.GuildID, message.Author.ID, message.ChannelID), true)
}

func (b *bot) pause(guildID string) string {
	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	if current == nil {
		return "現在、このサーバーでは読み上げをしていません。"
	}
	current.mu.Lock()
	current.paused = true
	current.mu.Unlock()
	return "次の読み上げを一時停止しました。再開するには `/resume` を実行してください。"
}

func (b *bot) resume(guildID string) string {
	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	if current == nil {
		return "現在、このサーバーでは読み上げをしていません。"
	}
	current.mu.Lock()
	current.paused = false
	current.mu.Unlock()
	return "読み上げを再開しました。"
}

func helpMessage() string {
	return "基本: `!join` / `!leave` / `!say 文章` / `!status`\n" +
		"キュー: `!queue` / `!remove 番号` / `!clear` / `!skip` / `!pause` / `!resume`\n" +
		"設定（管理者）: `!speaker ID` / `!volume 0.0-2.0` / `!speed 0.5-2.0`\n" +
		"除外（管理者）: `!channel allow|block|remove チャンネルID` / `!user block|remove ユーザーID`\n" +
		"その他: `!speakers`"
}

func (b *bot) speakerListMessage() string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, b.ttsURL+"/speakers", nil)
	if err != nil {
		return "話者一覧のリクエストを作成できませんでした。"
	}
	response, err := b.client.Do(request)
	if err != nil {
		return "VOICEVOXに接続できないため、話者一覧を取得できません。"
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("話者一覧の取得に失敗しました（%s）。", response.Status)
	}
	var speakers []struct {
		Name   string `json:"name"`
		Styles []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"styles"`
	}
	if err := json.NewDecoder(response.Body).Decode(&speakers); err != nil {
		return "話者一覧の形式を読み取れませんでした。"
	}
	var lines []string
	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			lines = append(lines, fmt.Sprintf("%d: %s（%s）", style.ID, speaker.Name, style.Name))
		}
	}
	if len(lines) == 0 {
		return "利用可能な話者がありません。"
	}
	result := "利用可能な話者:\n" + strings.Join(lines, "\n")
	if utf8.RuneCountInString(result) > 1900 {
		result = string([]rune(result)[:1900]) + "\n…（続きはVOICEVOXの話者一覧を確認してください）"
	}
	return result
}

func isAdministrator(member *discordgo.Member) bool {
	return member != nil && member.Permissions&discordgo.PermissionAdministrator != 0
}

func (b *bot) isAdminMessage(message *discordgo.MessageCreate) bool {
	if isAdministrator(message.Member) {
		return true
	}
	if b.session == nil || message.Author == nil {
		return false
	}
	member, err := b.session.State.Member(message.GuildID, message.Author.ID)
	return err == nil && isAdministrator(member)
}

func (b *bot) isAdminInteraction(interaction *discordgo.InteractionCreate) bool {
	return interaction != nil && isAdministrator(interaction.Member)
}

func commandOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Name == name {
			return option
		}
	}
	return nil
}

func commandOptionString(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	if option := commandOption(options, name); option != nil {
		return option.StringValue()
	}
	return ""
}

func settingsFilePath() string {
	if path := strings.TrimSpace(os.Getenv("YOMIAGE_SETTINGS_FILE")); path != "" {
		return path
	}
	return "settings.json"
}

func (b *bot) loadSettings() error {
	data, err := os.ReadFile(b.settingsPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	loaded := make(map[string]guildSettings)
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("decode settings: %w", err)
	}
	b.mu.Lock()
	for guildID, setting := range loaded {
		b.settings[guildID] = normalizeSettings(setting, b.speaker)
	}
	b.mu.Unlock()
	return nil
}

func normalizeSettings(setting guildSettings, defaultSpeaker int) guildSettings {
	if setting.Speaker <= 0 {
		setting.Speaker = defaultSpeaker
	}
	if !setting.VolumeSet && setting.Volume == 0 {
		setting.Volume = defaultVolume
	}
	if !setting.SpeedSet && setting.Speed == 0 {
		setting.Speed = defaultSpeed
	}
	if setting.AllowedChannels == nil {
		setting.AllowedChannels = make(map[string]bool)
	}
	if setting.BlockedChannels == nil {
		setting.BlockedChannels = make(map[string]bool)
	}
	if setting.BlockedUsers == nil {
		setting.BlockedUsers = make(map[string]bool)
	}
	return setting
}

func (b *bot) saveSettings() error {
	b.mu.Lock()
	settings := make(map[string]guildSettings, len(b.settings))
	for guildID, setting := range b.settings {
		settings[guildID] = setting
	}
	path := b.settingsPath
	b.mu.Unlock()
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".yomiage-settings-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func (b *bot) guildSettings(guildID string) guildSettings {
	b.mu.Lock()
	defer b.mu.Unlock()
	setting := normalizeSettings(b.settings[guildID], b.speaker)
	return setting
}

func (b *bot) updateSettings(guildID string, update func(*guildSettings)) {
	b.mu.Lock()
	setting := normalizeSettings(b.settings[guildID], b.speaker)
	update(&setting)
	b.settings[guildID] = setting
	b.mu.Unlock()
	if err := b.saveSettings(); err != nil {
		log.Printf("⚠️ 設定の保存に失敗しました: %v", err)
	}
}

func (b *bot) setVolume(guildID string, volume float64) string {
	if volume < 0 || volume > 2 || math.IsNaN(volume) || math.IsInf(volume, 0) {
		return "音量は 0.0〜2.0 の範囲で指定してください。"
	}
	b.updateSettings(guildID, func(setting *guildSettings) { setting.Volume, setting.VolumeSet = volume, true })
	return fmt.Sprintf("音量を %.2f に設定しました。", volume)
}

func (b *bot) setSpeed(guildID string, speed float64) string {
	if speed < 0.5 || speed > 2 || math.IsNaN(speed) || math.IsInf(speed, 0) {
		return "速度は 0.5〜2.0 の範囲で指定してください。"
	}
	b.updateSettings(guildID, func(setting *guildSettings) { setting.Speed, setting.SpeedSet = speed, true })
	return fmt.Sprintf("読み上げ速度を %.2f に設定しました。", speed)
}

func (b *bot) configureChannel(guildID string, args []string) string {
	if len(args) != 2 || args[1] == "" {
		return "使い方: !channel allow|block|remove チャンネルID"
	}
	action, channelID := strings.ToLower(args[0]), args[1]
	if action != "allow" && action != "block" && action != "remove" {
		return "操作は allow / block / remove のいずれかで指定してください。"
	}
	b.updateSettings(guildID, func(setting *guildSettings) {
		switch action {
		case "allow":
			setting.AllowedChannels[channelID] = true
			delete(setting.BlockedChannels, channelID)
		case "block":
			setting.BlockedChannels[channelID] = true
			delete(setting.AllowedChannels, channelID)
		case "remove":
			delete(setting.AllowedChannels, channelID)
			delete(setting.BlockedChannels, channelID)
		}
	})
	return fmt.Sprintf("チャンネル %s の設定を %s にしました。", channelID, action)
}

func (b *bot) configureUser(guildID string, args []string) string {
	if len(args) != 2 || args[1] == "" {
		return "使い方: !user block|remove ユーザーID"
	}
	action, userID := strings.ToLower(args[0]), args[1]
	if action != "block" && action != "remove" {
		return "ユーザー操作は block / remove のいずれかで指定してください。"
	}
	b.updateSettings(guildID, func(setting *guildSettings) {
		if action == "block" {
			setting.BlockedUsers[userID] = true
		} else {
			delete(setting.BlockedUsers, userID)
		}
	})
	return fmt.Sprintf("ユーザー %s の設定を %s にしました。", userID, action)
}

func (b *bot) isExcluded(guildID, channelID, userID string) bool {
	setting := b.guildSettings(guildID)
	if setting.BlockedUsers[userID] || setting.BlockedChannels[channelID] {
		return true
	}
	if len(setting.AllowedChannels) > 0 && !setting.AllowedChannels[channelID] {
		return true
	}
	return false
}

func (b *bot) skip(guildID string) string {
	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	if current == nil {
		return "現在、このサーバーでは読み上げをしていません。"
	}
	current.mu.Lock()
	cancel := current.speechCancel
	current.mu.Unlock()
	if cancel == nil {
		return "現在読み上げ中の文章はありません。"
	}
	cancel()
	return "現在の読み上げをスキップしました。"
}

func (b *bot) joinForUser(guildID, userID, replyChannelID string) string {
	voiceState, err := b.session.State.VoiceState(guildID, userID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		return "先にボイスチャンネルへ参加してください。"
	}

	b.mu.Lock()
	if b.players[guildID] != nil {
		b.mu.Unlock()
		return "すでにボイスチャンネルに接続しています。"
	}
	b.mu.Unlock()

	// Discord APIへの接続中はロックを保持しない。接続が遅い場合でも、
	// ステータス表示や他サーバーの操作をブロックしないようにする。
	voice, err := b.session.ChannelVoiceJoin(guildID, voiceState.ChannelID, false, true)
	if err != nil {
		log.Printf("❌ ボイスチャンネルへの接続に失敗しました: %v", err)
		return "ボイスチャンネルに接続できませんでした。"
	}
	ctx, cancel := context.WithCancel(context.Background())
	current := &player{guildID: guildID, voice: voice, notifyChannelID: replyChannelID, queue: make(chan string, queueCapacity), ctx: ctx, cancel: cancel}
	b.mu.Lock()
	if b.players[guildID] != nil {
		b.mu.Unlock()
		cancel()
		if disconnectErr := voice.Disconnect(); disconnectErr != nil {
			log.Printf("⚠️ 重複したボイス接続の切断に失敗しました: %v", disconnectErr)
		}
		return "すでにボイスチャンネルに接続しています。"
	}
	b.players[guildID] = current
	b.mu.Unlock()

	go b.playQueue(guildID, current)
	return "読み上げを開始しました。"
}

func (b *bot) leaveForCommand(guildID string) string {
	if b.leave(guildID) {
		return "読み上げを終了しました。"
	}
	return "現在このサーバーでは読み上げをしていません。"
}

func (b *bot) setSpeaker(guildID string, speaker int) string {
	if speaker <= 0 {
		return "話者IDは1以上で指定してください。"
	}
	b.updateSettings(guildID, func(setting *guildSettings) { setting.Speaker = speaker })
	return fmt.Sprintf("このサーバーの話者IDを %d に設定しました。✨", speaker)
}

func (b *bot) guildSpeaker(guildID string) int {
	return b.guildSettings(guildID).Speaker
}

func (b *bot) enqueueText(guildID, text, origin string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "読み上げる文章が空です。"
	}
	if utf8.RuneCountInString(text) > maxMessageRunes {
		return fmt.Sprintf("%d 文字以内で指定してください。", maxMessageRunes)
	}

	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	if current == nil {
		return "先に `/join` でボイスチャンネルに接続してください。"
	}
	if current.enqueue(text) {
		return fmt.Sprintf("%s から読み上げをキューに追加しました。", origin)
	}
	select {
	case <-current.ctx.Done():
		return "現在の読み上げ処理が終了してから再度お試しください。"
	default:
		return "読み上げキューがいっぱいです。少し待ってから再度お試しください。"
	}
}

func (p *player) enqueue(text string) bool {
	p.queueMu.Lock()
	defer p.queueMu.Unlock()
	select {
	case p.queue <- text:
		p.queued = append(p.queued, text)
		return true
	default:
		return false
	}
}

func (p *player) dequeue(text string) {
	p.queueMu.Lock()
	defer p.queueMu.Unlock()
	if len(p.queued) > 0 {
		p.queued = p.queued[1:]
	}
}

func (b *bot) statusMessage(guildID string) string {
	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	setting := b.guildSettings(guildID)
	if current == nil {
		return fmt.Sprintf("🎧 現在、このサーバーでは読み上げを停止中です。話者ID: %d / 音量: %.2f / 速度: %.2f", setting.Speaker, setting.Volume, setting.Speed)
	}
	return fmt.Sprintf("🔊 読み上げ中です。話者ID: %d / 音量: %.2f / 速度: %.2f / 待機中: %d件", setting.Speaker, setting.Volume, setting.Speed, current.queueLength())
}

func (b *bot) queueMessage(guildID string) string {
	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	if current == nil {
		return "現在、このサーバーでは読み上げをしていません。"
	}
	current.queueMu.Lock()
	items := append([]string(nil), current.queued...)
	current.queueMu.Unlock()
	if len(items) == 0 {
		return "📚 読み上げキューには 0 件の待機があります。"
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = fmt.Sprintf("%d. %s", i+1, item)
	}
	result := fmt.Sprintf("📚 読み上げキュー（%d件）:\n%s", len(items), strings.Join(lines, "\n"))
	if utf8.RuneCountInString(result) > 1900 {
		return string([]rune(result)[:1900]) + "\n…"
	}
	return result
}

func (b *bot) clearQueue(guildID string) string {
	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	if current == nil {
		return "現在、このサーバーでは読み上げをしていません。"
	}
	current.queueMu.Lock()
	defer current.queueMu.Unlock()
	cleared := len(current.queued)
	for {
		select {
		case <-current.queue:
		default:
			current.queued = nil
			return fmt.Sprintf("待機中の読み上げを %d 件取り消しました。", cleared)
		}
	}
}

func (p *player) queueLength() int {
	p.queueMu.Lock()
	defer p.queueMu.Unlock()
	return len(p.queued)
}

func (b *bot) removeQueue(guildID string, index int) string {
	b.mu.Lock()
	current := b.players[guildID]
	b.mu.Unlock()
	if current == nil {
		return "現在、このサーバーでは読み上げをしていません。"
	}
	current.queueMu.Lock()
	defer current.queueMu.Unlock()
	if index < 1 || index > len(current.queued) {
		return fmt.Sprintf("キュー番号は 1〜%d で指定してください。", len(current.queued))
	}
	removed := current.queued[index-1]
	current.queued = append(current.queued[:index-1], current.queued[index:]...)
	for len(current.queue) > 0 {
		<-current.queue
	}
	for _, text := range current.queued {
		current.queue <- text
	}
	return fmt.Sprintf("キュー %d 番を削除しました: %s", index, removed)
}

func (b *bot) leave(guildID string) bool {
	b.mu.Lock()
	current := b.players[guildID]
	delete(b.players, guildID)
	b.mu.Unlock()
	if current == nil {
		return false
	}
	current.cancel()
	if err := current.voice.Disconnect(); err != nil {
		log.Printf("⚠️ ボイスチャンネルの切断に失敗しました: %v", err)
	}
	return true
}

func (b *bot) stopAll() {
	b.mu.Lock()
	players := b.players
	b.players = make(map[string]*player)
	b.mu.Unlock()
	for _, current := range players {
		current.cancel()
		if err := current.voice.Disconnect(); err != nil {
			log.Printf("⚠️ ボイスチャンネルの切断に失敗しました: %v", err)
		}
	}
}

func (b *bot) playQueue(guildID string, current *player) {
	for {
		select {
		case <-current.ctx.Done():
			return
		case text := <-current.queue:
			current.dequeue(text)
			for {
				current.mu.Lock()
				paused := current.paused
				current.mu.Unlock()
				if !paused {
					break
				}
				select {
				case <-current.ctx.Done():
					return
				case <-time.After(200 * time.Millisecond):
				}
			}
			var err error
			for attempt := 0; attempt < 2; attempt++ {
				err = b.speak(current, text)
				if err == nil || errors.Is(err, context.Canceled) {
					break
				}
				log.Printf("⚠️ Guild %s のメッセージ取得に失敗しました（試行 %d）: %v", guildID, attempt+1, err)
				select {
				case <-current.ctx.Done():
					return
				case <-time.After(500 * time.Millisecond):
				}
			}
			if err != nil && !errors.Is(err, context.Canceled) && current.notifyChannelID != "" {
				b.reply(current.notifyChannelID, guildID, "音声の生成または再生に失敗しました。VOICEVOXとffmpegの状態を確認してください。", false)
			}
		}
	}
}

func (b *bot) speak(current *player, text string) error {
	ctx, cancel := context.WithTimeout(current.ctx, 45*time.Second)
	defer cancel()
	current.mu.Lock()
	current.speechCancel = cancel
	current.mu.Unlock()
	defer func() {
		current.mu.Lock()
		current.speechCancel = nil
		current.mu.Unlock()
	}()

	setting := b.guildSettings(current.guildID)
	audio, err := b.synthesizeWithSettings(ctx, text, setting)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp("", "yomiage-*.wav")
	if err != nil {
		return fmt.Errorf("create audio file: %w", err)
	}
	fileName := file.Name()
	defer os.Remove(fileName)
	if _, err := file.Write(audio); err != nil {
		file.Close()
		return fmt.Errorf("write audio file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close audio file: %w", err)
	}

	encodeOptions := dca.StdEncodeOptions
	encodeOptions.RawOutput = true
	encodeSession, err := dca.EncodeFile(fileName, encodeOptions)
	if err != nil {
		return fmt.Errorf("encode audio (is ffmpeg installed?): %w", err)
	}
	defer encodeSession.Cleanup()

	readyTicker := time.NewTicker(100 * time.Millisecond)
	defer readyTicker.Stop()
	for {
		current.voice.RLock()
		ready := current.voice.Ready
		current.voice.RUnlock()
		if ready {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-readyTicker.C:
		}
	}

	if err := current.voice.Speaking(true); err != nil {
		return fmt.Errorf("set speaking state: %w", err)
	}
	defer current.voice.Speaking(false)

	frameTicker := time.NewTicker(encodeSession.FrameDuration())
	defer frameTicker.Stop()
	for {
		frame, err := encodeSession.OpusFrame()
		if errors.Is(err, io.EOF) {
			return encodeSession.Error()
		}
		if err != nil {
			return fmt.Errorf("read Opus frame: %w", err)
		}
		select {
		case current.voice.OpusSend <- frame:
		case <-ctx.Done():
			return ctx.Err()
		}
		select {
		case <-frameTicker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *bot) synthesizeWithSpeaker(ctx context.Context, text string, speakerID int) ([]byte, error) {
	return b.synthesizeWithSettings(ctx, text, guildSettings{Speaker: speakerID, Volume: defaultVolume, Speed: defaultSpeed})
}

func (b *bot) synthesizeWithSettings(ctx context.Context, text string, setting guildSettings) ([]byte, error) {
	setting = normalizeSettings(setting, b.speaker)
	speakerID := setting.Speaker
	queryURL := b.ttsURL + "/audio_query?text=" + url.QueryEscape(text) + "&speaker=" + strconv.Itoa(speakerID)
	queryRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create audio query: %w", err)
	}
	queryResponse, err := b.client.Do(queryRequest)
	if err != nil {
		return nil, fmt.Errorf("VOICEVOX audio query: %w", err)
	}
	defer queryResponse.Body.Close()
	if queryResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VOICEVOX audio query returned %s", queryResponse.Status)
	}

	var query map[string]any
	if err := json.NewDecoder(queryResponse.Body).Decode(&query); err != nil {
		return nil, fmt.Errorf("decode audio query: %w", err)
	}
	query["volumeScale"] = setting.Volume
	query["speedScale"] = setting.Speed
	queryBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("encode audio query: %w", err)
	}

	synthesisURL := b.ttsURL + "/synthesis?speaker=" + strconv.Itoa(speakerID)
	synthesisRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, synthesisURL, bytes.NewReader(queryBody))
	if err != nil {
		return nil, fmt.Errorf("create synthesis request: %w", err)
	}
	synthesisRequest.Header.Set("Content-Type", "application/json")
	synthesisResponse, err := b.client.Do(synthesisRequest)
	if err != nil {
		return nil, fmt.Errorf("VOICEVOX synthesis: %w", err)
	}
	defer synthesisResponse.Body.Close()
	if synthesisResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VOICEVOX synthesis returned %s", synthesisResponse.Status)
	}
	return io.ReadAll(synthesisResponse.Body)
}

func (b *bot) synthesize(ctx context.Context, text string) ([]byte, error) {
	return b.synthesizeWithSpeaker(ctx, text, b.speaker)
}

func (b *bot) reply(channelID, guildID, content string, ok bool) {
	if _, err := b.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{newEmbed("📣 読み上げBOT", content, ok, b.embedFieldsForGuild(guildID))},
		Components: []discordgo.MessageComponent{newActionRow()},
	}); err != nil {
		log.Printf("⚠️ メッセージ送信に失敗しました: %v", err)
	}
}

func newEmbed(title, description string, ok bool, fields []*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	color := 0x2ECC71
	if !ok {
		color = 0xE74C3C
	}
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
	}
}

func newActionRow() discordgo.MessageComponent {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: "join_button",
				Label:    "接続",
				Style:    discordgo.SuccessButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "🎙️"},
			},
			discordgo.Button{
				CustomID: "leave_button",
				Label:    "終了",
				Style:    discordgo.DangerButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "🛑"},
			},
			discordgo.Button{
				CustomID: "status_button",
				Label:    "状態",
				Style:    discordgo.SecondaryButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
			},
			discordgo.Button{
				CustomID: "queue_button",
				Label:    "キュー",
				Style:    discordgo.SecondaryButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "📚"},
			},
			discordgo.Button{
				CustomID: "clear_button",
				Label:    "全消去",
				Style:    discordgo.DangerButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "🧹"},
			},
		},
	}
}

func (b *bot) embedFieldsForGuild(guildID string) []*discordgo.MessageEmbedField {
	fields := []*discordgo.MessageEmbedField{{
		Name:   "サーバー",
		Value:  guildID,
		Inline: true,
	}}
	if guildID != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "話者ID",
			Value:  strconv.Itoa(b.guildSpeaker(guildID)),
			Inline: true,
		})
	}
	return fields
}

func (b *bot) registerCommands() error {
	guildIDs := strings.TrimSpace(os.Getenv("DISCORD_GUILD_ID"))
	if guildIDs != "" {
		for _, guildID := range strings.Split(guildIDs, ",") {
			guildID = strings.TrimSpace(guildID)
			if guildID == "" {
				continue
			}
			for _, command := range applicationCommands {
				if _, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, guildID, command); err != nil {
					return fmt.Errorf("create guild application command for %s: %w", guildID, err)
				}
			}
		}
		return nil
	}

	for _, command := range applicationCommands {
		if _, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", command); err != nil {
			return fmt.Errorf("create global application command: %w", err)
		}
	}
	return nil
}
