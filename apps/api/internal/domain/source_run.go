package domain

import "time"

type Source string

const (
	SourceFRED      Source = "fred"
	SourceALFRED    Source = "alfred"
	SourceCFTC      Source = "cftc"
	SourceAnalytics Source = "analytics"
)

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusRunning JobStatus = "running"
	JobStatusSuccess JobStatus = "success"
	JobStatusFailed  JobStatus = "failed"
)

// SourceRun tracks an ingestion or calculation job execution.
type SourceRun struct {
	ID               int64
	Source           Source
	JobName          string
	Status           JobStatus
	StartedAt        time.Time
	FinishedAt       *time.Time
	RecordsProcessed int
	ErrorMessage     *string
	Checksum         *string
}
