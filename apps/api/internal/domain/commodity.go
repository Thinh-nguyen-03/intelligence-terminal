package domain

import "time"

// Commodity represents a supported futures market.
type Commodity struct {
	ID                int64
	Slug              string
	Name              string
	CFTCCommodityCode string
	GroupName         string
	Active            bool
	CreatedAt         time.Time
}

// COTReportRaw is a raw weekly COT file (one row per file, all commodities).
type COTReportRaw struct {
	ID           int64
	ReportDate   time.Time
	ReportType   string
	FileChecksum string
	PayloadText  *string
	IngestedAt   time.Time
}

// COTPositionClean is a normalized positioning row from the Disaggregated report.
type COTPositionClean struct {
	ID                    int64
	CommodityID           int64
	ReportDate            time.Time
	OpenInterest          int64
	ProducerMerchantLong  int64
	ProducerMerchantShort int64
	SwapDealerLong        int64
	SwapDealerShort       int64
	ManagedMoneyLong      int64
	ManagedMoneyShort     int64
	OtherReportableLong   int64
	OtherReportableShort  int64
	NonreportableLong     int64
	NonreportableShort    int64
}

// CommoditySignalSnapshot is a derived positioning score for a commodity on a date.
type CommoditySignalSnapshot struct {
	ID                  int64
	CommodityID         int64
	AsOfDate            time.Time
	NetManagedMoney     int64
	NetMMPctOI          float64
	PositionZScore26W   *float64
	PositionZScore52W   *float64
	PositionPercentile  *float64
	WeeklyChangeNetMM   *int64
	CrowdingScore       float64
	SqueezeRiskScore    float64
	ReversalRiskScore   float64
	TrendSupportScore   float64
	ModelVersion        string
	CreatedAt           time.Time
}
