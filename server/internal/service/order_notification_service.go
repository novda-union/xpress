package service

import (
	"context"
	"log"
	"time"

	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
)

const branchDailySummaryNotificationType = "branch_daily_summary"

type orderNotificationBranchRepo interface {
	GetByID(ctx context.Context, id string) (*repository.BranchDetail, error)
	ListActiveWithTelegramChatIDs(ctx context.Context) ([]model.Branch, error)
}

type orderNotificationOrderRepo interface {
	GetBranchDailySummary(ctx context.Context, branchID string, summaryDate time.Time, startUTC, endUTC time.Time) (*model.BranchDailyOrderSummary, error)
}

type orderNotificationDeliveryRepo interface {
	Exists(ctx context.Context, notificationType, branchID string, localDate time.Time) (bool, error)
	Create(ctx context.Context, delivery *model.NotificationDelivery) error
}

type orderNotificationBot interface {
	SendOrderStatusToChat(groupChatID int64, branchName string, order *model.Order)
	SendBranchDailySummary(groupChatID int64, summary *model.BranchDailyOrderSummary)
}

type OrderNotificationService struct {
	branchRepo   orderNotificationBranchRepo
	orderRepo    orderNotificationOrderRepo
	deliveryRepo orderNotificationDeliveryRepo
	bot          orderNotificationBot
	location     *time.Location
}

func NewOrderNotificationService(
	branchRepo orderNotificationBranchRepo,
	orderRepo orderNotificationOrderRepo,
	deliveryRepo orderNotificationDeliveryRepo,
	bot orderNotificationBot,
) *OrderNotificationService {
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

func summaryWindowForDate(summaryDate time.Time, loc *time.Location) (time.Time, time.Time, time.Time) {
	if loc == nil {
		loc = time.FixedZone("Asia/Tashkent", 5*60*60)
	}

	localTime := summaryDate.In(loc)
	year, month, day := localTime.Date()
	localMidnight := time.Date(year, month, day, 0, 0, 0, 0, loc)
	startUTC := localMidnight.UTC()
	endUTC := localMidnight.AddDate(0, 0, 1).UTC()
	return localMidnight, startUTC, endUTC
}

func (s *OrderNotificationService) NotifyBranchOrderStatus(ctx context.Context, order *model.Order) error {
	if s == nil || order == nil || s.bot == nil || s.branchRepo == nil {
		return nil
	}
	if !shouldNotifyBranchStatus(order.Status) {
		return nil
	}

	detail, err := s.branchRepo.GetByID(ctx, order.BranchID)
	if err != nil {
		return err
	}
	if detail == nil || detail.Branch.TelegramGroupChatID == nil {
		log.Printf("notifications: skipping order #%d status %s for branch %s because no telegram group chat is configured", order.OrderNumber, order.Status, order.BranchID)
		return nil
	}

	s.bot.SendOrderStatusToChat(*detail.Branch.TelegramGroupChatID, detail.Branch.Name, order)
	return nil
}

func (s *OrderNotificationService) SendDailySummariesForDate(ctx context.Context, summaryDate time.Time) error {
	if s == nil || s.branchRepo == nil || s.orderRepo == nil || s.deliveryRepo == nil || s.bot == nil {
		return nil
	}

	branches, err := s.branchRepo.ListActiveWithTelegramChatIDs(ctx)
	if err != nil {
		return err
	}

	for _, branch := range branches {
		if err := s.SendDailySummaryForBranch(ctx, branch, summaryDate); err != nil {
			log.Printf("notifications: failed daily summary for branch %s (%s): %v", branch.ID, branch.Name, err)
		}
	}

	return nil
}

func (s *OrderNotificationService) SendDailySummaryForBranch(ctx context.Context, branch model.Branch, summaryDate time.Time) error {
	if s == nil || s.orderRepo == nil || s.deliveryRepo == nil || s.bot == nil {
		return nil
	}
	if branch.TelegramGroupChatID == nil {
		log.Printf("notifications: skipping daily summary for branch %s because no telegram group chat is configured", branch.ID)
		return nil
	}

	localDate, startUTC, endUTC := summaryWindowForDate(summaryDate, s.location)

	exists, err := s.deliveryRepo.Exists(ctx, branchDailySummaryNotificationType, branch.ID, localDate)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	summary, err := s.orderRepo.GetBranchDailySummary(ctx, branch.ID, localDate, startUTC, endUTC)
	if err != nil {
		return err
	}
	if summary == nil {
		summary = &model.BranchDailyOrderSummary{}
	}

	summary.BranchID = branch.ID
	summary.BranchName = branch.Name
	summary.LocalDate = localDate

	s.bot.SendBranchDailySummary(*branch.TelegramGroupChatID, summary)

	delivery := &model.NotificationDelivery{
		NotificationType: branchDailySummaryNotificationType,
		BranchID:         branch.ID,
		LocalDate:        localDate,
	}
	if err := s.deliveryRepo.Create(ctx, delivery); err != nil {
		return err
	}

	return nil
}
