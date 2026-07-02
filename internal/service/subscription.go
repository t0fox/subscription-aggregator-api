package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/subscription_service/internal/models"
	"github.com/subscription_service/internal/repository"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(subReq *models.SubscriptionCreateRequest) (*models.Subscription, error) {
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

	return sub, s.repo.Create(nil, sub)
}

func (s *SubscriptionService) GetByID(id string) (*models.Subscription, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid UUID format")
	}

	return s.repo.GetByID(nil, id)
}

func (s *SubscriptionService) GetAll() ([]models.Subscription, error) {
	return s.repo.GetAll(nil)
}

func (s *SubscriptionService) Update(id string, updateReq *models.SubscriptionUpdateRequest) (*models.Subscription, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid UUID format")
	}

	return s.repo.Update(nil, id, updateReq)
}

func (s *SubscriptionService) Delete(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid UUID format")
	}

	return s.repo.Delete(nil, id)
}

func (s *SubscriptionService) GetSumByPeriod(filter *models.SubscriptionFilter) (int, error) {
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

	return s.repo.GetSumByPeriod(nil, filter)
}
