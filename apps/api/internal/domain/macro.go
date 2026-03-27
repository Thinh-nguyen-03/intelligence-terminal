package domain

import "time"

// MacroSeries represents a supported FRED/ALFRED macro time series.
type MacroSeries struct {
	ID             int64
	Source         string
	SourceSeriesID string
	Slug           string
	Name           string
	Frequency      string
	Units          string
	Enabled        bool
	CreatedAt      time.Time
}

// MacroObservationRaw is the raw payload from FRED/ALFRED.
type MacroObservationRaw struct {
	ID              int64
	SeriesID        int64
	ObservationDate time.Time
	RealtimeStart   *time.Time
	RealtimeEnd     *time.Time
	RawValue        string
	PayloadJSON     []byte
	IngestedAt      time.Time
}

// MacroObservationClean is a cleaned, typed observation.
type MacroObservationClean struct {
	ID              int64
	SeriesID        int64
	ObservationDate time.Time
	Value           float64
	VintageDate     *time.Time
	IsLatest        bool
	Frequency       string
}

// MacroFactorSnapshot is a derived regime classification for a given date.
type MacroFactorSnapshot struct {
	ID               int64
	AsOfDate         time.Time
	GrowthScore      float64
	InflationScore   float64
	LaborScore       float64
	StressScore      float64
	RegimeLabel      string
	RegimePriorLabel *string
	Confidence       float64
	IsTransitioning  bool
	TransitionDetail *string
	ModelVersion     string
	CreatedAt        time.Time
}
