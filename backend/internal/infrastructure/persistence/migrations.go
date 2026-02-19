package persistence

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
)

// Структуры для миграций
type Event struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID        int64                  `bun:"id,pk,autoincrement"`
	Type      string                 `bun:"type,notnull"`
	Source    string                 `bun:"source,notnull"`
	Timestamp time.Time              `bun:"timestamp,notnull"`
	Data      map[string]interface{} `bun:"data,type:jsonb"`
	Severity  string                 `bun:"severity"`
	HostID    string                 `bun:"host_id,notnull"`
}

type Host struct {
	bun.BaseModel `bun:"table:hosts,alias:h"`

	ID          string    `bun:"id,pk"`
	Hostname    string    `bun:"hostname,notnull"`
	IPAddress   string    `bun:"ip_address"`
	LastSeen    time.Time `bun:"last_seen"`
	OSVersion   string    `bun:"os_version"`
	Status      string    `bun:"status"`
	Description string    `bun:"description"`
}

type NetworkTraffic struct {
	bun.BaseModel `bun:"table:network_traffic,alias:nt"`

	ID          int64     `bun:"id,pk,autoincrement"`
	HostID      string    `bun:"host_id,notnull"`
	Timestamp   time.Time `bun:"timestamp,notnull"`
	SrcIP       string    `bun:"src_ip,notnull"`
	DstIP       string    `bun:"dst_ip,notnull"`
	SrcPort     int       `bun:"src_port"`
	DstPort     int       `bun:"dst_port"`
	Protocol    string    `bun:"protocol"`
	BytesSent   int64     `bun:"bytes_sent"`
	BytesRecv   int64     `bun:"bytes_recv"`
	PacketsSent int64     `bun:"packets_sent"`
	PacketsRecv int64     `bun:"packets_recv"`
	Duration    float64   `bun:"duration"`
}

type WindowsArtifact struct {
	bun.BaseModel `bun:"table:windows_artifacts,alias:wa"`

	ID          int64                  `bun:"id,pk,autoincrement"`
	HostID      string                 `bun:"host_id,notnull"`
	Timestamp   time.Time              `bun:"timestamp,notnull"`
	Type        string                 `bun:"type,notnull"`
	Path        string                 `bun:"path"`
	Value       string                 `bun:"value"`
	Size        int64                  `bun:"size"`
	Hash        string                 `bun:"hash"`
	Owner       string                 `bun:"owner"`
	Permissions string                 `bun:"permissions"`
	Metadata    map[string]interface{} `bun:"metadata,type:jsonb"`
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           string `bun:"id,pk,type:uuid,default:uuid_generate_v4()"` // UUID как строка
	Username     string `bun:"username,unique,notnull"`
	Email        string `bun:"email,unique,notnull"`
	PasswordHash string `bun:"password_hash,notnull"`

	// Поля для 2FA
	TOTPSecret      string   `bun:"totp_secret"`                  // Зашифрованный секрет
	TOTPEnabled     bool     `bun:"totp_enabled,default:false"`   // Включена ли 2FA
	TOTPBackupCodes []string `bun:"totp_backup_codes,type:jsonb"` // Резервные коды в JSON

	IsActive  bool      `bun:"is_active,default:true"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	LastLogin time.Time `bun:"last_login"`
}

type TwoFAToken struct {
	bun.BaseModel `bun:"table:two_fa_tokens,alias:t"`

	ID        int64     `bun:"id,pk,autoincrement"`
	UserID    string    `bun:"user_id,notnull"`
	TokenHash string    `bun:"token_hash,notnull"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
	Used      bool      `bun:"used,notnull,default:false"`
}

type NetworkLog struct {
	bun.BaseModel `bun:"table:network_logs,alias:nl"`

	ID            int64     `bun:"id,pk,autoincrement"`
	SourceIP      string    `bun:"source_ip"`
	DestinationIP string    `bun:"destination_ip"`
	Protocol      string    `bun:"protocol"`
	LogType       string    `bun:"log_type"`
	RawData       string    `bun:"raw_data"`
	Timestamp     time.Time `bun:"timestamp"`
	Severity      string    `bun:"severity"`
}

type AttackPattern struct {
	bun.BaseModel `bun:"table:attack_patterns,alias:ap"`

	ID          int64     `bun:"id,pk,autoincrement"`
	Name        string    `bun:"name,unique"`
	Description string    `bun:"description"`
	MITREID     string    `bun:"mitre_id"`
	Severity    string    `bun:"severity"`
	Indicators  []string  `bun:"indicators,type:jsonb"`
	CreatedAt   time.Time `bun:"created_at,nullzero,default:current_timestamp"`
}

type ThreatReport struct {
	bun.BaseModel `bun:"table:threat_reports,alias:tr"`

	ID                   int64     `bun:"id,pk,autoincrement"`
	Timestamp            time.Time `bun:"timestamp"`
	AnalysisType         string    `bun:"analysis_type"`
	MaliciousProbability float64   `bun:"malicious_probability"`
	DetectedPatterns     []string  `bun:"detected_patterns,type:jsonb"`
	Confidence           float64   `bun:"confidence"`
	RawData              []byte    `bun:"raw_data,type:bytea"`
	ThreatType           string    `bun:"threat_type"`
	Indicators           []string  `bun:"indicators,type:jsonb"`
}

type Incident struct {
	bun.BaseModel `bun:"table:incidents,alias:i"`

	ID          int64     `bun:"id,pk,autoincrement"`
	Type        string    `bun:"type,notnull"`
	SourceIP    string    `bun:"source_ip"`
	ThreatLevel int       `bun:"threat_level"`
	Description string    `bun:"description"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

func AddCategoryColumnToAttackPatterns(ctx context.Context, db *bun.DB) error {
	// Проверяем наличие столбца category
	var exists bool
	err := db.NewRaw(`
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_name = 'attack_patterns' 
            AND column_name = 'category'
        )
    `).Scan(ctx, &exists)

	if err != nil {
		return fmt.Errorf("failed to check if category column exists: %w", err)
	}

	if !exists {
		// Добавляем столбец category, если он не существует
		_, err = db.NewRaw(`
            ALTER TABLE attack_patterns 
            ADD COLUMN category VARCHAR
        `).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to add category column: %w", err)
		}
		log.Println("Added category column to attack_patterns table")
	}

	return nil
}

func Add2FAColumnsToUsers(ctx context.Context, db *bun.DB) error {
	// Проверяем наличие столбцов 2FA
	columnsToCheck := []string{
		"totp_secret",
		"totp_enabled",
		"totp_backup_codes",
		"updated_at",
	}

	for _, column := range columnsToCheck {
		var exists bool
		err := db.NewRaw(`
			SELECT EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_name = 'users' 
				AND column_name = ?
			)
		`, column).Scan(ctx, &exists)

		if err != nil {
			return fmt.Errorf("failed to check if column %s exists: %w", column, err)
		}

		if !exists {
			var alterSQL string
			switch column {
			case "totp_secret":
				alterSQL = `ALTER TABLE users ADD COLUMN totp_secret VARCHAR`
			case "totp_enabled":
				alterSQL = `ALTER TABLE users ADD COLUMN totp_enabled BOOLEAN DEFAULT false`
			case "totp_backup_codes":
				// Правильный синтаксис для JSONB по умолчанию
				alterSQL = `ALTER TABLE users ADD COLUMN totp_backup_codes JSONB DEFAULT '[]'`
			case "updated_at":
				alterSQL = `ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`
			}

			_, err = db.NewRaw(alterSQL).Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to add column %s: %w", column, err)
			}
			log.Printf("Added column %s to users table", column)
		}
	}

	return nil
}

// RunMigrations запускает миграции базы данных
func RunMigrations(ctx context.Context, db *bun.DB) error {
	models := []interface{}{
		(*Event)(nil),
		(*Host)(nil),
		(*NetworkTraffic)(nil),
		(*WindowsArtifact)(nil),
		(*User)(nil),
		(*TwoFAToken)(nil),
		(*AttackPattern)(nil),
		(*ThreatReport)(nil),
		(*NetworkLog)(nil),
		(*Incident)(nil),
	}

	for _, model := range models {
		if _, err := db.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(ctx); err != nil {
			return err
		}
	}

	if err := Add2FAColumnsToUsers(ctx, db); err != nil {
		return fmt.Errorf("failed to add 2FA columns: %w", err)
	}

	// Проверяем наличие столбца threat_level
	var exists bool
	err := db.NewRaw(`
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_name = 'windows_artifacts' 
            AND column_name = 'threat_level'
        )
    `).Scan(ctx, &exists)

	if err == nil && !exists {
		// Добавляем столбец threat_level, если он не существует
		_, err = db.NewRaw(`
            ALTER TABLE windows_artifacts 
            ADD COLUMN IF NOT EXISTS threat_level INT DEFAULT 0
        `).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to add threat_level column: %w", err)
		}
		log.Println("Added threat_level column to windows_artifacts table")
	}

	if err := AddCategoryColumnToAttackPatterns(ctx, db); err != nil {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}
