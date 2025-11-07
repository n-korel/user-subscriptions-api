package subscriptions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/n-korel/user-subscriptions-api/internal/logger"
)

type SubscriptionRepository interface {
	GetAll(ctx context.Context) ([]Subscription, error)
	GetByID(ctx context.Context, id int) (*Subscription, error)
	Create(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	Update(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error)
	Delete(ctx context.Context, id int) error
	GetCostByPeriod(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (int, int, error)
}

type repository struct {
	db  *pgxpool.Pool
	log logger.LoggerInterface
}

func NewRepository(db *pgxpool.Pool, log logger.LoggerInterface) SubscriptionRepository {
	return &repository{db: db, log: log}
}

func (r *repository) GetAll(ctx context.Context) ([]Subscription, error) {
	rows, err := r.db.Query(ctx, "SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions ORDER BY created_at DESC")
	if err != nil {
		r.log.Error("Failed to query subscriptions", map[string]any{"error": err})
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]Subscription, 0)
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			r.log.Error("Failed to scan subscription", map[string]any{"error": err})
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	r.log.Info("Retrieved all subscriptions", map[string]any{"count": len(subscriptions)})
	return subscriptions, nil
}

func (r *repository) GetByID(ctx context.Context, id int) (*Subscription, error) {
	var sub Subscription
	err := r.db.QueryRow(ctx, "SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id = $1", id).
		Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		r.log.Warn("Subscription not found", map[string]any{"id": id})
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	return &sub, nil
}

func (r *repository) Create(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	var sub Subscription
	err := r.db.QueryRow(ctx,
		"INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5) RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at",
		req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate,
	).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		r.log.Error("Failed to create subscription", map[string]any{"error": err, "service": req.ServiceName})
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	r.log.Info("Subscription created", map[string]any{"id": sub.ID, "service": req.ServiceName, "user_id": req.UserID})
	return &sub, nil
}

func (r *repository) Update(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error) {
	var sub Subscription
	err := r.db.QueryRow(ctx,
		"UPDATE subscriptions SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5, updated_at=CURRENT_TIMESTAMP WHERE id=$6 RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at",
		req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate, id,
	).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		r.log.Error("Failed to update subscription", map[string]any{"error": err, "id": id})
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	r.log.Info("Subscription updated", map[string]any{"id": id})
	return &sub, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil {
		r.log.Error("Failed to delete subscription", map[string]any{"error": err, "id": id})
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	if result.RowsAffected() == 0 {
		r.log.Warn("Subscription not found for deletion", map[string]any{"id": id})
		return fmt.Errorf("subscription not found")
	}

	r.log.Info("Subscription deleted", map[string]any{"id": id})
	return nil
}

func (r *repository) GetCostByPeriod(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (int, int, error) {
	query := "SELECT COALESCE(SUM(price), 0) as total_cost, COUNT(*) as count FROM subscriptions WHERE 1=1"
	args := []any{}
	argCount := 1

	if startDate != "" {
		query += fmt.Sprintf(" AND start_date >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate != "" {
		query += fmt.Sprintf(" AND (end_date IS NULL OR end_date >= $%d)", argCount)
		args = append(args, endDate)
		argCount++
	}

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, userID)
		argCount++
	}

	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argCount)
		args = append(args, *serviceName)
	}

	var totalCost, count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&totalCost, &count)
	if err != nil {
		r.log.Error("Failed to calculate cost", map[string]any{"error": err})
		return 0, 0, fmt.Errorf("failed to calculate cost: %w", err)
	}

	r.log.Info("Cost calculated", map[string]any{"total": totalCost, "count": count})
	return totalCost, count, nil
}
