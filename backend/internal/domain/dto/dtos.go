package dto

import "time"

// WindowsArtifact представляет системный артефакт Windows
type WindowsArtifactDTO struct {
	UUID        string                 `json:"id"`
	HostID      string                 `json:"host_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	Path        string                 `json:"path"`
	Value       string                 `json:"value"`
	Size        int64                  `json:"size"`
	Hash        string                 `json:"hash"`
	Owner       string                 `json:"owner"`
	Permissions string                 `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
	ThreatLevel int                    `json:"threat_level"`
}

type IncidentDTO struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"` // brute_force, ai_anomaly
	SourceIP    string    `json:"source_ip"`
	ThreatLevel int       `json:"threat_level"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
