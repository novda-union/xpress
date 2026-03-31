package telegram

import (
	"strings"
	"testing"
	"time"

	"github.com/xpressgo/server/internal/model"
)

func TestFormatBranchOrderStatusMessage(t *testing.T) {
	order := &model.Order{
		OrderNumber: 42,
		Status:      "accepted",
		TotalPrice:  55000,
		Items: []model.OrderItem{
			{
				ItemName: "Latte",
				Quantity: 2,
				Modifiers: []model.OrderItemModifier{
					{ModifierName: "Vanilla"},
				},
			},
		},
	}

	text := formatBranchOrderStatusMessage("Chilonzor", order)

	checks := []string{
		"Order #42 at Chilonzor",
		"Status: accepted",
		"2x Latte (Vanilla)",
		"Total: 55,000 UZS",
	}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Fatalf("expected message to contain %q, got %q", want, text)
		}
	}
}

func TestFormatBranchOrderStatusMessageRejectedIncludesReason(t *testing.T) {
	order := &model.Order{
		OrderNumber:     99,
		Status:          "rejected",
		TotalPrice:      12000,
		RejectionReason: "Out of stock",
		Items: []model.OrderItem{
			{
				ItemName: "Americano",
				Quantity: 1,
			},
		},
	}

	text := formatBranchOrderStatusMessage("Yunusobod", order)

	checks := []string{
		"Status: rejected",
		"Reason: Out of stock",
		"1x Americano",
		"Total: 12,000 UZS",
	}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Fatalf("expected rejected message to contain %q, got %q", want, text)
		}
	}
}

func TestFormatBranchDailySummaryMessage(t *testing.T) {
	summary := &model.BranchDailyOrderSummary{
		BranchName:            "Chilonzor",
		LocalDate:             time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		TotalOrders:           10,
		PendingOrders:         1,
		AcceptedOrders:        2,
		PreparingOrders:       3,
		ReadyOrders:           1,
		PickedUpOrders:        2,
		RejectedOrders:        1,
		CancelledOrders:       0,
		TotalCreatedOrderSum:  120000,
		TotalPickedUpOrderSum: 90000,
	}

	text := formatBranchDailySummaryMessage(summary)

	checks := []string{
		"Daily summary for Chilonzor",
		"Date: 2026-04-01",
		"Total orders: 10",
		"Pending: 1",
		"Accepted: 2",
		"Preparing: 3",
		"Ready: 1",
		"Picked up: 2",
		"Rejected: 1",
		"Cancelled: 0",
		"Created total: 120,000 UZS",
		"Picked up total: 90,000 UZS",
	}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Fatalf("expected summary to contain %q, got %q", want, text)
		}
	}
}
