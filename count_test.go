package main

import (
	"os"
	"testing"
	"time"
)

func TestCountStore_Create(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	record, err := store.Create("cat food", 100, 20, 10, 7)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	if record.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if record.Name != "cat food" {
		t.Errorf("Expected name 'cat food', got '%s'", record.Name)
	}
	if record.Quantity != 100 {
		t.Errorf("Expected quantity 100, got %d", record.Quantity)
	}
	if record.Threshold != 20 {
		t.Errorf("Expected threshold 20, got %d", record.Threshold)
	}
	if record.UsageCount != 10 {
		t.Errorf("Expected usage count 10, got %d", record.UsageCount)
	}
	if record.UsageDay != 7 {
		t.Errorf("Expected usage day 7, got %d", record.UsageDay)
	}
	if record.Datetime.IsZero() {
		t.Error("Expected non-zero datetime")
	}
}

func TestCountStore_Create_DuplicateName(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	_, err := store.Create("cat food", 100, 20, 10, 7)
	if err != nil {
		t.Fatalf("Failed to create first record: %v", err)
	}

	_, err = store.Create("cat food", 50, 10, 5, 3)
	if err == nil {
		t.Error("Expected error for duplicate name")
	}
}

func TestCountStore_GetByID(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	created, _ := store.Create("cat food", 100, 20, 10, 7)

	record, err := store.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}

	if record == nil {
		t.Fatal("Expected record, got nil")
	}
	if record.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, record.ID)
	}
	if record.Name != "cat food" {
		t.Errorf("Expected name 'cat food', got '%s'", record.Name)
	}
}

func TestCountStore_GetByID_NotFound(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	record, err := store.GetByID(99999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record != nil {
		t.Error("Expected nil record for non-existent ID")
	}
}

func TestCountStore_GetByName(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	_, _ = store.Create("cat food", 100, 20, 10, 7)

	record, err := store.GetByName("cat food")
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}

	if record == nil {
		t.Fatal("Expected record, got nil")
	}
	if record.Name != "cat food" {
		t.Errorf("Expected name 'cat food', got '%s'", record.Name)
	}
}

func TestCountStore_GetByName_NotFound(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	record, err := store.GetByName("nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record != nil {
		t.Error("Expected nil record for non-existent name")
	}
}

func TestCountStore_GetAll(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	store.Create("cat food", 100, 20, 10, 7)
	store.Create("cat litter", 50, 10, 5, 7)

	records, err := store.GetAll()
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

func TestCountStore_GetAll_Empty(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	records, err := store.GetAll()
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}
}

func TestCountStore_Update(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	created, _ := store.Create("cat food", 100, 20, 10, 7)

	updated, err := store.Update(created.ID, "cat food premium", 80, 15, 8, 5)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	if updated.Name != "cat food premium" {
		t.Errorf("Expected name 'cat food premium', got '%s'", updated.Name)
	}
	if updated.Quantity != 80 {
		t.Errorf("Expected quantity 80, got %d", updated.Quantity)
	}
	if updated.Threshold != 15 {
		t.Errorf("Expected threshold 15, got %d", updated.Threshold)
	}
	if updated.UsageCount != 8 {
		t.Errorf("Expected usage count 8, got %d", updated.UsageCount)
	}
	if updated.UsageDay != 5 {
		t.Errorf("Expected usage day 5, got %d", updated.UsageDay)
	}
}

func TestCountStore_Update_NotFound(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	_, err := store.Update(99999, "cat food", 80, 15, 8, 5)
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestCountStore_Delete(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	created, _ := store.Create("cat food", 100, 20, 10, 7)

	err := store.Delete(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	record, _ := store.GetByID(created.ID)
	if record != nil {
		t.Error("Record still exists after deletion")
	}
}

func TestCountStore_Delete_NotFound(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	err := store.Delete(99999)
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestCountStore_CalculateRemainingQuantity(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	store.Create("cat food", 100, 20, 7, 7)

	remaining, err := store.CalculateRemainingQuantity("cat food")
	if err != nil {
		t.Fatalf("Failed to calculate remaining: %v", err)
	}

	if remaining < 99 || remaining > 100 {
		t.Errorf("Expected remaining around 100 (just created), got %d", remaining)
	}
}

func TestCountStore_CalculateRemainingQuantity_OldRecord(t *testing.T) {
	store := setupTestCountStoreWithOldRecord(t, 72*time.Hour, "cat food", 100, 7, 7)
	defer store.Close()

	remaining, err := store.CalculateRemainingQuantity("cat food")
	if err != nil {
		t.Fatalf("Failed to calculate remaining: %v", err)
	}

	if remaining != 97 {
		t.Errorf("Expected remaining 97 (100 - 1*3 days), got %d", remaining)
	}
}

func TestCountStore_CalculateRemainingQuantity_NotFound(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	_, err := store.CalculateRemainingQuantity("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestCountStore_CalculateRemainingQuantity_Depleted(t *testing.T) {
	store := setupTestCountStoreWithOldRecord(t, 500*time.Hour, "cat food", 10, 1, 1)
	defer store.Close()

	remaining, err := store.CalculateRemainingQuantity("cat food")
	if err != nil {
		t.Fatalf("Failed to calculate remaining: %v", err)
	}

	if remaining != 0 {
		t.Errorf("Expected remaining 0 (depleted), got %d", remaining)
	}
}

func TestCountStore_CalculateExhaustionDate(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	store.Create("cat food", 100, 20, 7, 7)

	exhaustionDate, err := store.CalculateExhaustionDate("cat food")
	if err != nil {
		t.Fatalf("Failed to calculate exhaustion date: %v", err)
	}

	if exhaustionDate == nil {
		t.Fatal("Expected exhaustion date, got nil")
	}

	daysUntil := time.Until(*exhaustionDate).Hours() / 24
	if daysUntil < 99 || daysUntil > 101 {
		t.Errorf("Expected exhaustion around 100 days, got %.1f", daysUntil)
	}
}

func TestCountStore_CalculateExhaustionDate_NotFound(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	_, err := store.CalculateExhaustionDate("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestCountStore_CalculateExhaustionDate_NoUsage(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	store.Create("cat food", 100, 20, 0, 0)

	exhaustionDate, err := store.CalculateExhaustionDate("cat food")
	if err != nil {
		t.Fatalf("Failed to calculate exhaustion date: %v", err)
	}

	if exhaustionDate != nil {
		t.Errorf("Expected nil exhaustion date for no usage data, got %v", exhaustionDate)
	}
}

func TestCountStore_CheckIfBelowThreshold(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	store.Create("cat food", 100, 20, 10, 7)

	isBelow, err := store.CheckIfBelowThreshold("cat food")
	if err != nil {
		t.Fatalf("Failed to check threshold: %v", err)
	}

	if isBelow {
		t.Error("Expected false for record above threshold")
	}
}

func TestCountStore_CheckIfBelowThreshold_BelowThreshold(t *testing.T) {
	store := setupTestCountStoreWithOldRecord(t, 90*24*time.Hour, "cat food", 100, 10, 7)
	defer store.Close()

	isBelow, err := store.CheckIfBelowThreshold("cat food")
	if err != nil {
		t.Fatalf("Failed to check threshold: %v", err)
	}

	if !isBelow {
		t.Error("Expected true for record below threshold")
	}
}

func TestCountStore_CheckIfBelowThreshold_NotFound(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	_, err := store.CheckIfBelowThreshold("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestFormatCountRecord(t *testing.T) {
	record := &CountRecord{
		ID:         42,
		Name:       "cat food",
		Quantity:   100,
		Threshold:  20,
		UsageCount: 7,
		UsageDay:   7,
		Datetime:   time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
	}

	result := FormatCountRecord(record)

	if result == "" {
		t.Error("Expected non-empty result")
	}
	if len(result) < 20 {
		t.Errorf("Result too short: %s", result)
	}
}

func TestFormatCountRecordsList(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	store.Create("cat food", 100, 20, 7, 7)
	store.Create("cat litter", 50, 10, 5, 7)

	records, _ := store.GetAll()
	result := FormatCountRecordsList(records, store)

	if result == "" {
		t.Error("Expected non-empty result")
	}
	if len(result) < 10 {
		t.Errorf("Result too short: %s", result)
	}
}

func TestFormatCountRecordsList_Empty(t *testing.T) {
	store := setupTestCountStore(t)
	defer store.Close()

	records, _ := store.GetAll()
	result := FormatCountRecordsList(records, store)

	if result != "No essentials tracked." {
		t.Errorf("Expected 'No essentials tracked.', got '%s'", result)
	}
}

func setupTestCountStore(t *testing.T) *CountStore {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "count_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := NewCountStore(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create store: %v", err)
	}

	t.Cleanup(func() {
		store.Close()
		os.Remove(tmpFile.Name())
	})

	return store
}

func setupTestCountStoreWithOldRecord(t *testing.T, age time.Duration, name string, quantity int, usageCount int, usageDay int) *CountStore {
	t.Helper()
	store := setupTestCountStore(t)

	oldTime := time.Now().Add(-age)
	_, err := store.db.Exec(
		"INSERT INTO count_records (name, quantity, threshold, usage_count, usage_day, datetime) VALUES (?, ?, ?, ?, ?, ?)",
		name, quantity, 20, usageCount, usageDay, oldTime.Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to insert old record: %v", err)
	}

	return store
}
