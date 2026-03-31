package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
)

func TestNotificationDeliveryGroundwork(t *testing.T) {
	if repository.NewNotificationDeliveryRepo(nil) == nil {
		t.Fatal("expected notification delivery repo constructor to return a repo")
	}

	typ := reflect.TypeOf(model.NotificationDelivery{})
	for _, name := range []string{"ID", "NotificationType", "BranchID", "LocalDate", "SentAt", "CreatedAt"} {
		if _, ok := typ.FieldByName(name); !ok {
			t.Fatalf("expected NotificationDelivery to include field %s", name)
		}
	}
}

func TestShouldNotifyBranchStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"accepted":  true,
		"preparing": true,
		"ready":     true,
		"picked_up": true,
		"rejected":  true,
		"pending":   false,
		"cancelled": false,
		"completed": false,
		"something": false,
		"":          false,
	}

	for status, want := range cases {
		if got := shouldNotifyBranchStatus(status); got != want {
			t.Fatalf("shouldNotifyBranchStatus(%q) = %v, want %v", status, got, want)
		}
	}
}

func TestSummaryWindowForDateUsesTashkentMidnight(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	input := time.Date(2026, time.April, 1, 20, 30, 0, 0, time.UTC)
	localMidnight, startUTC, endUTC := summaryWindowForDate(input, loc)

	wantLocal := time.Date(2026, time.April, 2, 0, 0, 0, 0, loc)
	if !localMidnight.Equal(wantLocal) {
		t.Fatalf("localMidnight = %s, want %s", localMidnight, wantLocal)
	}

	if !startUTC.Equal(wantLocal.UTC()) {
		t.Fatalf("startUTC = %s, want %s", startUTC, wantLocal.UTC())
	}

	wantEndUTC := wantLocal.AddDate(0, 0, 1).UTC()
	if !endUTC.Equal(wantEndUTC) {
		t.Fatalf("endUTC = %s, want %s", endUTC, wantEndUTC)
	}
}

func TestShouldRunDailySummary(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	cases := map[string]bool{
		"2026-04-02T00:05:00+05:00": true,
		"2026-04-02T00:05:59+05:00": true,
		"2026-04-02T00:04:59+05:00": false,
		"2026-04-02T00:06:00+05:00": false,
		"2026-04-02T01:05:00+05:00": false,
	}

	for value, want := range cases {
		now, err := time.Parse(time.RFC3339, value)
		if err != nil {
			t.Fatalf("parse time %q: %v", value, err)
		}
		if got := shouldRunDailySummary(now.In(loc)); got != want {
			t.Fatalf("shouldRunDailySummary(%s) = %v, want %v", value, got, want)
		}
	}
}

func TestDailySummaryDateForRunUsesYesterday(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	now := time.Date(2026, time.April, 2, 0, 5, 0, 0, loc)
	want := time.Date(2026, time.April, 1, 0, 5, 0, 0, loc)

	got := dailySummaryDateForRun(now, loc)
	if !got.Equal(want) {
		t.Fatalf("dailySummaryDateForRun(%s) = %s, want %s", now, got, want)
	}
}

func TestNotifyBranchOrderStatusSendsAllowedStatuses(t *testing.T) {
	t.Parallel()

	branchID := "branch-1"
	bot := &fakeOrderNotificationBot{}
	svc := NewOrderNotificationService(
		&fakeBranchRepo{
			detail: &repository.BranchDetail{
				Branch: model.Branch{
					ID:                  branchID,
					Name:                "Chilonzor",
					TelegramGroupChatID: int64Ptr(777),
				},
			},
		},
		nil,
		nil,
		bot,
	)

	allowed := []string{"accepted", "preparing", "ready", "picked_up", "rejected"}
	for _, status := range allowed {
		bot.reset()
		order := &model.Order{OrderNumber: 42, BranchID: branchID, Status: status, TotalPrice: 50000}
		if err := svc.NotifyBranchOrderStatus(context.Background(), order); err != nil {
			t.Fatalf("NotifyBranchOrderStatus(%s) returned error: %v", status, err)
		}
		if len(bot.statusSends) != 1 {
			t.Fatalf("expected 1 status send for %s, got %d", status, len(bot.statusSends))
		}
		if bot.statusSends[0].branchName != "Chilonzor" {
			t.Fatalf("expected branch name to be forwarded, got %q", bot.statusSends[0].branchName)
		}
		if bot.statusSends[0].groupChatID != 777 {
			t.Fatalf("expected branch telegram chat to be forwarded, got %d", bot.statusSends[0].groupChatID)
		}
	}

	bot.reset()
	if err := svc.NotifyBranchOrderStatus(context.Background(), &model.Order{OrderNumber: 43, BranchID: branchID, Status: "cancelled"}); err != nil {
		t.Fatalf("NotifyBranchOrderStatus(cancelled) returned error: %v", err)
	}
	if len(bot.statusSends) != 0 {
		t.Fatalf("expected no send for cancelled status, got %d", len(bot.statusSends))
	}
}

func TestSendDailySummariesForDateOrchestratesBranches(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	summaryDate := time.Date(2026, time.April, 2, 10, 0, 0, 0, time.UTC)
	bot := &fakeOrderNotificationBot{}
	deliveryRepo := &fakeNotificationDeliveryRepo{}
	orderRepo := &fakeOrderNotificationOrderRepo{
		summary: &model.BranchDailyOrderSummary{
			TotalOrders:           10,
			PendingOrders:         1,
			AcceptedOrders:        2,
			PreparingOrders:       3,
			ReadyOrders:           1,
			PickedUpOrders:        2,
			RejectedOrders:        1,
			CancelledOrders:       0,
			TotalCreatedOrderSum:  120000,
			TotalPickedUpOrderSum: 70000,
		},
	}

	svc := &OrderNotificationService{
		branchRepo: &fakeBranchRepo{
			branches: []model.Branch{
				{ID: "branch-1", Name: "Chilonzor", TelegramGroupChatID: int64Ptr(777)},
				{ID: "branch-2", Name: "Yakkasaroy"},
			},
		},
		orderRepo:    orderRepo,
		deliveryRepo: deliveryRepo,
		bot:          bot,
		location:     loc,
	}

	if err := svc.SendDailySummariesForDate(context.Background(), summaryDate); err != nil {
		t.Fatalf("SendDailySummariesForDate returned error: %v", err)
	}

	if len(bot.summarySends) != 1 {
		t.Fatalf("expected 1 summary send, got %d", len(bot.summarySends))
	}
	if bot.summarySends[0].groupChatID != 777 {
		t.Fatalf("expected summary to target branch telegram chat, got %d", bot.summarySends[0].groupChatID)
	}
	if len(orderRepo.calls) != 1 {
		t.Fatalf("expected 1 summary repo call, got %d", len(orderRepo.calls))
	}

	wantLocalDate, _, _ := summaryWindowForDate(summaryDate, loc)
	if len(deliveryRepo.created) != 1 {
		t.Fatalf("expected 1 delivery record, got %d", len(deliveryRepo.created))
	}
	if !deliveryRepo.created[0].LocalDate.Equal(wantLocalDate) {
		t.Fatalf("expected delivery local date %s, got %s", wantLocalDate, deliveryRepo.created[0].LocalDate)
	}
	if deliveryRepo.created[0].BranchID != "branch-1" {
		t.Fatalf("expected delivery to be stored for branch-1, got %s", deliveryRepo.created[0].BranchID)
	}
	if got := orderRepo.calls[0].summaryAt; !got.Equal(wantLocalDate) {
		t.Fatalf("expected summary repo to receive %s, got %s", wantLocalDate, got)
	}
}

type fakeBranchRepo struct {
	detail   *repository.BranchDetail
	branches []model.Branch
}

func (f *fakeBranchRepo) GetByID(context.Context, string) (*repository.BranchDetail, error) {
	return f.detail, nil
}

func (f *fakeBranchRepo) ListActiveWithTelegramChatIDs(context.Context) ([]model.Branch, error) {
	return f.branches, nil
}

type fakeNotificationDeliveryRepo struct {
	exists  bool
	created []model.NotificationDelivery
}

func (f *fakeNotificationDeliveryRepo) Exists(context.Context, string, string, time.Time) (bool, error) {
	return f.exists, nil
}

func (f *fakeNotificationDeliveryRepo) Create(_ context.Context, delivery *model.NotificationDelivery) error {
	f.created = append(f.created, *delivery)
	return nil
}

type fakeOrderNotificationOrderRepo struct {
	summary *model.BranchDailyOrderSummary
	calls   []summaryCall
}

type summaryCall struct {
	branchID  string
	summaryAt time.Time
	startUTC  time.Time
	endUTC    time.Time
}

func (f *fakeOrderNotificationOrderRepo) GetBranchDailySummary(_ context.Context, branchID string, summaryDate time.Time, startUTC, endUTC time.Time) (*model.BranchDailyOrderSummary, error) {
	f.calls = append(f.calls, summaryCall{
		branchID:  branchID,
		summaryAt: summaryDate,
		startUTC:  startUTC,
		endUTC:    endUTC,
	})
	if f.summary == nil {
		return &model.BranchDailyOrderSummary{}, nil
	}
	summary := *f.summary
	return &summary, nil
}

type fakeOrderNotificationStatusSend struct {
	groupChatID int64
	branchName  string
	order       *model.Order
}

type fakeOrderNotificationBot struct {
	statusSends  []fakeOrderNotificationStatusSend
	summarySends []fakeOrderNotificationSummarySend
}

type fakeOrderNotificationSummarySend struct {
	groupChatID int64
	summary     *model.BranchDailyOrderSummary
}

func (f *fakeOrderNotificationBot) reset() {
	f.statusSends = nil
	f.summarySends = nil
}

func (f *fakeOrderNotificationBot) SendOrderStatusToChat(groupChatID int64, branchName string, order *model.Order) {
	f.statusSends = append(f.statusSends, fakeOrderNotificationStatusSend{
		groupChatID: groupChatID,
		branchName:  branchName,
		order:       order,
	})
}

func (f *fakeOrderNotificationBot) SendBranchDailySummary(groupChatID int64, summary *model.BranchDailyOrderSummary) {
	f.summarySends = append(f.summarySends, fakeOrderNotificationSummarySend{
		groupChatID: groupChatID,
		summary:     summary,
	})
}

func int64Ptr(v int64) *int64 {
	return &v
}
