package subscriptions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockService struct {
	GetAllSubscriptionsFunc   func(ctx context.Context) ([]Subscription, error)
	GetSubscriptionByIDFunc   func(ctx context.Context, id int) (*Subscription, error)
	CreateSubscriptionFunc    func(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	UpdateSubscriptionFunc    func(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error)
	DeleteSubscriptionFunc    func(ctx context.Context, id int) error
	GetCostByPeriodFunc       func(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (*CostResponse, error)
}

func (m *MockService) GetAllSubscriptions(ctx context.Context) ([]Subscription, error) {
	if m.GetAllSubscriptionsFunc != nil {
		return m.GetAllSubscriptionsFunc(ctx)
	}
	return []Subscription{}, nil
}

func (m *MockService) GetSubscriptionByID(ctx context.Context, id int) (*Subscription, error) {
	if m.GetSubscriptionByIDFunc != nil {
		return m.GetSubscriptionByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockService) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	if m.CreateSubscriptionFunc != nil {
		return m.CreateSubscriptionFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockService) UpdateSubscription(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error) {
	if m.UpdateSubscriptionFunc != nil {
		return m.UpdateSubscriptionFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockService) DeleteSubscription(ctx context.Context, id int) error {
	if m.DeleteSubscriptionFunc != nil {
		return m.DeleteSubscriptionFunc(ctx, id)
	}
	return nil
}

func (m *MockService) GetCostByPeriod(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (*CostResponse, error) {
	if m.GetCostByPeriodFunc != nil {
		return m.GetCostByPeriodFunc(ctx, startDate, endDate, userID, serviceName)
	}
	return nil, nil
}

func TestGetSubscriptions_Success(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	testSubs := []Subscription{
		{
			ID:          1,
			ServiceName: "Netflix",
			Price:       100,
			UserID:      uuid.New(),
			StartDate:   "01-2025",
		},
	}

	mockService.GetAllSubscriptionsFunc = func(ctx context.Context) ([]Subscription, error) {
		return testSubs, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/subscriptions", nil)
	w := httptest.NewRecorder()

	handler.GetSubscriptions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "success", response.Status)
}

func TestHandlerCreateSubscription_Success(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	reqBody := CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}

	mockService.CreateSubscriptionFunc = func(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
		return &Subscription{
			ID:          1,
			ServiceName: req.ServiceName,
			Price:       req.Price,
			UserID:      req.UserID,
			StartDate:   req.StartDate,
		}, nil
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/subscriptions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateSubscription(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "success", response.Status)
}

func TestCreateSubscription_InvalidJSON(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	req := httptest.NewRequest(http.MethodPost, "/v1/subscriptions", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.CreateSubscription(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "error", response.Status)
	assert.Contains(t, response.Error, "Invalid JSON")
}

func TestHandlerUpdateSubscription_Success(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	reqBody := UpdateSubscriptionRequest{
		ServiceName: "Netflix Premium",
		Price:       150,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}

	mockService.UpdateSubscriptionFunc = func(ctx context.Context, id int, req UpdateSubscriptionRequest) (*Subscription, error) {
		return &Subscription{
			ID:          id,
			ServiceName: req.ServiceName,
			Price:       req.Price,
			UserID:      req.UserID,
			StartDate:   req.StartDate,
		}, nil
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPatch, "/v1/subscriptions/1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.UpdateSubscription(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "success", response.Status)
}

func TestHandlerUpdateSubscription_InvalidID(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	req := httptest.NewRequest(http.MethodPatch, "/v1/subscriptions/invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.UpdateSubscription(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "error", response.Status)
	assert.Contains(t, response.Error, "Invalid subscription ID")
}

func TestHandlerDeleteSubscription_Success(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	mockService.DeleteSubscriptionFunc = func(ctx context.Context, id int) error {
		return nil
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/subscriptions/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.DeleteSubscription(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "success", response.Status)
}

func TestHandlerGetCostByPeriod_Success(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	mockService.GetCostByPeriodFunc = func(ctx context.Context, startDate, endDate string, userID *uuid.UUID, serviceName *string) (*CostResponse, error) {
		return &CostResponse{
			TotalCost: 1200,
			Count:     12,
		}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/subscriptions/cost?start_date=01-2025&end_date=12-2025", nil)
	w := httptest.NewRecorder()

	handler.GetCostByPeriod(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "success", response.Status)
}

func TestGetCostByPeriod_InvalidUserID(t *testing.T) {
	mockService := &MockService{}
	mockLog := &MockLogger{}
	handler := NewHandler(mockService, mockLog)

	req := httptest.NewRequest(http.MethodGet, "/v1/subscriptions/cost?start_date=01-2025&user_id=invalid-uuid", nil)
	w := httptest.NewRecorder()

	handler.GetCostByPeriod(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
    	t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, "error", response.Status)
	assert.Contains(t, response.Error, "Invalid user ID format")
}