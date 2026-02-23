package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CountRecord struct {
	ID         int64
	Name       string
	Quantity   int
	Threshold  int
	UsageCount int
	UsageDay   int
	Datetime   time.Time
}

type CountStore struct {
	db *sql.DB
}

func NewCountStore(dbPath string) (*CountStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &CountStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

func (s *CountStore) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS count_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		quantity INTEGER NOT NULL,
		threshold INTEGER NOT NULL,
		usage_count INTEGER NOT NULL,
		usage_day INTEGER NOT NULL,
		datetime TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_count_records_name ON count_records(name);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *CountStore) Close() error {
	return s.db.Close()
}

func (s *CountStore) Create(name string, quantity int, threshold int, usageCount int, usageDay int) (*CountRecord, error) {
	now := time.Now()
	query := `INSERT INTO count_records (name, quantity, threshold, usage_count, usage_day, datetime) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, name, quantity, threshold, usageCount, usageDay, now.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &CountRecord{
		ID:         id,
		Name:       name,
		Quantity:   quantity,
		Threshold:  threshold,
		UsageCount: usageCount,
		UsageDay:   usageDay,
		Datetime:   now,
	}, nil
}

func (s *CountStore) GetByID(id int64) (*CountRecord, error) {
	query := `SELECT id, name, quantity, threshold, usage_count, usage_day, datetime FROM count_records WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var record CountRecord
	var datetimeStr string
	err := row.Scan(&record.ID, &record.Name, &record.Quantity, &record.Threshold, &record.UsageCount, &record.UsageDay, &datetimeStr)
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

func (s *CountStore) GetByName(name string) (*CountRecord, error) {
	query := `SELECT id, name, quantity, threshold, usage_count, usage_day, datetime FROM count_records WHERE name = ?`
	row := s.db.QueryRow(query, name)

	var record CountRecord
	var datetimeStr string
	err := row.Scan(&record.ID, &record.Name, &record.Quantity, &record.Threshold, &record.UsageCount, &record.UsageDay, &datetimeStr)
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

func (s *CountStore) GetAll() ([]CountRecord, error) {
	query := `SELECT id, name, quantity, threshold, usage_count, usage_day, datetime FROM count_records ORDER BY name ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []CountRecord
	for rows.Next() {
		var record CountRecord
		var datetimeStr string
		if err := rows.Scan(&record.ID, &record.Name, &record.Quantity, &record.Threshold, &record.UsageCount, &record.UsageDay, &datetimeStr); err != nil {
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

func (s *CountStore) Update(id int64, name string, quantity int, threshold int, usageCount int, usageDay int) (*CountRecord, error) {
	existing, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("record not found")
	}

	query := `UPDATE count_records SET name = ?, quantity = ?, threshold = ?, usage_count = ?, usage_day = ?, datetime = ? WHERE id = ?`
	now := time.Now()
	_, err = s.db.Exec(query, name, quantity, threshold, usageCount, usageDay, now.Format(time.RFC3339), id)
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	return &CountRecord{
		ID:         id,
		Name:       name,
		Quantity:   quantity,
		Threshold:  threshold,
		UsageCount: usageCount,
		UsageDay:   usageDay,
		Datetime:   now,
	}, nil
}

func (s *CountStore) Delete(id int64) error {
	query := `DELETE FROM count_records WHERE id = ?`
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

func (s *CountStore) CalculateRemainingQuantity(name string) (int, error) {
	record, err := s.GetByName(name)
	if err != nil {
		return 0, err
	}
	if record == nil {
		return 0, fmt.Errorf("record not found")
	}

	if record.UsageDay <= 0 || record.UsageCount <= 0 {
		return record.Quantity, nil
	}

	dailyUsage := float64(record.UsageCount) / float64(record.UsageDay)
	daysSinceUpdate := time.Since(record.Datetime).Hours() / 24
	estimatedUsed := int(math.Round(dailyUsage * daysSinceUpdate))
	remaining := record.Quantity - estimatedUsed

	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

func (s *CountStore) CalculateExhaustionDate(name string) (*time.Time, error) {
	record, err := s.GetByName(name)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("record not found")
	}

	remaining, err := s.CalculateRemainingQuantity(name)
	if err != nil {
		return nil, err
	}

	if remaining <= 0 {
		now := time.Now()
		return &now, nil
	}

	if record.UsageDay <= 0 || record.UsageCount <= 0 {
		return nil, nil
	}

	dailyUsage := float64(record.UsageCount) / float64(record.UsageDay)
	if dailyUsage <= 0 {
		return nil, nil
	}

	daysUntilExhausted := float64(remaining) / dailyUsage
	exhaustionDate := time.Now().Add(time.Duration(daysUntilExhausted*24) * time.Hour)

	return &exhaustionDate, nil
}

func (s *CountStore) CheckIfBelowThreshold(name string) (bool, error) {
	remaining, err := s.CalculateRemainingQuantity(name)
	if err != nil {
		return false, err
	}

	record, err := s.GetByName(name)
	if err != nil {
		return false, err
	}
	if record == nil {
		return false, fmt.Errorf("record not found")
	}

	return remaining <= record.Threshold, nil
}

func FormatCountRecord(record *CountRecord) string {
	return fmt.Sprintf(
		"ID: %d\nName: %s\nQuantity: %d\nThreshold: %d\nUsage: %d per %d day(s)\nLast Updated: %s",
		record.ID,
		record.Name,
		record.Quantity,
		record.Threshold,
		record.UsageCount,
		record.UsageDay,
		record.Datetime.In(DisplayLocation).Format("2006-01-02 15:04"),
	)
}

func FormatCountRecordWithEstimate(record *CountRecord, remaining int, exhaustionDate *time.Time, isBelowThreshold bool) string {
	var status string
	if isBelowThreshold {
		status = " [LOW STOCK]"
	}

	result := fmt.Sprintf(
		"ID: %d\nName: %s\nQuantity: %d (Est. Remaining: %d)\nThreshold: %d%s\nUsage: %d per %d day(s)\nLast Updated: %s",
		record.ID,
		record.Name,
		record.Quantity,
		remaining,
		record.Threshold,
		status,
		record.UsageCount,
		record.UsageDay,
		record.Datetime.In(DisplayLocation).Format("2006-01-02 15:04"),
	)

	if exhaustionDate != nil {
		daysUntil := int(time.Until(*exhaustionDate).Hours() / 24)
		if daysUntil <= 0 {
			result += fmt.Sprintf("\nExhaustion: Already depleted!")
		} else {
			result += fmt.Sprintf("\nExhaustion: %s (%d days)", exhaustionDate.In(DisplayLocation).Format("2006-01-02"), daysUntil)
		}
	} else {
		result += "\nExhaustion: N/A (no usage data)"
	}

	return result
}

func FormatCountRecordsList(records []CountRecord, store *CountStore) string {
	if len(records) == 0 {
		return "No essentials tracked."
	}

	var result string
	for i, record := range records {
		remaining, _ := store.CalculateRemainingQuantity(record.Name)
		isBelowThreshold, _ := store.CheckIfBelowThreshold(record.Name)

		var status string
		if isBelowThreshold {
			status = " [LOW]"
		}

		exhaustionDate, _ := store.CalculateExhaustionDate(record.Name)
		var exhaustionInfo string
		if exhaustionDate != nil {
			daysUntil := int(time.Until(*exhaustionDate).Hours() / 24)
			if daysUntil <= 0 {
				exhaustionInfo = " [DEPLETED]"
			} else {
				exhaustionInfo = fmt.Sprintf(" [%d days left]", daysUntil)
			}
		}

		result += fmt.Sprintf("%d. %s - Qty: %d (Est: %d)%s%s\n",
			i+1, record.Name, record.Quantity, remaining, status, exhaustionInfo)
	}
	return result
}

func init() {
	log.Println("Count module initialized")
}
