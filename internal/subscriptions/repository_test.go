package subscriptions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)


func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	
	dsn := "postgres://postgres:my_pass@localhost:5432/user-subscriptions-api?sslmode=disable"
	
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Skip("Skipping repository tests: no database connection")
		return nil
	}

	if err := db.Ping(context.Background()); err != nil {
		t.Skip("Skipping repository tests: database not available")
		return nil
	}

	_, err = db.Exec(context.Background(), "DELETE FROM subscriptions")
	if err != nil {
		t.Fatalf("Failed to clean test database: %v", err)
	}

	return db
}

func TestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	mockLog := &MockLogger{}
	repo := NewRepository(db, mockLog)

	req := CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}

	sub, err := repo.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Greater(t, sub.ID, 0)
	assert.Equal(t, "Netflix", sub.ServiceName)
	assert.Equal(t, 100, sub.Price)
}

func TestRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	mockLog := &MockLogger{}
	repo := NewRepository(db, mockLog)

	req := CreateSubscriptionRequest{
		ServiceName: "Spotify",
		Price:       50,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}
	if _, err := repo.Create(context.Background(), req); err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	subs, err := repo.GetAll(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, subs)
	assert.Greater(t, len(subs), 0)
}

func TestRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	mockLog := &MockLogger{}
	repo := NewRepository(db, mockLog)

	req := CreateSubscriptionRequest{
		ServiceName: "Apple Music",
		Price:       60,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}
	created, _ := repo.Create(context.Background(), req)

	sub, err := repo.GetByID(context.Background(), created.ID)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, created.ID, sub.ID)
	assert.Equal(t, "Apple Music", sub.ServiceName)
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	mockLog := &MockLogger{}
	repo := NewRepository(db, mockLog)

	createReq := CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}
	created, _ := repo.Create(context.Background(), createReq)

	updateReq := UpdateSubscriptionRequest{
		ServiceName: "Netflix Premium",
		Price:       150,
		UserID:      created.UserID,
		StartDate:   "01-2025",
	}
	updated, err := repo.Update(context.Background(), created.ID, updateReq)

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Netflix Premium", updated.ServiceName)
	assert.Equal(t, 150, updated.Price)
}

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	mockLog := &MockLogger{}
	repo := NewRepository(db, mockLog)

	req := CreateSubscriptionRequest{
		ServiceName: "Disney+",
		Price:       80,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}
	created, _ := repo.Create(context.Background(), req)

	err := repo.Delete(context.Background(), created.ID)

	assert.NoError(t, err)

	sub, _ := repo.GetByID(context.Background(), created.ID)
	assert.Nil(t, sub)
}

func TestRepository_GetCostByPeriod(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	mockLog := &MockLogger{}
	repo := NewRepository(db, mockLog)

	userID := uuid.New()

	if _, err := repo.Create(context.Background(), CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      userID,
		StartDate:   "01-2025",
	}); err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	if _, err := repo.Create(context.Background(), CreateSubscriptionRequest{
		ServiceName: "Spotify",
		Price:       50,
		UserID:      userID,
		StartDate:   "01-2025",
	}); err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	totalCost, count, err := repo.GetCostByPeriod(context.Background(), "01-2025", "12-2025", &userID, nil)

	assert.NoError(t, err)
	assert.Equal(t, 150, totalCost)
	assert.Equal(t, 2, count)
}