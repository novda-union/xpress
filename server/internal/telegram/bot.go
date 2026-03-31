package telegram

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
)

type Bot struct {
	api        *tgbotapi.BotAPI
	appURL     string
	userRepo   *repository.UserRepo
	verifyRepo *repository.PhoneVerificationRepo
}

func NewBot(token, appURL string, userRepo *repository.UserRepo, verifyRepo *repository.PhoneVerificationRepo) (*Bot, error) {
	if token == "" {
		log.Println("telegram: no bot token provided, bot disabled")
		return &Bot{appURL: appURL}, nil
	}

	api, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, newIPv4HTTPClient())
	if err != nil {
		return nil, fmt.Errorf("telegram: failed to create bot: %w", err)
	}

	log.Printf("telegram: authorized on account %s", api.Self.UserName)
	return &Bot{
		api:        api,
		appURL:     appURL,
		userRepo:   userRepo,
		verifyRepo: verifyRepo,
	}, nil
}

func newIPv4HTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	transport.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, "tcp4", addr)
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

func (b *Bot) Start() {
	if b.api == nil {
		return
	}

	if _, err := b.api.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		log.Printf("telegram: failed to clear webhook: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	log.Println("telegram: bot is listening for updates...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Contact share arrives as a special message type
		if update.Message.Contact != nil {
			b.handleContact(update.Message)
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				b.handleStart(update.Message)
			default:
				log.Printf("telegram: unknown command: %s", update.Message.Command())
			}
			continue
		}

		// Treat any 4-digit text as a verification code attempt
		text := strings.TrimSpace(update.Message.Text)
		if len(text) == 4 {
			b.handleCodeEntry(update.Message, text)
		}
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Share My Phone Number"),
		),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true

	reply := tgbotapi.NewMessage(msg.Chat.ID,
		"Welcome to Xpressgo! 👋\n\nTo get started, please share your phone number so we can verify your identity.")
	reply.ReplyMarkup = keyboard

	if _, err := b.api.Send(reply); err != nil {
		log.Printf("telegram: failed to send start reply: %v", err)
	}
}

func (b *Bot) handleContact(msg *tgbotapi.Message) {
	contact := msg.Contact

	// Make sure the user is sharing their own number
	if contact.UserID != msg.From.ID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Please share your own phone number, not someone else's.")
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, err := b.api.Send(reply); err != nil {
			log.Printf("telegram: failed to send reply: %v", err)
		}
		return
	}

	phone := contact.PhoneNumber
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	// Upsert the user record with the phone we just received
	ctx := context.Background()
	user := &model.User{
		TelegramID: msg.From.ID,
		Phone:      phone,
		FirstName:  msg.From.FirstName,
		LastName:   msg.From.LastName,
		Username:   msg.From.UserName,
	}
	if err := b.userRepo.Upsert(ctx, user); err != nil {
		log.Printf("telegram: failed to upsert user: %v", err)
	}

	// Generate a 4-digit verification code
	code := fmt.Sprintf("%04d", rand.Intn(10000))
	expiresAt := time.Now().Add(10 * time.Minute)

	if err := b.verifyRepo.Save(ctx, msg.From.ID, phone, code, expiresAt); err != nil {
		log.Printf("telegram: failed to save verification code: %v", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Something went wrong. Please send /start to try again.")
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		b.api.Send(reply) //nolint:errcheck
		return
	}

	// Remove the contact keyboard first
	ack := tgbotapi.NewMessage(msg.Chat.ID, "📱 Got your number!")
	ack.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	b.api.Send(ack) //nolint:errcheck

	// Send the verification code message
	codeMsg := tgbotapi.NewMessage(msg.Chat.ID,
		fmt.Sprintf("Enter the SMS code we sent you.\n\nFor demo, your code is: *%s*", code))
	codeMsg.ParseMode = "Markdown"
	if _, err := b.api.Send(codeMsg); err != nil {
		log.Printf("telegram: failed to send verification code: %v", err)
	}
}

func (b *Bot) handleCodeEntry(msg *tgbotapi.Message, code string) {
	ctx := context.Background()
	_, err := b.verifyRepo.Consume(ctx, msg.From.ID, code)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ Wrong code or it has expired. Please send /start to request a new one.")
		if _, err := b.api.Send(reply); err != nil {
			log.Printf("telegram: failed to send error reply: %v", err)
		}
		return
	}

	webAppURL := strings.TrimRight(b.appURL, "/")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Open Menu 🍽️", webAppURL),
		),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID,
		"✅ Phone verified! You're all set.\n\nTap below to start ordering.")
	reply.ReplyMarkup = keyboard
	if _, err := b.api.Send(reply); err != nil {
		log.Printf("telegram: failed to send success reply: %v", err)
	}
}
