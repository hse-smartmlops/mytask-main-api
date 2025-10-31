package db

import (
	"log"
	"strings"

	"gorm.io/gorm"
)

func SetupFullTextSearch(db *gorm.DB) error {
    log.Println("Setting up Full-Text Search extensions and indexes...")
    
    // Создаем расширения
    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS unaccent").Error; err != nil {
        log.Printf("Warning: Failed to create unaccent extension: %v", err)
    }
    
    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
        log.Printf("Warning: Failed to create pg_trgm extension: %v", err)
    }

    // Индексы с правильными именами таблиц (как GORM их создает)
    indexes := []string{
        // === Users ===
        "CREATE INDEX IF NOT EXISTS idx_users_first_name_fts ON users USING gin(to_tsvector('russian', first_name))",
        "CREATE INDEX IF NOT EXISTS idx_users_last_name_fts ON users USING gin(to_tsvector('russian', last_name))",
        "CREATE INDEX IF NOT EXISTS idx_users_full_name_fts ON users USING gin(to_tsvector('russian', first_name || ' ' || last_name))",
        "CREATE INDEX IF NOT EXISTS idx_users_email_fts ON users USING gin(to_tsvector('russian', email))",
        "CREATE INDEX IF NOT EXISTS idx_users_profession_fts ON users USING gin(to_tsvector('russian', profession))",

        // === Tasks ===
        "CREATE INDEX IF NOT EXISTS idx_tasks_name_fts ON tasks USING gin(to_tsvector('russian', name))",
        "CREATE INDEX IF NOT EXISTS idx_tasks_description_fts ON tasks USING gin(to_tsvector('russian', description))",

        // === Projects ===
        "CREATE INDEX IF NOT EXISTS idx_projects_name_fts ON projects USING gin(to_tsvector('russian', name))",
        "CREATE INDEX IF NOT EXISTS idx_projects_description_fts ON projects USING gin(to_tsvector('russian', description))",

        // === Problems ===
        "CREATE INDEX IF NOT EXISTS idx_problems_name_fts ON problems USING gin(to_tsvector('russian', name))",

        // === Forum Messages ===
        "CREATE INDEX IF NOT EXISTS idx_forum_messages_description_fts ON forum_messages USING gin(to_tsvector('russian', description[1]))",

        // === Help Requests ===
        "CREATE INDEX IF NOT EXISTS idx_help_requests_description_fts ON help_requests USING gin(to_tsvector('russian', description))",

        // === Completed Work === (GORM создает таблицу как completed_works)
        "CREATE INDEX IF NOT EXISTS idx_completed_works_description_fts ON completed_works USING gin(to_tsvector('russian', description))",

        // === Tomorrow Plans === (GORM создает таблицу как tomorrow_plans)
        "CREATE INDEX IF NOT EXISTS idx_tomorrow_plans_description_fts ON tomorrow_plans USING gin(to_tsvector('russian', description))",

        // === Teams ===
        "CREATE INDEX IF NOT EXISTS idx_teams_name_fts ON teams USING gin(to_tsvector('russian', name))",
        "CREATE INDEX IF NOT EXISTS idx_teams_description_fts ON teams USING gin(to_tsvector('russian', description))",
    }

    successCount := 0
    for _, indexSQL := range indexes {
        if err := db.Exec(indexSQL).Error; err != nil {
            log.Printf("Warning: Failed to create index %s: %v", getIndexName(indexSQL), err)
        } else {
            successCount++
        }
    }
    
    log.Printf("Full-Text Search setup completed: %d/%d indexes created", successCount, len(indexes))
    return nil
}

func getIndexName(sql string) string {
    start := strings.Index(sql, "idx_")
    if start == -1 {
        return "unknown"
    }
    end := strings.Index(sql[start:], " ")
    if end == -1 {
        return sql[start:]
    }
    return sql[start:start+end]
}