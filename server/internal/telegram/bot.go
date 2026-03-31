package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	appURL string
}

func NewBot(token, appURL string) (*Bot, error) {
	if token == "" {
		log.Println("telegram: no bot token provided, bot disabled")
		return &Bot{appURL: appURL}, nil
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram: failed to create bot: %w", err)
	}

	log.Printf("telegram: authorized on account %s", api.Self.UserName)
	return &Bot{api: api, appURL: appURL}, nil
}

func (b *Bot) Start() {
	if b.api == nil {
		return
	}

	// Remove any existing webhook so long polling works
	if _, err := b.api.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		log.Printf("telegram: failed to clear webhook: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	log.Println("telegram: bot is listening for updates...")

	for update := range updates {
		log.Printf("telegram: received update: %+v", update.UpdateID)

		if update.Message == nil {
			continue
		}

		log.Printf("telegram: message from %s: %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				b.handleStart(update.Message)
			default:
				log.Printf("telegram: unknown command: %s", update.Message.Command())
			}
		}
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	webAppURL := strings.TrimRight(b.appURL, "/")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Open Menu", webAppURL),
		),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, "Welcome to Xpressgo! Tap the button below to browse the menu and place your order.")
	reply.ReplyMarkup = keyboard
	if _, err := b.api.Send(reply); err != nil {
		log.Printf("telegram: failed to send start reply: %v", err)
	}
}
