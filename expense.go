package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExpenseRecord struct {
	ID       int64
	Datetime time.Time
	ItemName string
	Category string
	Quantity float64
	Price    float64
}

type ExpenseStore struct {
	db *sql.DB
}

func NewExpenseStore(dbPath string) (*ExpenseStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &ExpenseStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

func (s *ExpenseStore) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS expense_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		datetime TEXT NOT NULL,
		item_name TEXT NOT NULL,
		category TEXT NOT NULL,
		quantity REAL NOT NULL,
		price REAL NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_expense_records_datetime ON expense_records(datetime DESC);
	CREATE INDEX IF NOT EXISTS idx_expense_records_category ON expense_records(category);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *ExpenseStore) Close() error {
	return s.db.Close()
}

func (s *ExpenseStore) Create(itemName string, category string, quantity float64, price float64) (*ExpenseRecord, error) {
	now := time.Now()
	query := `INSERT INTO expense_records (datetime, item_name, category, quantity, price) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, now.Format(time.RFC3339), itemName, category, quantity, price)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &ExpenseRecord{
		ID:       id,
		Datetime: now,
		ItemName: itemName,
		Category: category,
		Quantity: quantity,
		Price:    price,
	}, nil
}

func (s *ExpenseStore) GetByID(id int64) (*ExpenseRecord, error) {
	query := `SELECT id, datetime, item_name, category, quantity, price FROM expense_records WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var record ExpenseRecord
	var datetimeStr string
	err := row.Scan(&record.ID, &datetimeStr, &record.ItemName, &record.Category, &record.Quantity, &record.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get record: %w", err)
	}

	record.Datetime, err = time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime: %w", err)
	}

	return &record, nil
}

func (s *ExpenseStore) GetAll() ([]ExpenseRecord, error) {
	query := `SELECT id, datetime, item_name, category, quantity, price FROM expense_records ORDER BY datetime DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []ExpenseRecord
	for rows.Next() {
		var record ExpenseRecord
		var datetimeStr string
		if err := rows.Scan(&record.ID, &datetimeStr, &record.ItemName, &record.Category, &record.Quantity, &record.Price); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		record.Datetime, err = time.Parse(time.RFC3339, datetimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse datetime: %w", err)
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (s *ExpenseStore) Update(id int64, itemName string, category string, quantity float64, price float64) (*ExpenseRecord, error) {
	existing, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("record not found")
	}

	query := `UPDATE expense_records SET item_name = ?, category = ?, quantity = ?, price = ? WHERE id = ?`
	_, err = s.db.Exec(query, itemName, category, quantity, price, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	existing.ItemName = itemName
	existing.Category = category
	existing.Quantity = quantity
	existing.Price = price
	return existing, nil
}

func (s *ExpenseStore) Delete(id int64) error {
	query := `DELETE FROM expense_records WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("record not found")
	}

	return nil
}

func (s *ExpenseStore) GetTotalSpentCurrentMonth() (float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	query := `SELECT SUM(quantity * price) FROM expense_records WHERE datetime >= ?`
	row := s.db.QueryRow(query, startOfMonth.Format(time.RFC3339))

	var total sql.NullFloat64
	err := row.Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total: %w", err)
	}

	if !total.Valid {
		return 0, nil
	}

	return total.Float64, nil
}

func (s *ExpenseStore) GetByCategoryPast30Days(category string) ([]ExpenseRecord, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	query := `SELECT id, datetime, item_name, category, quantity, price FROM expense_records WHERE category = ? AND datetime >= ? ORDER BY datetime DESC`
	rows, err := s.db.Query(query, category, thirtyDaysAgo.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []ExpenseRecord
	for rows.Next() {
		var record ExpenseRecord
		var datetimeStr string
		if err := rows.Scan(&record.ID, &datetimeStr, &record.ItemName, &record.Category, &record.Quantity, &record.Price); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		record.Datetime, err = time.Parse(time.RFC3339, datetimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse datetime: %w", err)
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func SortExpenseRecordsByDateTime(records []ExpenseRecord, descending bool) []ExpenseRecord {
	sorted := make([]ExpenseRecord, len(records))
	copy(sorted, records)

	if descending {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Datetime.After(sorted[j].Datetime)
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Datetime.Before(sorted[j].Datetime)
		})
	}

	return sorted
}

func FormatExpenseRecord(record *ExpenseRecord) string {
	total := record.Quantity * record.Price
	return fmt.Sprintf(
		"ID: %d\nDate: %s\nItem: %s\nCategory: %s\nQuantity: %.2f\nPrice: %.2f\nTotal: %.2f",
		record.ID,
		record.Datetime.In(DisplayLocation).Format("2006-01-02 15:04"),
		record.ItemName,
		record.Category,
		record.Quantity,
		record.Price,
		total,
	)
}

func FormatExpenseRecordsList(records []ExpenseRecord) string {
	if len(records) == 0 {
		return "No expense records found."
	}

	var result string
	var grandTotal float64
	for i, record := range records {
		total := record.Quantity * record.Price
		grandTotal += total
		result += fmt.Sprintf("%d. %s - %s (%s) x%.0f @%.2f = %.2f\n",
			i+1,
			record.Datetime.In(DisplayLocation).Format("2006-01-02 15:04"),
			record.ItemName,
			record.Category,
			record.Quantity,
			record.Price,
			total)
	}
	result += fmt.Sprintf("\nGrand Total: %.2f", grandTotal)
	return result
}

func init() {
	log.Println("Expense module initialized")
}
