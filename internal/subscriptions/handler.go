package subscriptions

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/n-korel/user-subscriptions-api/internal/logger"
)

type Handler struct {
	service SubscriptionService
	log     logger.LoggerInterface
}

func NewHandler(service SubscriptionService, log logger.LoggerInterface) *Handler {
	return &Handler{service: service, log: log}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Route("/subscriptions", func(r chi.Router) {
			r.Get("/", h.GetSubscriptions)
			r.Post("/", h.CreateSubscription)
			r.Get("/cost", h.GetCostByPeriod)
			r.Route("/{id}", func(r chi.Router) {
				r.Patch("/", h.UpdateSubscription)
				r.Delete("/", h.DeleteSubscription)
			})
		})
	})
}

// GetSubscriptions godoc
//
//	@Summary		Get all subscriptions
//	@Description	Retrieve all subscriptions
//	@Tags			subscriptions
//	@Produce		json
//	@Success		200	{object}	Response
//	@Router			/subscriptions [get]
func (h *Handler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	h.log.Info("GET /subscriptions", nil)

	subs, err := h.service.GetAllSubscriptions(r.Context())
	if err != nil {
		h.log.Error("Failed to fetch subscriptions", map[string]any{"error": err})
		h.writeJSON(w, http.StatusInternalServerError, Response{Status: "error", Error: "Failed to fetch subscriptions"})
		return
	}

	h.writeJSON(w, http.StatusOK, Response{Status: "success", Data: subs})
}

// CreateSubscription godoc
//
//	@Summary		Create a new subscription
//	@Description	Create a new subscription record
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateSubscriptionRequest	true	"Subscription data"
//	@Success		201		{object}	Response
//	@Failure		400		{object}	Response
//	@Router			/subscriptions [post]
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	h.log.Info("POST /subscriptions", nil)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("Invalid request body", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: "Invalid request body"})
		return
	}

	var req CreateSubscriptionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.log.Error("Invalid JSON", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: "Invalid JSON"})
		return
	}

	sub, err := h.service.CreateSubscription(r.Context(), req)
	if err != nil {
		h.log.Error("Failed to create subscription", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: err.Error()})
		return
	}

	h.log.Info("Subscription created successfully", map[string]any{"id": sub.ID})
	h.writeJSON(w, http.StatusCreated, Response{Status: "success", Data: sub})
}

// UpdateSubscription godoc
//
//	@Summary		Update a subscription
//	@Description	Update an existing subscription
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Subscription ID"
//	@Param			request	body		UpdateSubscriptionRequest	true	"Subscription data"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		404		{object}	Response
//	@Router			/subscriptions/{id} [patch]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Error("Invalid subscription ID", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: "Invalid subscription ID"})
		return
	}

	h.log.Info("PATCH /subscriptions/{id}", map[string]any{"id": id})

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("Invalid request body", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: "Invalid request body"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.log.Error("Invalid JSON", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: "Invalid JSON"})
		return
	}

	sub, err := h.service.UpdateSubscription(r.Context(), id, req)
	if err != nil {
		h.log.Error("Failed to update subscription", map[string]any{"error": err, "id": id})
		h.writeJSON(w, http.StatusNotFound, Response{Status: "error", Error: err.Error()})
		return
	}

	h.log.Info("Subscription updated successfully", map[string]any{"id": id})
	h.writeJSON(w, http.StatusOK, Response{Status: "success", Data: sub})
}

// DeleteSubscription godoc
//
//	@Summary		Delete a subscription
//	@Description	Delete an existing subscription
//	@Tags			subscriptions
//	@Produce		json
//	@Param			id	path		int	true	"Subscription ID"
//	@Success		200	{object}	Response
//	@Failure		404	{object}	Response
//	@Router			/subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Error("Invalid subscription ID", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: "Invalid subscription ID"})
		return
	}

	h.log.Info("DELETE /subscriptions/{id}", map[string]any{"id": id})

	err = h.service.DeleteSubscription(r.Context(), id)
	if err != nil {
		h.log.Error("Failed to delete subscription", map[string]any{"error": err, "id": id})
		h.writeJSON(w, http.StatusNotFound, Response{Status: "error", Error: err.Error()})
		return
	}

	h.log.Info("Subscription deleted successfully", map[string]any{"id": id})
	h.writeJSON(w, http.StatusOK, Response{Status: "success", Data: map[string]string{"message": "Subscription deleted"}})
}

// GetCostByPeriod godoc
//
//	@Summary		Get subscriptions cost by period
//	@Description	Calculate total cost of subscriptions for a given period with optional filters
//	@Tags			subscriptions
//	@Produce		json
//	@Param			start_date		query		string	true	"Start date (MM-YYYY format)"
//	@Param			end_date		query		string	false	"End date (MM-YYYY format)"
//	@Param			user_id			query		string	false	"User ID (UUID)"
//	@Param			service_name	query		string	false	"Service name"
//	@Success		200				{object}	Response
//	@Failure		400				{object}	Response
//	@Router			/subscriptions/cost [get]
func (h *Handler) GetCostByPeriod(w http.ResponseWriter, r *http.Request) {
	h.log.Info("GET /subscriptions/cost", nil)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	userIDStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	var userID *uuid.UUID
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			h.log.Error("Invalid user ID format", map[string]any{"error": err, "user_id": userIDStr})
			h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: "Invalid user ID format"})
			return
		}
		userID = &uid
	}

	var serviceNamePtr *string
	if serviceName != "" {
		serviceNamePtr = &serviceName
	}

	cost, err := h.service.GetCostByPeriod(r.Context(), startDate, endDate, userID, serviceNamePtr)
	if err != nil {
		h.log.Error("Failed to calculate cost", map[string]any{"error": err})
		h.writeJSON(w, http.StatusBadRequest, Response{Status: "error", Error: err.Error()})
		return
	}

	h.log.Info("Cost calculated successfully", map[string]any{"total": cost.TotalCost, "count": cost.Count})
	h.writeJSON(w, http.StatusOK, Response{Status: "success", Data: cost})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
