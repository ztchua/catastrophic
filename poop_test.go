package main

import (
	"os"
	"testing"
	"time"
)

func TestPoopStore_Create(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	record, err := store.Create("solid")
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	if record.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if record.Texture != "solid" {
		t.Errorf("Expected texture 'solid', got '%s'", record.Texture)
	}
	if record.Datetime.IsZero() {
		t.Error("Expected non-zero datetime")
	}
}

func TestPoopStore_GetByID(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	created, _ := store.Create("soft")

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
	if record.Texture != "soft" {
		t.Errorf("Expected texture 'soft', got '%s'", record.Texture)
	}
}

func TestPoopStore_GetByID_NotFound(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	record, err := store.GetByID(99999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record != nil {
		t.Error("Expected nil record for non-existent ID")
	}
}

func TestPoopStore_GetAll(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	store.Create("solid")
	store.Create("soft")
	store.Create("liquid")

	records, err := store.GetAll()
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
}

func TestPoopStore_GetAll_Empty(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	records, err := store.GetAll()
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}
}

func TestPoopStore_GetRecent(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	for i := 0; i < 15; i++ {
		store.Create("texture")
		time.Sleep(1 * time.Millisecond)
	}

	records, err := store.GetRecent(10)
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 10 {
		t.Errorf("Expected 10 records, got %d", len(records))
	}

	for i := 1; i < len(records); i++ {
		if records[i-1].Datetime.Before(records[i].Datetime) {
			t.Error("Records not sorted in descending order")
		}
	}
}

func TestPoopStore_GetRecent_Empty(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	records, err := store.GetRecent(10)
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}
}

func TestPoopStore_GetMostRecent(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	now := time.Now()
	oldTime := now.Add(-1 * time.Hour)

	_, err := store.db.Exec(
		"INSERT INTO poop_records (datetime, texture) VALUES (?, ?)",
		oldTime.Format(time.RFC3339), "first",
	)
	if err != nil {
		t.Fatalf("Failed to insert first record: %v", err)
	}

	_, err = store.db.Exec(
		"INSERT INTO poop_records (datetime, texture) VALUES (?, ?)",
		now.Format(time.RFC3339), "second",
	)
	if err != nil {
		t.Fatalf("Failed to insert second record: %v", err)
	}

	record, err := store.GetMostRecent()
	if err != nil {
		t.Fatalf("Failed to get most recent: %v", err)
	}

	if record == nil {
		t.Fatal("Expected record, got nil")
	}
	if record.Texture != "second" {
		t.Errorf("Expected texture 'second', got '%s'", record.Texture)
	}
}

func TestPoopStore_GetMostRecent_Empty(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	record, err := store.GetMostRecent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record != nil {
		t.Error("Expected nil for empty store")
	}
}

func TestPoopStore_UpdateTexture(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	created, _ := store.Create("solid")

	updated, err := store.UpdateTexture(created.ID, "updated texture")
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	if updated.Texture != "updated texture" {
		t.Errorf("Expected texture 'updated texture', got '%s'", updated.Texture)
	}

	retrieved, _ := store.GetByID(created.ID)
	if retrieved.Texture != "updated texture" {
		t.Errorf("Update not persisted")
	}
}

func TestPoopStore_UpdateTexture_NotFound(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	_, err := store.UpdateTexture(99999, "texture")
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestPoopStore_UpdateDatetime(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	created, _ := store.Create("solid")
	newTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	updated, err := store.UpdateDatetime(created.ID, newTime)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	if !updated.Datetime.Equal(newTime) {
		t.Errorf("Expected datetime %v, got %v", newTime, updated.Datetime)
	}

	retrieved, _ := store.GetByID(created.ID)
	if !retrieved.Datetime.Equal(newTime) {
		t.Errorf("Update not persisted")
	}
}

func TestPoopStore_UpdateDatetime_NotFound(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	newTime := time.Now()
	_, err := store.UpdateDatetime(99999, newTime)
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestPoopStore_Delete(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	created, _ := store.Create("solid")

	err := store.Delete(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	record, _ := store.GetByID(created.ID)
	if record != nil {
		t.Error("Record still exists after deletion")
	}
}

func TestPoopStore_Delete_NotFound(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	err := store.Delete(99999)
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestPoopStore_CheckIfOverdue_NoRecords(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	record, isOverdue, err := store.CheckIfOverdue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record != nil {
		t.Error("Expected nil record")
	}
	if isOverdue {
		t.Error("Expected false for isOverdue")
	}
}

func TestPoopStore_CheckIfOverdue_RecentRecord(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	store.Create("solid")

	record, isOverdue, err := store.CheckIfOverdue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record == nil {
		t.Fatal("Expected record")
	}
	if isOverdue {
		t.Error("Expected false for recent record")
	}
}

func TestPoopStore_CheckIfOverdue_OldRecord(t *testing.T) {
	store := setupTestStoreWithOldRecord(t, 96*time.Hour)
	defer store.Close()

	record, isOverdue, err := store.CheckIfOverdue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record == nil {
		t.Fatal("Expected record")
	}
	if !isOverdue {
		t.Error("Expected true for old record (over 72 hours)")
	}
}

func TestSortRecordsByDateTime_Descending(t *testing.T) {
	now := time.Now()
	records := []PoopRecord{
		{ID: 1, Datetime: now.Add(-2 * time.Hour)},
		{ID: 2, Datetime: now},
		{ID: 3, Datetime: now.Add(-1 * time.Hour)},
	}

	sorted := SortRecordsByDateTime(records, true)

	if sorted[0].ID != 2 {
		t.Errorf("Expected first record ID 2, got %d", sorted[0].ID)
	}
	if sorted[1].ID != 3 {
		t.Errorf("Expected second record ID 3, got %d", sorted[1].ID)
	}
	if sorted[2].ID != 1 {
		t.Errorf("Expected third record ID 1, got %d", sorted[2].ID)
	}
}

func TestSortRecordsByDateTime_Ascending(t *testing.T) {
	now := time.Now()
	records := []PoopRecord{
		{ID: 1, Datetime: now.Add(-2 * time.Hour)},
		{ID: 2, Datetime: now},
		{ID: 3, Datetime: now.Add(-1 * time.Hour)},
	}

	sorted := SortRecordsByDateTime(records, false)

	if sorted[0].ID != 1 {
		t.Errorf("Expected first record ID 1, got %d", sorted[0].ID)
	}
	if sorted[1].ID != 3 {
		t.Errorf("Expected second record ID 3, got %d", sorted[1].ID)
	}
	if sorted[2].ID != 2 {
		t.Errorf("Expected third record ID 2, got %d", sorted[2].ID)
	}
}

func TestSortRecordsByDateTime_DoesNotModifyOriginal(t *testing.T) {
	now := time.Now()
	records := []PoopRecord{
		{ID: 1, Datetime: now.Add(-2 * time.Hour)},
		{ID: 2, Datetime: now},
	}

	SortRecordsByDateTime(records, true)

	if records[0].ID != 1 {
		t.Error("Original slice was modified")
	}
}

func TestFormatRecord(t *testing.T) {
	record := &PoopRecord{
		ID:       42,
		Datetime: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		Texture:  "solid",
	}

	result := FormatRecord(record)

	// 14:30 UTC = 22:30 UTC+8
	expected := "ID: 42\nDate: 2024-01-15 22:30\nTexture: solid"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatRecordsList(t *testing.T) {
	records := []PoopRecord{
		{ID: 1, Datetime: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), Texture: "solid"},
		{ID: 2, Datetime: time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC), Texture: "soft"},
	}

	result := FormatRecordsList(records)

	if result == "" {
		t.Error("Expected non-empty result")
	}
	if len(result) < 10 {
		t.Errorf("Result too short: %s", result)
	}
}

func TestFormatRecordsList_Empty(t *testing.T) {
	records := []PoopRecord{}

	result := FormatRecordsList(records)

	if result != "No records found." {
		t.Errorf("Expected 'No records found.', got '%s'", result)
	}
}

func setupTestStore(t *testing.T) *PoopStore {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "poop_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := NewPoopStore(tmpFile.Name())
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

func setupTestStoreWithOldRecord(t *testing.T, age time.Duration) *PoopStore {
	t.Helper()
	store := setupTestStore(t)

	oldTime := time.Now().Add(-age)
	_, err := store.db.Exec(
		"INSERT INTO poop_records (datetime, texture) VALUES (?, ?)",
		oldTime.Format(time.RFC3339), "old",
	)
	if err != nil {
		t.Fatalf("Failed to insert old record: %v", err)
	}

	return store
}
