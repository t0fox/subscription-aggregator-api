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
	GetAll(ctx context.Context, limit, offset int) ([]models.Subscription, error)
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

func (s *SubscriptionService) Create(ctx context.Context, subReq *models.SubscriptionCreateRequest) (*models.Subscription, error) {
	if _, err := uuid.Parse(subReq.UserID); err != nil {
		return nil, fmt.Errorf("%w: invalid UUID format", models.ErrValidation)
	}

	startDate, err := time.Parse("01-2006", subReq.StartDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid start date format, expected MM-YYYY", models.ErrValidation)
	}

	var endDate *time.Time
	if subReq.EndDate != nil {
		parsedEnd, err := time.Parse("01-2006", *subReq.EndDate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid end date format, expected MM-YYYY", models.ErrValidation)
		}
		if parsedEnd.Before(startDate) {
			return nil, fmt.Errorf("%w: invalid end date: must be greater than or equal to start date", models.ErrValidation)
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

	return sub, s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: invalid UUID format", models.ErrValidation)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) GetAll(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *SubscriptionService) Update(ctx context.Context, id string, updateReq *models.SubscriptionUpdateRequest) (*models.Subscription, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: invalid UUID format", models.ErrValidation)
	}
	if updateReq.Price != nil && *updateReq.Price < 1 {
		return nil, fmt.Errorf("%w: invalid price: must be greater than zero", models.ErrValidation)
	}

	if updateReq.EndDate != nil {
		parsedEnd, err := time.Parse("01-2006", *updateReq.EndDate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid end date format, expected MM-YYYY", models.ErrValidation)
		}

		current, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if parsedEnd.Before(current.StartDate) {
			return nil, fmt.Errorf("%w: invalid end date: must be greater than or equal to start date", models.ErrValidation)
		}
	}

	return s.repo.Update(ctx, id, updateReq)
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: invalid UUID format", models.ErrValidation)
	}

	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) GetSumByPeriod(ctx context.Context, filter *models.SubscriptionFilter) (int, error) {
	periodStart, periodEnd, err := validateSumFilter(filter)
	if err != nil {
		return 0, err
	}

	subscriptions, err := s.repo.GetByPeriod(ctx, filter)
	if err != nil {
		return 0, err
	}

	return calculateSubscriptionsTotal(subscriptions, periodStart, periodEnd), nil
}

func validateSumFilter(filter *models.SubscriptionFilter) (time.Time, time.Time, error) {
	if filter.UserID != nil {
		if _, err := uuid.Parse(*filter.UserID); err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid UUID format", models.ErrValidation)
		}
	}

	if filter.StartDate == nil || filter.EndDate == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid period: start_date and end_date are required", models.ErrValidation)
	}

	startDate, err := time.Parse("01-2006", *filter.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid start date format, expected MM-YYYY", models.ErrValidation)
	}

	endDate, err := time.Parse("01-2006", *filter.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid end date format, expected MM-YYYY", models.ErrValidation)
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid period: end date must be greater than or equal to start date", models.ErrValidation)
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
