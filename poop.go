package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type PoopRecord struct {
	ID       int64
	ChatID   int64
	Datetime time.Time
	Texture  string
}

type PoopStore struct {
	db *sql.DB
}

func NewPoopStore(dbPath string) (*PoopStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &PoopStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

func (s *PoopStore) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS poop_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		datetime TEXT NOT NULL,
		texture TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_poop_records_chat_datetime ON poop_records(chat_id, datetime DESC);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *PoopStore) Close() error {
	return s.db.Close()
}

func (s *PoopStore) Create(chatID int64, texture string) (*PoopRecord, error) {
	now := time.Now()
	query := `INSERT INTO poop_records (chat_id, datetime, texture) VALUES (?, ?, ?)`
	result, err := s.db.Exec(query, chatID, now.Format(time.RFC3339), texture)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &PoopRecord{
		ID:       id,
		ChatID:   chatID,
		Datetime: now,
		Texture:  texture,
	}, nil
}

func (s *PoopStore) GetByID(id int64, chatID int64) (*PoopRecord, error) {
	query := `SELECT id, chat_id, datetime, texture FROM poop_records WHERE id = ? AND chat_id = ?`
	row := s.db.QueryRow(query, id, chatID)

	var record PoopRecord
	var datetimeStr string
	err := row.Scan(&record.ID, &record.ChatID, &datetimeStr, &record.Texture)
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

func (s *PoopStore) GetAll(chatID int64) ([]PoopRecord, error) {
	query := `SELECT id, chat_id, datetime, texture FROM poop_records WHERE chat_id = ? ORDER BY datetime DESC`
	rows, err := s.db.Query(query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []PoopRecord
	for rows.Next() {
		var record PoopRecord
		var datetimeStr string
		if err := rows.Scan(&record.ID, &record.ChatID, &datetimeStr, &record.Texture); err != nil {
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

func (s *PoopStore) GetRecent(chatID int64, limit int) ([]PoopRecord, error) {
	query := `SELECT id, chat_id, datetime, texture FROM poop_records WHERE chat_id = ? ORDER BY datetime DESC LIMIT ?`
	rows, err := s.db.Query(query, chatID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []PoopRecord
	for rows.Next() {
		var record PoopRecord
		var datetimeStr string
		if err := rows.Scan(&record.ID, &record.ChatID, &datetimeStr, &record.Texture); err != nil {
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

func (s *PoopStore) GetMostRecent(chatID int64) (*PoopRecord, error) {
	records, err := s.GetRecent(chatID, 1)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

func (s *PoopStore) Update(id int64, chatID int64, texture string) (*PoopRecord, error) {
	existing, err := s.GetByID(id, chatID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("record not found")
	}

	query := `UPDATE poop_records SET texture = ? WHERE id = ? AND chat_id = ?`
	_, err = s.db.Exec(query, texture, id, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	existing.Texture = texture
	return existing, nil
}

func (s *PoopStore) Delete(id int64, chatID int64) error {
	query := `DELETE FROM poop_records WHERE id = ? AND chat_id = ?`
	result, err := s.db.Exec(query, id, chatID)
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

func (s *PoopStore) CheckIfOverdue(chatID int64) (*PoopRecord, bool, error) {
	record, err := s.GetMostRecent(chatID)
	if err != nil {
		return nil, false, err
	}
	if record == nil {
		return nil, false, nil
	}

	threeDaysAgo := time.Now().Add(-72 * time.Hour)
	isOverdue := record.Datetime.Before(threeDaysAgo)
	return record, isOverdue, nil
}

func SortRecordsByDateTime(records []PoopRecord, descending bool) []PoopRecord {
	sorted := make([]PoopRecord, len(records))
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

func FormatRecord(record *PoopRecord) string {
	return fmt.Sprintf(
		"ID: %d\nDate: %s\nTexture: %s",
		record.ID,
		record.Datetime.Format("2006-01-02 15:04"),
		record.Texture,
	)
}

func FormatRecordsList(records []PoopRecord) string {
	if len(records) == 0 {
		return "No records found."
	}

	var result string
	for i, record := range records {
		result += fmt.Sprintf("%d. %s - %s\n", i+1, record.Datetime.Format("2006-01-02 15:04"), record.Texture)
	}
	return result
}

func init() {
	log.Println("Poop module initialized")
}
