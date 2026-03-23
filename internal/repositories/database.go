package repositories

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slimebot/internal/domain"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupSessionMemoryFTS5(db *gorm.DB) error {
	if err := db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS memory_facts_fts USING fts5(
		memory_type,
		subject,
		predicate,
		value,
		summary,
		content='memory_facts',
		content_rowid='rowid',
		tokenize='unicode61'
	)`).Error; err != nil {
		slog.Warn("memory_facts_fts_unavailable", "err", err)
		return nil
	}
	triggers := []string{
		`CREATE TRIGGER IF NOT EXISTS memory_facts_ai AFTER INSERT ON memory_facts BEGIN
			INSERT INTO memory_facts_fts(rowid, memory_type, subject, predicate, value, summary)
			VALUES (new.rowid, new.memory_type, new.subject, new.predicate, new.value, new.summary);
		END`,
		`CREATE TRIGGER IF NOT EXISTS memory_facts_ad AFTER DELETE ON memory_facts BEGIN
			INSERT INTO memory_facts_fts(memory_facts_fts, rowid, memory_type, subject, predicate, value, summary)
			VALUES('delete', old.rowid, old.memory_type, old.subject, old.predicate, old.value, old.summary);
		END`,
		`CREATE TRIGGER IF NOT EXISTS memory_facts_au AFTER UPDATE ON memory_facts BEGIN
			INSERT INTO memory_facts_fts(memory_facts_fts, rowid, memory_type, subject, predicate, value, summary)
			VALUES('delete', old.rowid, old.memory_type, old.subject, old.predicate, old.value, old.summary);
			INSERT INTO memory_facts_fts(rowid, memory_type, subject, predicate, value, summary)
			VALUES (new.rowid, new.memory_type, new.subject, new.predicate, new.value, new.summary);
		END`,
	}
	for _, sql := range triggers {
		if err := db.Exec(sql).Error; err != nil {
			slog.Warn("memory_facts_fts_trigger_failed", "err", err)
			return nil
		}
	}
	return nil
}

func NewSQLite(dbPath string) (*gorm.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(absPath), &gorm.Config{
		PrepareStmt: true,
		Logger:      newGormSlogLogger(200 * time.Millisecond),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&domain.Session{},
		&domain.Message{},
		&domain.MemoryFact{},
		&domain.ToolCallRecord{},
		&domain.AppSetting{},
		&domain.LLMConfig{},
		&domain.MCPConfig{},
		&domain.MessagePlatformConfig{},
		&domain.Skill{},
	); err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}
	_ = setupSessionMemoryFTS5(db)

	return db, nil
}
