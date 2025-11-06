package subscriptions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockRepository struct {
	GetAllFunc          func(ctx context.Context) ([]Subscription, error)
	GetByIDFunc         func(ctx context.Context, id int) (*Subscription, error)
	CreateFunc          func(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	UpdateFunc          func(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error)
	DeleteFunc          func(ctx context.Context, id int) error
	GetCostByPeriodFunc func(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (int, int, error)
}

func (m *MockRepository) GetAll(ctx context.Context) ([]Subscription, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return []Subscription{}, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id int) (*Subscription, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) Create(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return &Subscription{
		ID:          1,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}, nil
}

func (m *MockRepository) Update(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return &Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}, nil
}

func (m *MockRepository) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRepository) GetCostByPeriod(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (int, int, error) {
	if m.GetCostByPeriodFunc != nil {
		return m.GetCostByPeriodFunc(ctx, startDate, endDate, userID, serviceName)
	}
	return 0, 0, nil
}

type MockLogger struct{}

func (m *MockLogger) Info(message string, fields map[string]any)  {}
func (m *MockLogger) Error(message string, fields map[string]any) {}
func (m *MockLogger) Warn(message string, fields map[string]any)  {}
func (m *MockLogger) Debug(message string, fields map[string]any) {}
func (m *MockLogger) Fatal(message string, fields map[string]any) {}
func (m *MockLogger) Sync() error                                         { return nil }

func TestServiceCreateSubscription_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockLog := &MockLogger{}
	svc := NewService(mockRepo, mockLog)

	req := CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}

	sub, err := svc.CreateSubscription(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, "Netflix", sub.ServiceName)
	assert.Equal(t, 100, sub.Price)
}

func TestCreateSubscription_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateSubscriptionRequest
		errMsg  string
	}{
		{
			name: "Empty service name",
			req: CreateSubscriptionRequest{
				ServiceName: "",
				Price:       100,
				UserID:      uuid.New(),
				StartDate:   "01-2025",
			},
			errMsg: "service_name is required",
		},
		{
			name: "Invalid price",
			req: CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       0,
				UserID:      uuid.New(),
				StartDate:   "01-2025",
			},
			errMsg: "price must be greater than 0",
		},
		{
			name: "Invalid user ID",
			req: CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       100,
				UserID:      uuid.Nil,
				StartDate:   "01-2025",
			},
			errMsg: "user_id is required",
		},
		{
			name: "Invalid date format",
			req: CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       100,
				UserID:      uuid.New(),
				StartDate:   "2025-01",
			},
			errMsg: "date must be in MM-YYYY format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			mockLog := &MockLogger{}
			svc := NewService(mockRepo, mockLog)

			sub, err := svc.CreateSubscription(context.Background(), tt.req)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
			assert.Nil(t, sub)
		})
	}
}

func TestServiceUpdateSubscription_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockLog := &MockLogger{}
	svc := NewService(mockRepo, mockLog)

	req := UpdateSubscriptionRequest{
		ServiceName: "Netflix Premium",
		Price:       150,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}

	sub, err := svc.UpdateSubscription(context.Background(), 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, "Netflix Premium", sub.ServiceName)
}

func TestServiceDeleteSubscription(t *testing.T) {
	mockRepo := &MockRepository{}
	mockLog := &MockLogger{}
	svc := NewService(mockRepo, mockLog)

	err := svc.DeleteSubscription(context.Background(), 1)
	
	assert.NoError(t, err)
}

func TestServiceGetCostByPeriod_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockLog := &MockLogger{}
	svc := NewService(mockRepo, mockLog)

	mockRepo.GetCostByPeriodFunc = func(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (int, int, error) {
		return 1200, 12, nil
	}

	result, err := svc.GetCostByPeriod(context.Background(), "01-2025", "12-2025", nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1200, result.TotalCost)
	assert.Equal(t, 12, result.Count)
}

func TestGetCostByPeriod_Validation(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		errMsg    string
	}{
		{
			name:      "Missing both dates",
			startDate: "",
			endDate:   "",
			errMsg:    "at least one date parameter is required",
		},
		{
			name:      "Invalid date format",
			startDate: "2025-01",
			endDate:   "",
			errMsg:    "date must be in MM-YYYY format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			mockLog := &MockLogger{}
			svc := NewService(mockRepo, mockLog)

			result, err := svc.GetCostByPeriod(context.Background(), tt.startDate, tt.endDate, nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
			assert.Nil(t, result)
		})
	}
}