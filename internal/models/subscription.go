package models

import "time"

type Subscription struct {
	ID          string     `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name" binding:"required"`
	Price       int        `json:"price" db:"price" binding:"required,min=1"`
	UserID      string     `json:"user_id" db:"user_id" binding:"required"`
	StartDate   time.Time  `json:"start_date" db:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type SubscriptionCreateRequest struct {
	ServiceName string  `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int     `json:"price" binding:"required,min=1" example:"400"`
	UserID      string  `json:"user_id" binding:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type SubscriptionUpdateRequest struct {
	ServiceName *string `json:"service_name" example:"Yandex Plus"`
	Price       *int    `json:"price" example:"400"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type SubscriptionFilter struct {
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

type SubscriptionSumRequest struct {
	UserID      *string `json:"user_id,omitempty" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus"`
	StartDate   string  `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     string  `json:"end_date" binding:"required" example:"12-2025"`
}

type SumResponse struct {
	Total int `json:"total" example:"400"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid UUID format"`
}
