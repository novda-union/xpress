# Telegram Order Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add branch-specific Telegram notifications for every admin order status transition and a branch-specific end-of-day summary in `Asia/Tashkent`, using the existing bot.

**Architecture:** Keep notification triggering in the Go backend. Extend the order handler and Telegram package for status-change sends, add repository support for branch daily summary aggregation plus delivery-log persistence, and start an in-process summary scheduler from server startup. Daily summaries are de-duplicated with a persisted send log keyed by branch, summary date, and notification type.

**Tech Stack:** Go, Echo, pgx, PostgreSQL migrations, existing Telegram bot package, repo `Makefile`

---

## File Structure

**Modify**
- `server/internal/handler/order_handler.go`
  Responsibility: trigger Telegram branch notifications after successful admin status updates.
- `server/internal/repository/order_repo.go`
  Responsibility: add branch summary aggregation queries.
- `server/internal/repository/branch_repo.go`
  Responsibility: add branch listing for summary scheduling, filtered to active branches with chat IDs when needed.
- `server/internal/telegram/notifications.go`
  Responsibility: add status-change group message formatter and daily summary formatter/sender.
- `server/internal/telegram/bot.go`
  Responsibility: expose scheduler-friendly send surface only if needed by package internals.
- `server/cmd/server/main.go`
  Responsibility: wire new repos/services and start the daily summary scheduler.

**Create**
- `server/migrations/000004_daily_notification_deliveries.up.sql`
  Responsibility: create persistence for branch daily-summary send de-duplication.
- `server/migrations/000004_daily_notification_deliveries.down.sql`
  Responsibility: drop persistence for branch daily-summary send de-duplication.
- `server/internal/model/notification.go`
  Responsibility: define branch daily summary and delivery-log models.
- `server/internal/repository/notification_delivery_repo.go`
  Responsibility: read/write daily summary delivery log rows.
- `server/internal/service/order_notification_service.go`
  Responsibility: orchestrate branch status notifications and daily summary generation.
- `server/internal/service/order_notification_scheduler.go`
  Responsibility: run the `Asia/Tashkent` ticker and dispatch pending branch summaries.
- `server/internal/service/order_notification_service_test.go`
  Responsibility: unit tests for message-trigger rules and summary scheduling helpers.
- `server/internal/telegram/notifications_test.go`
  Responsibility: unit tests for Telegram message formatting.

**Verification**
- `docs/superpowers/specs/2026-04-01-telegram-order-notifications-design.md`
  Responsibility: source-of-truth design reference while implementing.

---

### Task 1: Add Daily Summary Delivery Persistence

**Files:**
- Create: `server/migrations/000004_daily_notification_deliveries.up.sql`
- Create: `server/migrations/000004_daily_notification_deliveries.down.sql`
- Create: `server/internal/model/notification.go`
- Create: `server/internal/repository/notification_delivery_repo.go`
- Test: `server/internal/service/order_notification_service_test.go`

- [ ] **Step 1: Write the failing repository/service test shape**

```go
func TestDailySummaryDeliveryKey(t *testing.T) {
	record := model.NotificationDelivery{
		NotificationType: "branch_daily_summary",
		BranchID:         "branch-1",
		LocalDate:        "2026-04-01",
	}

	if record.NotificationType != "branch_daily_summary" {
		t.Fatalf("expected notification type to be preserved")
	}
}
```

- [ ] **Step 2: Run the focused server test to verify the new files are missing**

Run: `cd server && go test ./internal/service -run TestDailySummaryDeliveryKey`

Expected: FAIL with missing `model.NotificationDelivery` and missing test file/package symbols.

- [ ] **Step 3: Add the delivery-log migration**

```sql
CREATE TABLE notification_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_type TEXT NOT NULL,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    local_date DATE NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (notification_type, branch_id, local_date)
);

CREATE INDEX notification_deliveries_branch_date_idx
    ON notification_deliveries (branch_id, local_date);
```

Down migration:

```sql
DROP TABLE IF EXISTS notification_deliveries;
```

- [ ] **Step 4: Add the delivery model and repository**

`server/internal/model/notification.go`

```go
package model

import "time"

type NotificationDelivery struct {
	ID               string    `json:"id"`
	NotificationType string    `json:"notification_type"`
	BranchID         string    `json:"branch_id"`
	LocalDate        string    `json:"local_date"`
	SentAt           time.Time `json:"sent_at"`
	CreatedAt        time.Time `json:"created_at"`
}
```

`server/internal/repository/notification_delivery_repo.go`

```go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationDeliveryRepo struct {
	db *pgxpool.Pool
}

func NewNotificationDeliveryRepo(db *pgxpool.Pool) *NotificationDeliveryRepo {
	return &NotificationDeliveryRepo{db: db}
}

func (r *NotificationDeliveryRepo) Exists(ctx context.Context, notificationType, branchID, localDate string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM notification_deliveries
			WHERE notification_type = $1 AND branch_id = $2 AND local_date = $3::date
		)
	`, notificationType, branchID, localDate).Scan(&exists)
	return exists, err
}

func (r *NotificationDeliveryRepo) Create(ctx context.Context, notificationType, branchID, localDate string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notification_deliveries (notification_type, branch_id, local_date)
		VALUES ($1, $2, $3::date)
	`, notificationType, branchID, localDate)
	return err
}
```

- [ ] **Step 5: Run the focused test again**

Run: `cd server && go test ./internal/service -run TestDailySummaryDeliveryKey`

Expected: PASS or compile failure moves to the next missing dependency instead of model/repo shape.

- [ ] **Step 6: Commit**

```bash
git add server/migrations/000004_daily_notification_deliveries.up.sql server/migrations/000004_daily_notification_deliveries.down.sql server/internal/model/notification.go server/internal/repository/notification_delivery_repo.go server/internal/service/order_notification_service_test.go
git commit -m "feat(notifications): add daily summary delivery tracking"
```

### Task 2: Add Branch Daily Summary Queries

**Files:**
- Modify: `server/internal/repository/order_repo.go`
- Modify: `server/internal/repository/branch_repo.go`
- Modify: `server/internal/model/notification.go`
- Test: `server/internal/service/order_notification_service_test.go`

- [ ] **Step 1: Write the failing summary-shape test**

```go
func TestBuildDailySummaryIncludesPickedUpTotals(t *testing.T) {
	summary := model.BranchDailyOrderSummary{
		BranchID:              "branch-1",
		BranchName:            "Chilonzor",
		LocalDate:             "2026-04-01",
		TotalOrders:           10,
		PickedUpOrders:        7,
		TotalCreatedOrderSum:  120000,
		TotalPickedUpOrderSum: 90000,
	}

	if summary.PickedUpOrders != 7 {
		t.Fatalf("expected picked up order count")
	}
}
```

- [ ] **Step 2: Run the focused test to verify missing summary types**

Run: `cd server && go test ./internal/service -run TestBuildDailySummaryIncludesPickedUpTotals`

Expected: FAIL with missing `model.BranchDailyOrderSummary`.

- [ ] **Step 3: Extend models with summary types**

Add to `server/internal/model/notification.go`:

```go
type BranchDailyOrderSummary struct {
	BranchID              string `json:"branch_id"`
	BranchName            string `json:"branch_name"`
	LocalDate             string `json:"local_date"`
	TotalOrders           int    `json:"total_orders"`
	PendingOrders         int    `json:"pending_orders"`
	AcceptedOrders        int    `json:"accepted_orders"`
	PreparingOrders       int    `json:"preparing_orders"`
	ReadyOrders           int    `json:"ready_orders"`
	PickedUpOrders        int    `json:"picked_up_orders"`
	RejectedOrders        int    `json:"rejected_orders"`
	CancelledOrders       int    `json:"cancelled_orders"`
	TotalCreatedOrderSum  int64  `json:"total_created_order_sum"`
	TotalPickedUpOrderSum int64  `json:"total_picked_up_order_sum"`
}
```

- [ ] **Step 4: Add repository helpers for branch listing and summary aggregation**

Add to `server/internal/repository/branch_repo.go`:

```go
func (r *BranchRepo) ListActiveWithTelegramChat(ctx context.Context) ([]model.Branch, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, store_id, name, address, lat, lng, banner_image_url, telegram_group_chat_id, is_active, created_at, updated_at
		FROM branches
		WHERE is_active = true AND telegram_group_chat_id IS NOT NULL
		ORDER BY created_at ASC, name ASC
	`)
	// scan into []model.Branch
}
```

Add to `server/internal/repository/order_repo.go`:

```go
func (r *OrderRepo) GetBranchDailySummary(ctx context.Context, branchID string, startUTC, endUTC time.Time) (*model.BranchDailyOrderSummary, error) {
	summary := &model.BranchDailyOrderSummary{BranchID: branchID}
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) AS total_orders,
			COUNT(*) FILTER (WHERE status = 'pending') AS pending_orders,
			COUNT(*) FILTER (WHERE status = 'accepted') AS accepted_orders,
			COUNT(*) FILTER (WHERE status = 'preparing') AS preparing_orders,
			COUNT(*) FILTER (WHERE status = 'ready') AS ready_orders,
			COUNT(*) FILTER (WHERE status = 'picked_up') AS picked_up_orders,
			COUNT(*) FILTER (WHERE status = 'rejected') AS rejected_orders,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_orders,
			COALESCE(SUM(total_price), 0) AS total_created_order_sum,
			COALESCE(SUM(total_price) FILTER (WHERE status = 'picked_up'), 0) AS total_picked_up_order_sum
		FROM orders
		WHERE branch_id = $1 AND created_at >= $2 AND created_at < $3
	`, branchID, startUTC, endUTC).Scan(
		&summary.TotalOrders,
		&summary.PendingOrders,
		&summary.AcceptedOrders,
		&summary.PreparingOrders,
		&summary.ReadyOrders,
		&summary.PickedUpOrders,
		&summary.RejectedOrders,
		&summary.CancelledOrders,
		&summary.TotalCreatedOrderSum,
		&summary.TotalPickedUpOrderSum,
	)
	return summary, err
}
```

- [ ] **Step 5: Run the focused test again**

Run: `cd server && go test ./internal/service -run TestBuildDailySummaryIncludesPickedUpTotals`

Expected: PASS or fail later in service code, not on missing summary model shape.

- [ ] **Step 6: Commit**

```bash
git add server/internal/model/notification.go server/internal/repository/order_repo.go server/internal/repository/branch_repo.go server/internal/service/order_notification_service_test.go
git commit -m "feat(notifications): add branch daily summary queries"
```

### Task 3: Extend Telegram Message Formatting

**Files:**
- Modify: `server/internal/telegram/notifications.go`
- Test: `server/internal/telegram/notifications_test.go`

- [ ] **Step 1: Write the failing formatter tests**

```go
func TestFormatBranchStatusMessageRejectedIncludesReason(t *testing.T) {
	order := &model.Order{
		OrderNumber:     42,
		Status:          "rejected",
		TotalPrice:      55000,
		RejectionReason: "Out of stock",
		Items: []model.OrderItem{
			{ItemName: "Latte", Quantity: 2},
		},
	}

	text := formatBranchOrderStatusMessage("Chilonzor", order)

	if !strings.Contains(text, "Out of stock") {
		t.Fatalf("expected rejection reason in message: %s", text)
	}
}
```

```go
func TestFormatBranchDailySummaryIncludesCounts(t *testing.T) {
	summary := &model.BranchDailyOrderSummary{
		BranchName:            "Chilonzor",
		LocalDate:             "2026-04-01",
		TotalOrders:           10,
		PickedUpOrders:        7,
		RejectedOrders:        1,
		CancelledOrders:       1,
		TotalCreatedOrderSum:  120000,
		TotalPickedUpOrderSum: 90000,
	}

	text := formatBranchDailySummaryMessage(summary)

	if !strings.Contains(text, "Total orders: 10") {
		t.Fatalf("expected total orders line: %s", text)
	}
}
```

- [ ] **Step 2: Run the Telegram package test**

Run: `cd server && go test ./internal/telegram -run 'TestFormatBranch(StatusMessageRejectedIncludesReason|DailySummaryIncludesCounts)'`

Expected: FAIL with missing formatter helpers.

- [ ] **Step 3: Add formatter helpers and send methods**

Add to `server/internal/telegram/notifications.go`:

```go
func (b *Bot) SendOrderStatusToChat(groupChatID int64, branchName string, order *model.Order) {
	if b.api == nil {
		log.Printf("telegram: bot disabled, would notify group %d about order #%d status %s", groupChatID, order.OrderNumber, order.Status)
		return
	}

	msg := tgbotapi.NewMessage(groupChatID, formatBranchOrderStatusMessage(branchName, order))
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("telegram: failed to send status message to group %d: %v", groupChatID, err)
	}
}

func (b *Bot) SendBranchDailySummary(groupChatID int64, summary *model.BranchDailyOrderSummary) {
	if b.api == nil {
		log.Printf("telegram: bot disabled, would notify group %d about daily summary for branch %s", groupChatID, summary.BranchID)
		return
	}

	msg := tgbotapi.NewMessage(groupChatID, formatBranchDailySummaryMessage(summary))
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("telegram: failed to send daily summary to group %d: %v", groupChatID, err)
	}
}
```

Also add unexported helpers:

```go
func formatBranchOrderStatusMessage(branchName string, order *model.Order) string
func formatBranchDailySummaryMessage(summary *model.BranchDailyOrderSummary) string
```

- [ ] **Step 4: Re-run the Telegram package test**

Run: `cd server && go test ./internal/telegram -run 'TestFormatBranch(StatusMessageRejectedIncludesReason|DailySummaryIncludesCounts)'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/internal/telegram/notifications.go server/internal/telegram/notifications_test.go
git commit -m "feat(telegram): add branch status and summary messages"
```

### Task 4: Add Notification Orchestration Service

**Files:**
- Create: `server/internal/service/order_notification_service.go`
- Modify: `server/internal/service/order_notification_service_test.go`
- Modify: `server/internal/model/notification.go`

- [ ] **Step 1: Write the failing service orchestration tests**

```go
func TestShouldNotifyBranchStatusTransitions(t *testing.T) {
	for _, status := range []string{"accepted", "preparing", "ready", "picked_up", "rejected"} {
		if !shouldNotifyBranchStatus(status) {
			t.Fatalf("expected status %s to trigger notification", status)
		}
	}

	if shouldNotifyBranchStatus("cancelled") {
		t.Fatalf("did not expect cancelled to trigger admin branch status notification")
	}
}
```

```go
func TestSummaryWindowForDateUsesTashkentMidnight(t *testing.T) {
	startUTC, endUTC, err := summaryWindowForDate("2026-04-01")
	if err != nil {
		t.Fatal(err)
	}
	if !endUTC.After(startUTC) {
		t.Fatalf("expected end to be after start")
	}
}
```

- [ ] **Step 2: Run the service package tests**

Run: `cd server && go test ./internal/service -run 'Test(ShouldNotifyBranchStatusTransitions|SummaryWindowForDateUsesTashkentMidnight)'`

Expected: FAIL with missing helper functions and service file.

- [ ] **Step 3: Implement the orchestration service**

Create `server/internal/service/order_notification_service.go` with:

```go
package service

import (
	"context"
	"log"
	"time"

	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
	"github.com/xpressgo/server/internal/telegram"
)

const branchDailySummaryNotificationType = "branch_daily_summary"

type OrderNotificationService struct {
	branchRepo             *repository.BranchRepo
	orderRepo              *repository.OrderRepo
	deliveryRepo           *repository.NotificationDeliveryRepo
	bot                    *telegram.Bot
	location               *time.Location
}

func NewOrderNotificationService(branchRepo *repository.BranchRepo, orderRepo *repository.OrderRepo, deliveryRepo *repository.NotificationDeliveryRepo, bot *telegram.Bot) *OrderNotificationService {
	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		loc = time.FixedZone("Asia/Tashkent", 5*60*60)
	}
	return &OrderNotificationService{
		branchRepo:   branchRepo,
		orderRepo:    orderRepo,
		deliveryRepo: deliveryRepo,
		bot:          bot,
		location:     loc,
	}
}

func shouldNotifyBranchStatus(status string) bool {
	switch status {
	case "accepted", "preparing", "ready", "picked_up", "rejected":
		return true
	default:
		return false
	}
}

func summaryWindowForDate(localDate string) (time.Time, time.Time, error) {
	// parse `2006-01-02` in Asia/Tashkent and return UTC boundaries
}

func (s *OrderNotificationService) NotifyBranchOrderStatus(ctx context.Context, order *model.Order) {
	// load branch, guard nil chat id, check shouldNotifyBranchStatus, send via bot
}

func (s *OrderNotificationService) SendDailySummaryForBranch(ctx context.Context, branch model.Branch, localDate string) error {
	// de-dup, aggregate, format, send, persist delivery record
}
```

- [ ] **Step 4: Re-run the focused service tests**

Run: `cd server && go test ./internal/service -run 'Test(ShouldNotifyBranchStatusTransitions|SummaryWindowForDateUsesTashkentMidnight)'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/internal/service/order_notification_service.go server/internal/service/order_notification_service_test.go server/internal/model/notification.go
git commit -m "feat(notifications): add order notification service"
```

### Task 5: Wire Admin Status Notifications Into The Handler

**Files:**
- Modify: `server/internal/handler/order_handler.go`
- Modify: `server/cmd/server/main.go`
- Test: `server/internal/service/order_notification_service_test.go`

- [ ] **Step 1: Write the failing integration-shape test**

```go
func TestShouldNotifyBranchStatusRejectsCancelled(t *testing.T) {
	if shouldNotifyBranchStatus("cancelled") {
		t.Fatalf("cancelled should not be treated as admin status notification")
	}
}
```

- [ ] **Step 2: Run the focused test**

Run: `cd server && go test ./internal/service -run TestShouldNotifyBranchStatusRejectsCancelled`

Expected: PASS if helpers are present; otherwise complete Task 4 first before continuing.

- [ ] **Step 3: Inject the notification service into the order handler**

Update constructor and struct in `server/internal/handler/order_handler.go`:

```go
type OrderHandler struct {
	orderService         *service.OrderService
	orderNotificationSvc *service.OrderNotificationService
	branchRepo           *repository.BranchRepo
	telegramBot          *telegram.Bot
	hub                  *ws.Hub
}

func NewOrderHandler(orderService *service.OrderService, orderNotificationSvc *service.OrderNotificationService, branchRepo *repository.BranchRepo, telegramBot *telegram.Bot, hub *ws.Hub) *OrderHandler {
	return &OrderHandler{
		orderService:         orderService,
		orderNotificationSvc: orderNotificationSvc,
		branchRepo:           branchRepo,
		telegramBot:          telegramBot,
		hub:                  hub,
	}
}
```

Update `AdminUpdateStatus` to trigger after successful persistence:

```go
if h.orderNotificationSvc != nil {
	h.orderNotificationSvc.NotifyBranchOrderStatus(c.Request().Context(), order)
}
```

- [ ] **Step 4: Wire the service in `server/cmd/server/main.go`**

```go
deliveryRepo := repository.NewNotificationDeliveryRepo(db)
orderNotificationService := service.NewOrderNotificationService(branchRepo, orderRepo, deliveryRepo, bot)

handlers := &handler.Handlers{
	Order: handler.NewOrderHandler(orderService, orderNotificationService, branchRepo, bot, hub),
}
```

- [ ] **Step 5: Run focused server tests**

Run: `cd server && go test ./internal/service ./internal/telegram`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/internal/handler/order_handler.go server/cmd/server/main.go server/internal/service/order_notification_service_test.go
git commit -m "feat(order): notify branch chats on admin status changes"
```

### Task 6: Add The Daily Summary Scheduler

**Files:**
- Create: `server/internal/service/order_notification_scheduler.go`
- Modify: `server/cmd/server/main.go`
- Modify: `server/internal/service/order_notification_service_test.go`

- [ ] **Step 1: Write the failing scheduler helper test**

```go
func TestShouldSendDailySummaryAtFivePastMidnight(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Tashkent")
	now := time.Date(2026, 4, 2, 0, 5, 0, 0, loc)

	if !shouldRunDailySummary(now) {
		t.Fatalf("expected summary run at 00:05 Asia/Tashkent")
	}
}
```

- [ ] **Step 2: Run the focused scheduler test**

Run: `cd server && go test ./internal/service -run TestShouldSendDailySummaryAtFivePastMidnight`

Expected: FAIL with missing scheduler helper.

- [ ] **Step 3: Implement the scheduler**

Create `server/internal/service/order_notification_scheduler.go`:

```go
package service

import (
	"context"
	"log"
	"time"
)

func shouldRunDailySummary(now time.Time) bool {
	return now.Hour() == 0 && now.Minute() == 5
}

func (s *OrderNotificationService) StartDailySummaryScheduler(ctx context.Context) {
	if s == nil || s.bot == nil {
		return
	}

	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now().In(s.location)
				if !shouldRunDailySummary(now) {
					continue
				}
				localDate := now.AddDate(0, 0, -1).Format("2006-01-02")
				s.runDailySummaries(ctx, localDate)
			}
		}
	}()
}
```

Add `runDailySummaries` to the service:

```go
func (s *OrderNotificationService) runDailySummaries(ctx context.Context, localDate string) {
	branches, err := s.branchRepo.ListActiveWithTelegramChat(ctx)
	if err != nil {
		log.Printf("notifications: failed to load branches for daily summaries: %v", err)
		return
	}
	for _, branch := range branches {
		if err := s.SendDailySummaryForBranch(ctx, branch, localDate); err != nil {
			log.Printf("notifications: failed to send daily summary for branch %s date %s: %v", branch.ID, localDate, err)
		}
	}
}
```

- [ ] **Step 4: Start the scheduler from `server/cmd/server/main.go`**

```go
appCtx, cancel := context.WithCancel(context.Background())
defer cancel()

if orderNotificationService != nil {
	orderNotificationService.StartDailySummaryScheduler(appCtx)
}
```

Keep this near other startup wiring, after the bot/service creation and before `e.Start(...)`.

- [ ] **Step 5: Re-run the focused service tests**

Run: `cd server && go test ./internal/service -run 'Test(ShouldSendDailySummaryAtFivePastMidnight|SummaryWindowForDateUsesTashkentMidnight)'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/internal/service/order_notification_scheduler.go server/internal/service/order_notification_service.go server/cmd/server/main.go server/internal/service/order_notification_service_test.go
git commit -m "feat(notifications): schedule branch daily summaries"
```

### Task 7: Full Verification And Docs Triage

**Files:**
- Modify: `docs/superpowers/plans/2026-04-01-telegram-order-notifications.md`
- Verify: `docs/superpowers/specs/2026-04-01-telegram-order-notifications-design.md`

- [ ] **Step 1: Run focused backend tests**

Run: `cd server && go test ./internal/service ./internal/telegram`

Expected: PASS

- [ ] **Step 2: Run full server tests**

Run: `cd server && go test ./...`

Expected: PASS

- [ ] **Step 3: Run repo quality flow**

Run: `make quality-server`

Expected: PASS with formatting, lint, and test checks green for the Go server.

- [ ] **Step 4: Run docs triage**

Run: `make docs-check`

Expected: PASS or a focused advisory report. If the report flags `AGENTS.md`, `README.md`, or notification-related docs as stale, update them in a follow-up docs commit.

- [ ] **Step 5: Review the final diff**

Run: `git status --short`

Expected: only intended notification, migration, test, and plan/doc files are modified.

- [ ] **Step 6: Commit the completed implementation**

```bash
git add server/cmd/server/main.go server/internal/handler/order_handler.go server/internal/model/notification.go server/internal/repository/order_repo.go server/internal/repository/branch_repo.go server/internal/repository/notification_delivery_repo.go server/internal/service/order_notification_service.go server/internal/service/order_notification_scheduler.go server/internal/service/order_notification_service_test.go server/internal/telegram/notifications.go server/internal/telegram/notifications_test.go server/migrations/000004_daily_notification_deliveries.up.sql server/migrations/000004_daily_notification_deliveries.down.sql
git commit -m "feat(telegram): add branch order status and daily summary notifications"
```

## Self-Review

- Spec coverage:
  - admin-driven status notifications are covered in Tasks 3, 4, and 5
  - branch-specific daily summaries are covered in Tasks 1, 2, 4, and 6
  - `Asia/Tashkent` scheduler handling is covered in Tasks 4 and 6
  - duplicate prevention is covered in Tasks 1, 4, and 6
  - verification is covered in Task 7
- Placeholder scan:
  - no `TBD`, `TODO`, or deferred implementation placeholders remain
- Type consistency:
  - the plan uses `NotificationDelivery`, `BranchDailyOrderSummary`, `OrderNotificationService`, `shouldNotifyBranchStatus`, and `summaryWindowForDate` consistently across tasks
