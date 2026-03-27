package domain

import (
	"encoding/json"
	"time"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// Alert is a generated regime-conditioned positioning alert.
type Alert struct {
	ID               int64
	CommodityID      int64
	AsOfDate         time.Time
	Severity         Severity
	AlertType        string
	Headline         string
	Summary          string
	ExplanationJSON  json.RawMessage
	RegimeLabel      string
	RegimeConfidence float64
	FinalAlertScore  float64
	IsActive         bool
	CreatedAt        time.Time
}

// ModelConfig is a tunable parameter for the analytics engine.
type ModelConfig struct {
	ID          int64
	ConfigKey   string
	ConfigValue string
	Description *string
	UpdatedAt   time.Time
}
