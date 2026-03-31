package service

import (
	"context"
	"log"
	"time"
)

const dailySummaryRunMinute = 5

func shouldRunDailySummary(now time.Time) bool {
	return now.Hour() == 0 && now.Minute() == dailySummaryRunMinute
}

func dailySummaryDateForRun(now time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.FixedZone("Asia/Tashkent", 5*60*60)
	}

	return now.In(loc).AddDate(0, 0, -1)
}

func (s *OrderNotificationService) StartDailySummaryScheduler(ctx context.Context) {
	if s == nil || s.bot == nil || s.branchRepo == nil || s.orderRepo == nil || s.deliveryRepo == nil {
		return
	}

	location := s.location
	if location == nil {
		location = time.FixedZone("Asia/Tashkent", 5*60*60)
	}

	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				localNow := now.In(location)
				if !shouldRunDailySummary(localNow) {
					continue
				}

				summaryDate := dailySummaryDateForRun(localNow, location)
				if err := s.SendDailySummariesForDate(ctx, summaryDate); err != nil {
					log.Printf("notifications: failed to run daily summary scheduler for %s: %v", summaryDate.Format("2006-01-02"), err)
				}
			}
		}
	}()
}
