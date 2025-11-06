package subscriptions

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/n-korel/user-subscriptions-api/internal/logger"
)

type SubscriptionService interface {
	GetAllSubscriptions(ctx context.Context) ([]Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int) (*Subscription, error)
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	UpdateSubscription(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error)
	DeleteSubscription(ctx context.Context, id int) error
	GetCostByPeriod(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (*CostResponse, error)
}

type service struct {
	repo SubscriptionRepository
	log  logger.LoggerInterface
}

func NewService(repo SubscriptionRepository, log logger.LoggerInterface) SubscriptionService {
	return &service{repo: repo, log: log}
}

func (s *service) GetAllSubscriptions(ctx context.Context) ([]Subscription, error) {
	return s.repo.GetAll(ctx)
}

func (s *service) GetSubscriptionByID(ctx context.Context, id int) (*Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	if err := s.validateSubscriptionRequest(req); err != nil {
		s.log.Warn("Validation failed", map[string]any{"error": err.Error()})
		return nil, err
	}

	return s.repo.Create(ctx, req)
}

func (s *service) UpdateSubscription(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error) {
	if err := s.validateSubscriptionRequest(CreateSubscriptionRequest{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}); err != nil {
		s.log.Warn("Validation failed", map[string]any{"error": err.Error(), "id": id})
		return nil, err
	}

	return s.repo.Update(ctx, id, req)
}

func (s *service) DeleteSubscription(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetCostByPeriod(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (*CostResponse, error) {
	if startDate == "" && endDate == "" {
		return nil, fmt.Errorf("at least one date parameter is required")
	}

	if err := s.validateDateFormat(startDate); err != nil {
		return nil, err
	}

	if endDate != "" {
		if err := s.validateDateFormat(endDate); err != nil {
			return nil, err
		}
	}

	totalCost, count, err := s.repo.GetCostByPeriod(ctx, startDate, endDate, userID, serviceName)
	if err != nil {
		return nil, err
	}

	return &CostResponse{TotalCost: totalCost, Count: count}, nil
}

func (s *service) validateSubscriptionRequest(req CreateSubscriptionRequest) error {
	if req.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	if req.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required and must be valid UUID")
	}

	if err := s.validateDateFormat(req.StartDate); err != nil {
		return err
	}

	if req.EndDate != nil && *req.EndDate != "" {
		if err := s.validateDateFormat(*req.EndDate); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) validateDateFormat(date string) error {
	if date == "" {
		return fmt.Errorf("date cannot be empty")
	}

	pattern := `^\d{2}-\d{4}$`
	matched, err := regexp.MatchString(pattern, date)
	if err != nil || !matched {
		return fmt.Errorf("date must be in MM-YYYY format")
	}

	return nil
}
