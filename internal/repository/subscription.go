package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/subscription_service/internal/models"
)

type SubscriptionRepository struct {
	db *pgx.Conn
}

func NewSubscriptionRepository(db *pgx.Conn) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, now, now).Scan(&sub.ID)
	
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	
	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id = $1`
	
	row := r.db.QueryRow(ctx, query, id)
	sub := &models.Subscription{}
	
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	
	return sub, nil
}

func (r *SubscriptionRepository) GetAll(ctx context.Context) ([]models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions ORDER BY created_at DESC`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	defer rows.Close()
	
	var subscriptions []models.Subscription
	for rows.Next() {
		sub := models.Subscription{}
		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}
	
	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, id string, update *models.SubscriptionUpdateRequest) (*models.Subscription, error) {
	setFields := []string{}
	args := []interface{}{id}
	argPos := 2
	
	if update.ServiceName != nil {
		setFields = append(setFields, fmt.Sprintf("service_name = $%d", argPos))
		args = append(args, *update.ServiceName)
		argPos++
	}
	
	if update.Price != nil {
		setFields = append(setFields, fmt.Sprintf("price = $%d", argPos))
		args = append(args, *update.Price)
		argPos++
	}
	
	if update.EndDate != nil {
		parsedEnd, err := time.Parse("01-2006", *update.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format")
		}
		setFields = append(setFields, fmt.Sprintf("end_date = $%d", argPos))
		args = append(args, parsedEnd)
		argPos++
	}
	
	if len(setFields) == 0 {
		return r.GetByID(ctx, id)
	}
	
	setFields = append(setFields, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, time.Now())
	argPos++
	
	query := fmt.Sprintf("UPDATE subscriptions SET %s WHERE id = $1 RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at", 
		fmt.Sprintf("%s, ", setFields[:len(setFields)-1]) + setFields[len(setFields)-1])
	
	row := r.db.QueryRow(ctx, query, args...)
	sub := &models.Subscription{}
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}
	
	return sub, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}
	
	return nil
}

func (r *SubscriptionRepository) GetSumByPeriod(ctx context.Context, filter *models.SubscriptionFilter) (int, error) {
	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argPos := 1
	
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *filter.UserID)
		argPos++
	}
	
	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argPos)
		args = append(args, *filter.ServiceName)
		argPos++
	}
	
	var total int
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate sum: %w", err)
	}
	
	return total, nil
}
