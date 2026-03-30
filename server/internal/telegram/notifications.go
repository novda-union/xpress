package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xpressgo/server/internal/model"
)

// SendOrderStatusToUser sends order status update to user via DM.
func (b *Bot) SendOrderStatusToUser(telegramID int64, order *model.Order, storeName string) {
	if b.api == nil {
		log.Printf("telegram: bot disabled, would notify user %d about order #%d status: %s", telegramID, order.OrderNumber, order.Status)
		return
	}

	var text string
	switch order.Status {
	case "accepted":
		text = fmt.Sprintf("Your order #%d at %s has been accepted! Preparing now.", order.OrderNumber, storeName)
	case "preparing":
		text = fmt.Sprintf("Your order #%d is being prepared.", order.OrderNumber)
	case "ready":
		text = fmt.Sprintf("Your order #%d is ready for pickup!", order.OrderNumber)
	case "rejected":
		text = fmt.Sprintf("Sorry, %s couldn't accept your order #%d.", storeName, order.OrderNumber)
		if order.RejectionReason != "" {
			text += fmt.Sprintf(" Reason: %s", order.RejectionReason)
		}
	default:
		return
	}

	msg := tgbotapi.NewMessage(telegramID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("telegram: failed to send message to user %d: %v", telegramID, err)
	}
}

func (b *Bot) SendNewOrderToChat(groupChatID int64, locationName string, order *model.Order) {
	if b.api == nil {
		log.Printf("telegram: bot disabled, would notify group %d about new order #%d", groupChatID, order.OrderNumber)
		return
	}

	var items []string
	for _, item := range order.Items {
		entry := fmt.Sprintf("%dx %s", item.Quantity, item.ItemName)
		if len(item.Modifiers) > 0 {
			var mods []string
			for _, m := range item.Modifiers {
				mods = append(mods, m.ModifierName)
			}
			entry += fmt.Sprintf(" (%s)", strings.Join(mods, ", "))
		}
		items = append(items, entry)
	}

	text := fmt.Sprintf(
		"New order #%d at %s!\n%s\nCustomer arrives in ~%d min\nTotal: %s UZS",
		order.OrderNumber,
		locationName,
		strings.Join(items, "\n"),
		order.ETAMinutes,
		formatPrice(order.TotalPrice),
	)

	msg := tgbotapi.NewMessage(groupChatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("telegram: failed to send message to group %d: %v", groupChatID, err)
	}
}

func (b *Bot) SendOrderCancelledToChat(groupChatID int64, locationName string, orderNumber int) {
	if b.api == nil {
		log.Printf("telegram: bot disabled, would notify group %d about cancelled order #%d", groupChatID, orderNumber)
		return
	}

	text := fmt.Sprintf("Order #%d at %s was cancelled by the customer.", orderNumber, locationName)
	msg := tgbotapi.NewMessage(groupChatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("telegram: failed to send cancellation to group %d: %v", groupChatID, err)
	}
}

// SendNewOrderToStore keeps the older store-scoped helper available.
func (b *Bot) SendNewOrderToStore(groupChatID int64, order *model.Order) {
	b.SendNewOrderToChat(groupChatID, "your store", order)
}

// SendOrderCancelledToStore keeps the older store-scoped helper available.
func (b *Bot) SendOrderCancelledToStore(groupChatID int64, orderNumber int) {
	b.SendOrderCancelledToChat(groupChatID, "your store", orderNumber)
}

func formatPrice(price int64) string {
	s := fmt.Sprintf("%d", price)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
