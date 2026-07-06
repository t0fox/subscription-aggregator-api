package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/t0fox/subscription-aggregator-api/internal/models"
)

type fakeSubscriptionRepository struct {
	subscriptions   []models.Subscription
	getByIDResult   *models.Subscription
}

func (f *fakeSubscriptionRepository) Create(context.Context, *models.Subscription) error {
	return nil
}

func (f *fakeSubscriptionRepository) GetByID(context.Context, string) (*models.Subscription, error) {
	if f.getByIDResult != nil {
		return f.getByIDResult, nil
	}
	return &models.Subscription{}, nil
}

func (f *fakeSubscriptionRepository) GetAll(context.Context, int, int) ([]models.Subscription, error) {
	return nil, nil
}

func (f *fakeSubscriptionRepository) Update(context.Context, string, *models.SubscriptionUpdateRequest) (*models.Subscription, error) {
	return &models.Subscription{}, nil
}

func (f *fakeSubscriptionRepository) Delete(context.Context, string) error {
	return nil
}

func (f *fakeSubscriptionRepository) GetByPeriod(context.Context, *models.SubscriptionFilter) ([]models.Subscription, error) {
	return f.subscriptions, nil
}

func TestSubscriptionServiceGetSumByPeriod(t *testing.T) {
	userID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"
	serviceName := "Yandex Plus"

	tests := []struct {
		name          string
		subscriptions []models.Subscription
		filter        models.SubscriptionFilter
		want          int
		wantErr       string
	}{
		{
			name: "single month",
			subscriptions: []models.Subscription{
				newSubscription(400, "07-2025", stringPtr("07-2025")),
			},
			filter: models.SubscriptionFilter{
				UserID:      &userID,
				ServiceName: &serviceName,
				StartDate:   stringPtr("07-2025"),
				EndDate:     stringPtr("07-2025"),
			},
			want: 400,
		},
		{
			name: "multiple months",
			subscriptions: []models.Subscription{
				newSubscription(400, "07-2025", stringPtr("12-2025")),
			},
			filter: models.SubscriptionFilter{
				UserID:      &userID,
				ServiceName: &serviceName,
				StartDate:   stringPtr("07-2025"),
				EndDate:     stringPtr("12-2025"),
			},
			want: 2400,
		},
		{
			name:          "empty result",
			subscriptions: []models.Subscription{},
			filter: models.SubscriptionFilter{
				UserID:      &userID,
				ServiceName: &serviceName,
				StartDate:   stringPtr("07-2025"),
				EndDate:     stringPtr("12-2025"),
			},
			want: 0,
		},
		{
			name: "invalid date range",
			filter: models.SubscriptionFilter{
				StartDate: stringPtr("12-2025"),
				EndDate:   stringPtr("07-2025"),
			},
			wantErr: "invalid period",
		},
		{
			name: "partial overlap",
			subscriptions: []models.Subscription{
				newSubscription(100, "06-2025", stringPtr("08-2025")),
			},
			filter: models.SubscriptionFilter{
				StartDate: stringPtr("07-2025"),
				EndDate:   stringPtr("09-2025"),
			},
			want: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSubscriptionService(&fakeSubscriptionRepository{subscriptions: tt.subscriptions})
			got, err := svc.GetSumByPeriod(context.Background(), &tt.filter)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("total = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSubscriptionServiceCreate(t *testing.T) {
	userID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"

	tests := []struct {
		name    string
		req     models.SubscriptionCreateRequest
		wantErr string
	}{
		{
			name: "valid without end date",
			req: models.SubscriptionCreateRequest{
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      userID,
				StartDate:   "07-2025",
			},
		},
		{
			name: "valid with end date",
			req: models.SubscriptionCreateRequest{
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      userID,
				StartDate:   "07-2025",
				EndDate:     stringPtr("12-2025"),
			},
		},
		{
			name: "invalid uuid",
			req: models.SubscriptionCreateRequest{
				UserID:    "not-a-uuid",
				StartDate: "07-2025",
			},
			wantErr: "invalid UUID",
		},
		{
			name: "invalid start date",
			req: models.SubscriptionCreateRequest{
				UserID:    userID,
				StartDate: "2025-07",
			},
			wantErr: "invalid start date",
		},
		{
			name: "invalid end date format",
			req: models.SubscriptionCreateRequest{
				UserID:    userID,
				StartDate: "07-2025",
				EndDate:   stringPtr("2025-12"),
			},
			wantErr: "invalid end date format",
		},
		{
			name: "end before start",
			req: models.SubscriptionCreateRequest{
				UserID:    userID,
				StartDate: "12-2025",
				EndDate:   stringPtr("07-2025"),
			},
			wantErr: "must be greater than or equal to start date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSubscriptionService(&fakeSubscriptionRepository{})
			got, err := svc.Create(context.Background(), &tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected subscription, got nil")
			}
		})
	}
}

func TestSubscriptionServiceUpdate(t *testing.T) {
	userID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"

	tests := []struct {
		name    string
		id      string
		req     models.SubscriptionUpdateRequest
		repo    *fakeSubscriptionRepository
		wantErr string
	}{
		{
			name: "valid price update",
			id:   userID,
			req:  models.SubscriptionUpdateRequest{Price: intPtr(500)},
			repo: &fakeSubscriptionRepository{},
		},
		{
			name:    "invalid uuid",
			id:      "not-a-uuid",
			req:     models.SubscriptionUpdateRequest{},
			repo:    &fakeSubscriptionRepository{},
			wantErr: "invalid UUID",
		},
		{
			name:    "price below one",
			id:      userID,
			req:     models.SubscriptionUpdateRequest{Price: intPtr(0)},
			repo:    &fakeSubscriptionRepository{},
			wantErr: "invalid price",
		},
		{
			name:    "invalid end date format",
			id:      userID,
			req:     models.SubscriptionUpdateRequest{EndDate: stringPtr("2025-12")},
			repo:    &fakeSubscriptionRepository{},
			wantErr: "invalid end date format",
		},
		{
			name: "end before start",
			id:   userID,
			req:  models.SubscriptionUpdateRequest{EndDate: stringPtr("07-2025")},
			repo: &fakeSubscriptionRepository{
				getByIDResult: &models.Subscription{StartDate: mustParseMonth("12-2025")},
			},
			wantErr: "must be greater than or equal to start date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSubscriptionService(tt.repo)
			_, err := svc.Update(context.Background(), tt.id, &tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSubscriptionServiceGetByIDValidation(t *testing.T) {
	svc := NewSubscriptionService(&fakeSubscriptionRepository{})

	if _, err := svc.GetByID(context.Background(), "not-a-uuid"); err == nil {
		t.Fatal("expected error for invalid uuid")
	}
	if _, err := svc.GetByID(context.Background(), "60601fee-2bf1-4721-ae6f-7636e79a0cba"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscriptionServiceDeleteValidation(t *testing.T) {
	svc := NewSubscriptionService(&fakeSubscriptionRepository{})

	if err := svc.Delete(context.Background(), "not-a-uuid"); err == nil {
		t.Fatal("expected error for invalid uuid")
	}
	if err := svc.Delete(context.Background(), "60601fee-2bf1-4721-ae6f-7636e79a0cba"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newSubscription(price int, startDate string, endDate *string) models.Subscription {
	start := mustParseMonth(startDate)
	var end *time.Time
	if endDate != nil {
		parsedEnd := mustParseMonth(*endDate)
		end = &parsedEnd
	}

	return models.Subscription{
		ServiceName: "Yandex Plus",
		Price:       price,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   start,
		EndDate:     end,
	}
}

func mustParseMonth(value string) time.Time {
	parsed, err := time.Parse("01-2006", value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}
