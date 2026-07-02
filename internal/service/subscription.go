package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/t0fox/subscription-aggregator-api/internal/models"
	"github.com/t0fox/subscription-aggregator-api/internal/repository"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
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

	return s.repo.Update(context.Background(), id, updateReq)
}

func (s *SubscriptionService) Delete(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid UUID format")
	}

	return s.repo.Delete(context.Background(), id)
}

func (s *SubscriptionService) GetSumByPeriod(filter *models.SubscriptionFilter) (int, error) {
	if filter.UserID != nil {
		if _, err := uuid.Parse(*filter.UserID); err != nil {
			return 0, fmt.Errorf("invalid UUID format")
		}
	}

	if filter.StartDate != nil {
		_, err := time.Parse("01-2006", *filter.StartDate)
		if err != nil {
			return 0, fmt.Errorf("invalid start date format, expected MM-YYYY")
		}
	}

	if filter.EndDate != nil {
		_, err := time.Parse("01-2006", *filter.EndDate)
		if err != nil {
			return 0, fmt.Errorf("invalid end date format, expected MM-YYYY")
		}
	}

	return s.repo.GetSumByPeriod(context.Background(), filter)
}
