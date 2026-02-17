package main

import (
	"os"
	"testing"
	"time"
)

func TestExpenseStore_Create(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	record, err := store.Create(12345, "Cat Food", "food", 2.0, 25.50)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	if record.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if record.ChatID != 12345 {
		t.Errorf("Expected ChatID 12345, got %d", record.ChatID)
	}
	if record.ItemName != "Cat Food" {
		t.Errorf("Expected item name 'Cat Food', got '%s'", record.ItemName)
	}
	if record.Category != "food" {
		t.Errorf("Expected category 'food', got '%s'", record.Category)
	}
	if record.Quantity != 2.0 {
		t.Errorf("Expected quantity 2.0, got %f", record.Quantity)
	}
	if record.Price != 25.50 {
		t.Errorf("Expected price 25.50, got %f", record.Price)
	}
	if record.Datetime.IsZero() {
		t.Error("Expected non-zero datetime")
	}
}

func TestExpenseStore_GetByID(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	created, _ := store.Create(12345, "Litter", "supplies", 1.0, 15.00)

	record, err := store.GetByID(created.ID, 12345)
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}

	if record == nil {
		t.Fatal("Expected record, got nil")
	}
	if record.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, record.ID)
	}
	if record.ItemName != "Litter" {
		t.Errorf("Expected item name 'Litter', got '%s'", record.ItemName)
	}
}

func TestExpenseStore_GetByID_NotFound(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	record, err := store.GetByID(99999, 12345)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record != nil {
		t.Error("Expected nil record for non-existent ID")
	}
}

func TestExpenseStore_GetByID_WrongChatID(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	created, _ := store.Create(12345, "Cat Food", "food", 1.0, 10.00)

	record, err := store.GetByID(created.ID, 99999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if record != nil {
		t.Error("Expected nil record for wrong chat ID")
	}
}

func TestExpenseStore_GetAll(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	store.Create(12345, "Cat Food", "food", 2.0, 25.00)
	store.Create(12345, "Litter", "supplies", 1.0, 15.00)
	store.Create(12345, "Vet Visit", "vet", 1.0, 100.00)

	records, err := store.GetAll(12345)
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
}

func TestExpenseStore_GetAll_Empty(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	records, err := store.GetAll(12345)
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}
}

func TestExpenseStore_GetAll_IsolatedByChatID(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	store.Create(12345, "Cat Food", "food", 1.0, 10.00)
	store.Create(67890, "Dog Food", "food", 1.0, 15.00)

	records1, _ := store.GetAll(12345)
	records2, _ := store.GetAll(67890)

	if len(records1) != 1 {
		t.Errorf("Expected 1 record for chat 12345, got %d", len(records1))
	}
	if len(records2) != 1 {
		t.Errorf("Expected 1 record for chat 67890, got %d", len(records2))
	}
}

func TestExpenseStore_Update(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	created, _ := store.Create(12345, "Cat Food", "food", 2.0, 25.00)

	updated, err := store.Update(created.ID, 12345, "Premium Cat Food", "food", 3.0, 30.00)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	if updated.ItemName != "Premium Cat Food" {
		t.Errorf("Expected item name 'Premium Cat Food', got '%s'", updated.ItemName)
	}
	if updated.Quantity != 3.0 {
		t.Errorf("Expected quantity 3.0, got %f", updated.Quantity)
	}
	if updated.Price != 30.00 {
		t.Errorf("Expected price 30.00, got %f", updated.Price)
	}

	retrieved, _ := store.GetByID(created.ID, 12345)
	if retrieved.ItemName != "Premium Cat Food" {
		t.Errorf("Update not persisted")
	}
}

func TestExpenseStore_Update_NotFound(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	_, err := store.Update(99999, 12345, "Item", "category", 1.0, 10.00)
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestExpenseStore_Update_WrongChatID(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	created, _ := store.Create(12345, "Cat Food", "food", 1.0, 10.00)

	_, err := store.Update(created.ID, 99999, "Item", "category", 1.0, 10.00)
	if err == nil {
		t.Error("Expected error for wrong chat ID")
	}
}

func TestExpenseStore_Delete(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	created, _ := store.Create(12345, "Cat Food", "food", 1.0, 10.00)

	err := store.Delete(created.ID, 12345)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	record, _ := store.GetByID(created.ID, 12345)
	if record != nil {
		t.Error("Record still exists after deletion")
	}
}

func TestExpenseStore_Delete_NotFound(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	err := store.Delete(99999, 12345)
	if err == nil {
		t.Error("Expected error for non-existent record")
	}
}

func TestExpenseStore_Delete_WrongChatID(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	created, _ := store.Create(12345, "Cat Food", "food", 1.0, 10.00)

	err := store.Delete(created.ID, 99999)
	if err == nil {
		t.Error("Expected error for wrong chat ID")
	}
}

func TestExpenseStore_GetTotalSpentCurrentMonth(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	store.Create(12345, "Cat Food", "food", 2.0, 25.00)
	store.Create(12345, "Litter", "supplies", 1.0, 15.00)

	total, err := store.GetTotalSpentCurrentMonth(12345)
	if err != nil {
		t.Fatalf("Failed to get total: %v", err)
	}

	expected := 2.0*25.00 + 1.0*15.00
	if total != expected {
		t.Errorf("Expected total %.2f, got %.2f", expected, total)
	}
}

func TestExpenseStore_GetTotalSpentCurrentMonth_Empty(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	total, err := store.GetTotalSpentCurrentMonth(12345)
	if err != nil {
		t.Fatalf("Failed to get total: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected total 0, got %.2f", total)
	}
}

func TestExpenseStore_GetTotalSpentCurrentMonth_PreviousMonthExcluded(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	lastMonth := time.Now().AddDate(0, -1, 0)
	_, err := store.db.Exec(
		"INSERT INTO expense_records (chat_id, datetime, item_name, category, quantity, price) VALUES (?, ?, ?, ?, ?, ?)",
		12345, lastMonth.Format(time.RFC3339), "Old Food", "food", 1.0, 100.00,
	)
	if err != nil {
		t.Fatalf("Failed to insert old record: %v", err)
	}

	store.Create(12345, "New Food", "food", 1.0, 50.00)

	total, err := store.GetTotalSpentCurrentMonth(12345)
	if err != nil {
		t.Fatalf("Failed to get total: %v", err)
	}

	if total != 50.00 {
		t.Errorf("Expected total 50.00, got %.2f", total)
	}
}

func TestExpenseStore_GetByCategoryPast30Days(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	store.Create(12345, "Cat Food", "food", 2.0, 25.00)
	store.Create(12345, "Treats", "food", 1.0, 10.00)
	store.Create(12345, "Litter", "supplies", 1.0, 15.00)

	records, err := store.GetByCategoryPast30Days(12345, "food")
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 food records, got %d", len(records))
	}

	for _, r := range records {
		if r.Category != "food" {
			t.Errorf("Expected category 'food', got '%s'", r.Category)
		}
	}
}

func TestExpenseStore_GetByCategoryPast30Days_NoMatch(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	store.Create(12345, "Cat Food", "food", 1.0, 10.00)

	records, err := store.GetByCategoryPast30Days(12345, "vet")
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records for non-matching category, got %d", len(records))
	}
}

func TestExpenseStore_GetByCategoryPast30Days_OldRecordsExcluded(t *testing.T) {
	store := setupExpenseTestStore(t)
	defer store.Close()

	thirtyOneDaysAgo := time.Now().AddDate(0, 0, -31)
	_, err := store.db.Exec(
		"INSERT INTO expense_records (chat_id, datetime, item_name, category, quantity, price) VALUES (?, ?, ?, ?, ?, ?)",
		12345, thirtyOneDaysAgo.Format(time.RFC3339), "Old Food", "food", 1.0, 10.00,
	)
	if err != nil {
		t.Fatalf("Failed to insert old record: %v", err)
	}

	store.Create(12345, "New Food", "food", 1.0, 15.00)

	records, err := store.GetByCategoryPast30Days(12345, "food")
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 recent food record, got %d", len(records))
	}
	if records[0].ItemName != "New Food" {
		t.Errorf("Expected 'New Food', got '%s'", records[0].ItemName)
	}
}

func TestSortExpenseRecordsByDateTime_Descending(t *testing.T) {
	now := time.Now()
	records := []ExpenseRecord{
		{ID: 1, Datetime: now.Add(-2 * time.Hour)},
		{ID: 2, Datetime: now},
		{ID: 3, Datetime: now.Add(-1 * time.Hour)},
	}

	sorted := SortExpenseRecordsByDateTime(records, true)

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

func TestSortExpenseRecordsByDateTime_Ascending(t *testing.T) {
	now := time.Now()
	records := []ExpenseRecord{
		{ID: 1, Datetime: now.Add(-2 * time.Hour)},
		{ID: 2, Datetime: now},
		{ID: 3, Datetime: now.Add(-1 * time.Hour)},
	}

	sorted := SortExpenseRecordsByDateTime(records, false)

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

func TestSortExpenseRecordsByDateTime_DoesNotModifyOriginal(t *testing.T) {
	now := time.Now()
	records := []ExpenseRecord{
		{ID: 1, Datetime: now.Add(-2 * time.Hour)},
		{ID: 2, Datetime: now},
	}

	SortExpenseRecordsByDateTime(records, true)

	if records[0].ID != 1 {
		t.Error("Original slice was modified")
	}
}

func TestFormatExpenseRecord(t *testing.T) {
	record := &ExpenseRecord{
		ID:       42,
		Datetime: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		ItemName: "Cat Food",
		Category: "food",
		Quantity: 2.0,
		Price:    25.50,
	}

	result := FormatExpenseRecord(record)

	if result == "" {
		t.Error("Expected non-empty result")
	}
	if len(result) < 20 {
		t.Errorf("Result too short: %s", result)
	}
}

func TestFormatExpenseRecordsList(t *testing.T) {
	records := []ExpenseRecord{
		{ID: 1, Datetime: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), ItemName: "Cat Food", Category: "food", Quantity: 2.0, Price: 25.00},
		{ID: 2, Datetime: time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC), ItemName: "Litter", Category: "supplies", Quantity: 1.0, Price: 15.00},
	}

	result := FormatExpenseRecordsList(records)

	if result == "" {
		t.Error("Expected non-empty result")
	}
	if len(result) < 10 {
		t.Errorf("Result too short: %s", result)
	}
}

func TestFormatExpenseRecordsList_Empty(t *testing.T) {
	records := []ExpenseRecord{}

	result := FormatExpenseRecordsList(records)

	if result != "No expense records found." {
		t.Errorf("Expected 'No expense records found.', got '%s'", result)
	}
}

func TestFormatExpenseRecordsList_GrandTotal(t *testing.T) {
	records := []ExpenseRecord{
		{ID: 1, Datetime: time.Now(), ItemName: "Item 1", Category: "food", Quantity: 2.0, Price: 10.00},
		{ID: 2, Datetime: time.Now(), ItemName: "Item 2", Category: "food", Quantity: 1.0, Price: 20.00},
	}

	result := FormatExpenseRecordsList(records)

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func setupExpenseTestStore(t *testing.T) *ExpenseStore {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "expense_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := NewExpenseStore(tmpFile.Name())
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
