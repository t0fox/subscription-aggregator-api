package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/t0fox/subscription-aggregator-api/internal/models"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id string) (*models.Subscription, error)
	GetAll(ctx context.Context) ([]models.Subscription, error)
	Update(ctx context.Context, id string, update *models.SubscriptionUpdateRequest) (*models.Subscription, error)
	Delete(ctx context.Context, id string) error
	GetByPeriod(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(subReq *models.SubscriptionCreateRequest) (*models.Subscription, error) {
	if _, err := uuid.Parse(subReq.UserID); err != nil {
		return nil, fmt.Errorf("invalid UUID format")
	}

	startDate, err := time.Parse("01-2006", subReq.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format, expected MM-YYYY")
	}

	var endDate *time.Time
	if subReq.EndDate != nil {
		parsedEnd, err := time.Parse("01-2006", *subReq.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format, expected MM-YYYY")
		}
		if parsedEnd.Before(startDate) {
			return nil, fmt.Errorf("invalid end date: must be greater than or equal to start date")
		}
		endDate = &parsedEnd
	}

	sub := &models.Subscription{
		ServiceName: subReq.ServiceName,
		Price:       subReq.Price,
		UserID:      subReq.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	return sub, s.repo.Create(context.Background(), sub)
}

func (s *SubscriptionService) GetByID(id string) (*models.Subscription, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid UUID format")
	}

	return s.repo.GetByID(context.Background(), id)
}

func (s *SubscriptionService) GetAll() ([]models.Subscription, error) {
	return s.repo.GetAll(context.Background())
}

func (s *SubscriptionService) Update(id string, updateReq *models.SubscriptionUpdateRequest) (*models.Subscription, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid UUID format")
	}
	if updateReq.Price != nil && *updateReq.Price < 1 {
		return nil, fmt.Errorf("invalid price: must be greater than zero")
	}

	if updateReq.EndDate != nil {
		parsedEnd, err := time.Parse("01-2006", *updateReq.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format, expected MM-YYYY")
		}

		current, err := s.repo.GetByID(context.Background(), id)
		if err != nil {
			return nil, err
		}
		if parsedEnd.Before(current.StartDate) {
			return nil, fmt.Errorf("invalid end date: must be greater than or equal to start date")
		}
	}

	return s.repo.Update(context.Background(), id, updateReq)
}

func (s *SubscriptionService) Delete(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid UUID format")
	}

	return s.repo.Delete(context.Background(), id)
}

func (s *SubscriptionService) GetSumByPeriod(filter *models.SubscriptionFilter) (int, error) {
	periodStart, periodEnd, err := validateSumFilter(filter)
	if err != nil {
		return 0, err
	}

	subscriptions, err := s.repo.GetByPeriod(context.Background(), filter)
	if err != nil {
		return 0, err
	}

	return calculateSubscriptionsTotal(subscriptions, periodStart, periodEnd), nil
}

func validateSumFilter(filter *models.SubscriptionFilter) (time.Time, time.Time, error) {
	if filter.UserID != nil {
		if _, err := uuid.Parse(*filter.UserID); err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid UUID format")
		}
	}

	if filter.StartDate == nil || filter.EndDate == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period: start_date and end_date are required")
	}

	startDate, err := time.Parse("01-2006", *filter.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format, expected MM-YYYY")
	}

	endDate, err := time.Parse("01-2006", *filter.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format, expected MM-YYYY")
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period: end date must be greater than or equal to start date")
	}

	return startDate, endDate, nil
}

func calculateSubscriptionsTotal(subscriptions []models.Subscription, periodStart, periodEnd time.Time) int {
	total := 0
	for _, sub := range subscriptions {
		overlapStart := maxMonth(sub.StartDate, periodStart)
		overlapEnd := periodEnd
		if sub.EndDate != nil {
			overlapEnd = minMonth(*sub.EndDate, periodEnd)
		}

		months := monthsBetweenInclusive(overlapStart, overlapEnd)
		if months > 0 {
			total += sub.Price * months
		}
	}

	return total
}

func maxMonth(a, b time.Time) time.Time {
	a = monthStart(a)
	b = monthStart(b)
	if a.After(b) {
		return a
	}
	return b
}

func minMonth(a, b time.Time) time.Time {
	a = monthStart(a)
	b = monthStart(b)
	if a.Before(b) {
		return a
	}
	return b
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func monthsBetweenInclusive(start, end time.Time) int {
	start = monthStart(start)
	end = monthStart(end)
	if end.Before(start) {
		return 0
	}

	return (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1
}
