package repositories

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slimebot/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ensureSessionMemoriesFTS(db *gorm.DB) error {
	if err := db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS session_memories_fts USING fts5(session_id UNINDEXED, body, tokenize='unicode61')`).Error; err != nil {
		slog.Warn("session_memories_fts_unavailable", "err", err)
		return nil
	}
	var cnt int64
	if err := db.Raw(`SELECT COUNT(*) FROM session_memories_fts`).Scan(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	return db.Exec(`
INSERT INTO session_memories_fts(session_id, body)
SELECT session_id, TRIM(keywords_text || ' ' || summary) FROM session_memories
WHERE TRIM(COALESCE(keywords_text,'') || COALESCE(summary,'')) <> ''
`).Error
}

func NewSQLite(dbPath string) (*gorm.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(absPath), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&domain.Session{},
		&domain.Message{},
		&domain.SessionMemory{},
		&domain.ToolCallRecord{},
		&domain.AppSetting{},
		&domain.LLMConfig{},
		&domain.MCPConfig{},
		&domain.MessagePlatformConfig{},
		&domain.Skill{},
	); err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}
	_ = ensureSessionMemoriesFTS(db)

	return db, nil
}
