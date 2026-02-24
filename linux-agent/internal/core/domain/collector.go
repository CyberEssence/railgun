package domain

import (
	"context"
	"fmt"
	"time"
)

// Collector defines the interface for all collectors
type Collector interface {
	// Name returns the collector name
	Name() string

	// Enabled returns whether the collector is enabled
	Enabled() bool

	// Collect gathers data and returns it
	Collect(ctx context.Context) (interface{}, error)

	// Type returns the type of data this collector produces
	Type() string
}

// BaseCollector provides common functionality for collectors
type BaseCollector struct {
	CollectorName string
	CollectorType string
	IsEnabled     bool
	HostID        string
	Hostname      string
}

// Name returns the collector name
func (b *BaseCollector) Name() string {
	return b.CollectorName
}

// Enabled returns whether the collector is enabled
func (b *BaseCollector) Enabled() bool {
	return b.IsEnabled
}

// Type returns the collector type
func (b *BaseCollector) Type() string {
	return b.CollectorType
}

// CollectorConfig holds configuration for collectors
type CollectorConfig struct {
	System    bool          `yaml:"system"`
	Network   bool          `yaml:"network"`
	Processes bool          `yaml:"processes"`
	Security  bool          `yaml:"security"`
	Docker    bool          `yaml:"docker"`
	Interval  time.Duration `yaml:"interval"`
	BatchSize int           `yaml:"batch_size"`
}

// CollectionResult represents the result of a collection operation
type CollectionResult struct {
	CollectorName string
	CollectorType string
	Data          interface{}
	Error         error
	Duration      time.Duration
	Timestamp     time.Time
}

// CollectorStats holds statistics for a collector
type CollectorStats struct {
	Name           string
	Type           string
	TotalRuns      int64
	SuccessfulRuns int64
	FailedRuns     int64
	TotalDuration  time.Duration
	AvgDuration    time.Duration
	LastRun        time.Time
	LastError      string
	LastDataSize   int64
}

// CollectorRegistry defines the interface for collector registry
type CollectorRegistry interface {
	// Register adds a collector to the registry
	Register(collector Collector) error

	// Unregister removes a collector from the registry
	Unregister(name string) error

	// Get returns a collector by name
	Get(name string) (Collector, error)

	// List returns all registered collectors
	List() []Collector

	// ListEnabled returns all enabled collectors
	ListEnabled() []Collector

	// ListByType returns collectors of a specific type
	ListByType(collectorType string) []Collector
}

// CollectorManager defines the interface for managing collectors
type CollectorManager interface {
	// StartAll starts all enabled collectors
	StartAll(ctx context.Context) []CollectionResult

	// Start starts a specific collector
	Start(ctx context.Context, name string) (CollectionResult, error)

	// Stop stops all collectors
	Stop() error

	// GetStats returns statistics for all collectors
	GetStats() map[string]CollectorStats
}

// CollectorError represents an error during collection
type CollectorError struct {
	CollectorName string
	CollectorType string
	Err           error
	Time          time.Time
}

func (e *CollectorError) Error() string {
	return fmt.Sprintf("collector %s (%s) error: %v", e.CollectorName, e.CollectorType, e.Err)
}

// CollectorValidation defines methods for validating collector data
type CollectorValidation interface {
	// Validate checks if the collected data is valid
	Validate(data interface{}) error

	// Sanitize cleans and normalizes the collected data
	Sanitize(data interface{}) interface{}
}

// DataTransformer defines methods for transforming collector data
type DataTransformer interface {
	// Transform converts data to a different format
	Transform(data interface{}) (interface{}, error)

	// Normalize ensures consistent data structure
	Normalize(data interface{}) (interface{}, error)
}

// Common collector types
const (
	CollectorTypeSystem   = "system"
	CollectorTypeNetwork  = "network"
	CollectorTypeProcess  = "process"
	CollectorTypeSecurity = "security"
	CollectorTypeDocker   = "docker"
	CollectorTypeFile     = "file"
	CollectorTypeLog      = "log"
	CollectorTypeCustom   = "custom"
)

// CollectorPriority defines execution priority for collectors
type CollectorPriority int

const (
	PriorityLow CollectorPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// CollectorMetadata contains metadata about a collector
type CollectorMetadata struct {
	Name        string
	Type        string
	Description string
	Version     string
	Author      string
	Priority    CollectorPriority
	Tags        []string
	DependsOn   []string // Names of collectors this one depends on
}

// Schedule defines when a collector should run
type Schedule struct {
	Interval       time.Duration
	CronExpression string
	RunOnStart     bool
	MaxRuns        int64 // 0 means unlimited
	Timeout        time.Duration
}

// CollectorDefinition defines a complete collector with all its attributes
type CollectorDefinition struct {
	Metadata   CollectorMetadata
	Schedule   Schedule
	Config     interface{}
	Validation CollectorValidation
	Transform  DataTransformer
}

// CollectionContext provides context for collection operations
type CollectionContext struct {
	Context     context.Context
	Collector   Collector
	Schedule    Schedule
	Attempt     int
	MaxAttempts int
	StartTime   time.Time
}

// CollectionResponse represents a response from a collector
type CollectionResponse struct {
	Success  bool
	Data     interface{}
	Error    error
	Duration time.Duration
	Warnings []string
	Metadata map[string]interface{}
}

// NewBaseCollector creates a new base collector
func NewBaseCollector(name, collectorType string, enabled bool, hostID, hostname string) *BaseCollector {
	return &BaseCollector{
		CollectorName: name,
		CollectorType: collectorType,
		IsEnabled:     enabled,
		HostID:        hostID,
		Hostname:      hostname,
	}
}

// NewCollectionResult creates a new collection result
func NewCollectionResult(collector Collector, data interface{}, err error, duration time.Duration) CollectionResult {
	return CollectionResult{
		CollectorName: collector.Name(),
		CollectorType: collector.Type(),
		Data:          data,
		Error:         err,
		Duration:      duration,
		Timestamp:     time.Now(),
	}
}

// UpdateStats updates collector statistics
func (s *CollectorStats) UpdateStats(success bool, duration time.Duration, dataSize int64, err error) {
	s.TotalRuns++
	if success {
		s.SuccessfulRuns++
	} else {
		s.FailedRuns++
	}

	s.TotalDuration += duration
	if s.TotalRuns > 0 {
		s.AvgDuration = time.Duration(int64(s.TotalDuration) / s.TotalRuns)
	}

	s.LastRun = time.Now()
	if err != nil {
		s.LastError = err.Error()
	}

	if dataSize > 0 {
		s.LastDataSize = dataSize
	}
}

// String returns a string representation of collector priority
func (p CollectorPriority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParseCollectorPriority parses a string into CollectorPriority
func ParseCollectorPriority(priority string) CollectorPriority {
	switch priority {
	case "low":
		return PriorityLow
	case "normal":
		return PriorityNormal
	case "high":
		return PriorityHigh
	case "critical":
		return PriorityCritical
	default:
		return PriorityNormal
	}
}
