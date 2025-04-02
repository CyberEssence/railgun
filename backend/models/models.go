package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Event представляет собой базовое событие в системе
type Event struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID        int64                  `bun:"id,pk,autoincrement" json:"id"`
	Type      string                 `bun:"type,notnull" json:"type"`
	Source    string                 `bun:"source,notnull" json:"source"`
	Timestamp time.Time              `bun:"timestamp,notnull" json:"timestamp"`
	Data      map[string]interface{} `bun:"data,type:jsonb" json:"data"`
	Severity  string                 `bun:"severity" json:"severity"`
	HostID    string                 `bun:"host_id,notnull" json:"host_id"`
}

// Host представляет систему Windows
type Host struct {
	bun.BaseModel `bun:"table:hosts,alias:h"`

	ID          string    `bun:"id,pk" json:"id"`
	Hostname    string    `bun:"hostname,notnull" json:"hostname"`
	IPAddress   string    `bun:"ip_address" json:"ip_address"`
	LastSeen    time.Time `bun:"last_seen" json:"last_seen"`
	OSVersion   string    `bun:"os_version" json:"os_version"`
	Status      string    `bun:"status" json:"status"`
	Description string    `bun:"description" json:"description"`
}

// NetworkTraffic представляет запись о сетевом трафике
type NetworkTraffic struct {
	bun.BaseModel `bun:"table:network_traffic,alias:nt"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	HostID      string    `bun:"host_id,notnull" json:"host_id"`
	Timestamp   time.Time `bun:"timestamp,notnull" json:"timestamp"`
	SrcIP       string    `bun:"src_ip,notnull" json:"src_ip"`
	DstIP       string    `bun:"dst_ip,notnull" json:"dst_ip"`
	SrcPort     int       `bun:"src_port" json:"src_port"`
	DstPort     int       `bun:"dst_port" json:"dst_port"`
	Protocol    string    `bun:"protocol" json:"protocol"`
	BytesSent   int64     `bun:"bytes_sent" json:"bytes_sent"`
	BytesRecv   int64     `bun:"bytes_recv" json:"bytes_recv"`
	PacketsSent int64     `bun:"packets_sent" json:"packets_sent"`
	PacketsRecv int64     `bun:"packets_recv" json:"packets_recv"`
	Duration    float64   `bun:"duration" json:"duration"`
}

// WindowsArtifact представляет системный артефакт Windows
type WindowsArtifact struct {
	bun.BaseModel `bun:"table:windows_artifacts,alias:wa"`

	ID          int64                  `bun:"id,pk,autoincrement" json:"id"`
	HostID      string                 `bun:"host_id,notnull" json:"host_id"`
	Timestamp   time.Time              `bun:"timestamp,notnull" json:"timestamp"`
	Type        string                 `bun:"type,notnull" json:"type"` // registry, file, process, etc.
	Path        string                 `bun:"path" json:"path"`
	Value       string                 `bun:"value" json:"value"`
	Size        int64                  `bun:"size" json:"size"`
	Hash        string                 `bun:"hash" json:"hash"`
	Owner       string                 `bun:"owner" json:"owner"`
	Permissions string                 `bun:"permissions" json:"permissions"`
	Metadata    map[string]interface{} `bun:"metadata,type:jsonb" json:"metadata"`
}

// RunMigrations запускает миграции базы данных
func RunMigrations(ctx context.Context, db *bun.DB) error {
	models := []interface{}{
		(*Event)(nil),
		(*Host)(nil),
		(*NetworkTraffic)(nil),
		(*WindowsArtifact)(nil),
	}

	for _, model := range models {
		if _, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}
