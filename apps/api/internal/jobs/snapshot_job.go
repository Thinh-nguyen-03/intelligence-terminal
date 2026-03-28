package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/analytics"
	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// SnapshotJob computes macro factor snapshots, commodity signal snapshots, and alerts.
type SnapshotJob struct {
	macroRepo     *storage.MacroRepo
	snapshotRepo  *storage.SnapshotRepo
	signalRepo    *storage.SignalRepo
	alertRepo     *storage.AlertRepo
	cotRepo       *storage.COTRepo
	configRepo    *storage.ConfigRepo
	sourceRunRepo *storage.SourceRunRepo
}

func NewSnapshotJob(
	macroRepo *storage.MacroRepo,
	snapshotRepo *storage.SnapshotRepo,
	signalRepo *storage.SignalRepo,
	alertRepo *storage.AlertRepo,
	cotRepo *storage.COTRepo,
	configRepo *storage.ConfigRepo,
	sourceRunRepo *storage.SourceRunRepo,
) *SnapshotJob {
	return &SnapshotJob{
		macroRepo:     macroRepo,
		snapshotRepo:  snapshotRepo,
		signalRepo:    signalRepo,
		alertRepo:     alertRepo,
		cotRepo:       cotRepo,
		configRepo:    configRepo,
		sourceRunRepo: sourceRunRepo,
	}
}

// Run computes a factor snapshot for the current date (or a specific date if provided).
func (j *SnapshotJob) Run(ctx context.Context) error {
	return j.RunForDate(ctx, time.Now().UTC())
}

// RunForDate computes a factor snapshot as of a given date.
func (j *SnapshotJob) RunForDate(ctx context.Context, asOf time.Time) error {
	runID, err := j.sourceRunRepo.Start(ctx, domain.SourceAnalytics, "rebuild-snapshots")
	if err != nil {
		return fmt.Errorf("starting run: %w", err)
	}

	records, err := j.computeAll(ctx, asOf)
	if err != nil {
		if failErr := j.sourceRunRepo.Fail(ctx, runID, err.Error()); failErr != nil {
			slog.Error("failed to mark run as failed", "error", failErr)
		}
		return err
	}

	if err := j.sourceRunRepo.Complete(ctx, runID, records, nil); err != nil {
		return fmt.Errorf("completing run: %w", err)
	}

	slog.Info("snapshot computation complete", "as_of", asOf.Format("2006-01-02"), "records", records)
	return nil
}

func (j *SnapshotJob) computeAll(ctx context.Context, asOf time.Time) (int, error) {
	if err := j.computeMacroSnapshot(ctx, asOf); err != nil {
		return 0, fmt.Errorf("macro snapshot: %w", err)
	}

	signalCount, err := j.computeCommoditySignals(ctx, asOf)
	if err != nil {
		return 1, fmt.Errorf("commodity signals: %w", err)
	}

	alertCount, err := j.generateAlerts(ctx, asOf)
	if err != nil {
		slog.Error("alert generation failed", "error", err)
		// Non-fatal: signals were still computed successfully
	}

	return 1 + signalCount + alertCount, nil
}

func (j *SnapshotJob) computeMacroSnapshot(ctx context.Context, asOf time.Time) error {
	// Load model parameters
	params, err := j.configRepo.LoadModelParams(ctx)
	if err != nil {
		return fmt.Errorf("loading model params: %w", err)
	}

	// Load all enabled series
	allSeries, err := j.macroRepo.ListEnabledSeries(ctx)
	if err != nil {
		return fmt.Errorf("listing series: %w", err)
	}

	// Build slug -> series map
	slugToSeries := make(map[string]domain.MacroSeries)
	for _, s := range allSeries {
		slugToSeries[s.Slug] = s
	}

	// Load observations for each series (2 years of data for z-score calculation)
	lookbackStart := asOf.AddDate(-2, 0, 0)
	seriesData := make(analytics.SeriesMap)

	for _, s := range allSeries {
		obs, err := j.macroRepo.ListCleanObservations(ctx, s.ID, lookbackStart, asOf)
		if err != nil {
			slog.Warn("failed to load observations", "slug", s.Slug, "error", err)
			continue
		}
		analytics.SortObservations(obs)
		seriesData[s.Slug] = obs
		slog.Debug("loaded series data", "slug", s.Slug, "count", len(obs))
	}

	// Compute factor scores
	inflationScore := analytics.ComputeInflationScore(seriesData)
	growthScore := analytics.ComputeGrowthScore(seriesData)
	laborScore := analytics.ComputeLaborScore(seriesData)
	stressScore := analytics.ComputeStressScore(seriesData)

	slog.Info("computed factor scores",
		"growth", fmt.Sprintf("%.3f", growthScore),
		"inflation", fmt.Sprintf("%.3f", inflationScore),
		"labor", fmt.Sprintf("%.3f", laborScore),
		"stress", fmt.Sprintf("%.3f", stressScore))

	// Classify regime
	regime := analytics.ClassifyRegime(growthScore, inflationScore, stressScore, params)

	// Get prior regime label for tracking transitions
	var priorLabel *string
	prior, err := j.snapshotRepo.GetLatestSnapshot(ctx, params.ModelVersion)
	if err != nil {
		slog.Warn("failed to get prior snapshot", "error", err)
	}
	if prior != nil {
		priorLabel = &prior.RegimeLabel
	}

	// Build and store snapshot
	snapshot := &domain.MacroFactorSnapshot{
		AsOfDate:         asOf,
		GrowthScore:      growthScore,
		InflationScore:   inflationScore,
		LaborScore:       laborScore,
		StressScore:      stressScore,
		RegimeLabel:      regime.Label,
		RegimePriorLabel: priorLabel,
		Confidence:       regime.Confidence,
		IsTransitioning:  regime.IsTransitioning,
		ModelVersion:     params.ModelVersion,
	}

	if regime.TransitionDetail != "" {
		snapshot.TransitionDetail = &regime.TransitionDetail
	}

	if err := j.snapshotRepo.UpsertFactorSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("storing snapshot: %w", err)
	}

	slog.Info("regime classified",
		"label", regime.Label,
		"confidence", regime.Confidence,
		"transitioning", regime.IsTransitioning)

	return nil
}

func (j *SnapshotJob) computeCommoditySignals(ctx context.Context, asOf time.Time) (int, error) {
	params, err := j.configRepo.LoadModelParams(ctx)
	if err != nil {
		return 0, fmt.Errorf("loading model params: %w", err)
	}

	commodities, err := j.cotRepo.ListActiveCommodities(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing commodities: %w", err)
	}

	if len(commodities) == 0 {
		slog.Warn("no active commodities found for signal computation")
		return 0, nil
	}

	// Load 52+ weeks of COT data for z-score and percentile calculations
	lookbackStart := asOf.AddDate(-2, 0, 0)
	count := 0

	for _, c := range commodities {
		if ctx.Err() != nil {
			return count, ctx.Err()
		}

		positions, err := j.signalRepo.ListCOTPositions(ctx, c.ID, lookbackStart, asOf)
		if err != nil {
			slog.Warn("failed to load COT positions", "commodity", c.Slug, "error", err)
			continue
		}

		if len(positions) == 0 {
			slog.Info("no COT data for commodity", "commodity", c.Slug)
			continue
		}

		sig := analytics.ComputePositionSignal(positions, params)
		if sig == nil {
			continue
		}

		snapshot := &domain.CommoditySignalSnapshot{
			CommodityID:        c.ID,
			AsOfDate:           asOf,
			NetManagedMoney:    sig.NetManagedMoney,
			NetMMPctOI:         sig.NetMMPctOI,
			PositionZScore26W:  sig.ZScore26W,
			PositionZScore52W:  sig.ZScore52W,
			PositionPercentile: sig.Percentile52W,
			WeeklyChangeNetMM:  sig.WeeklyChangeNetMM,
			CrowdingScore:      sig.CrowdingScore,
			SqueezeRiskScore:   sig.SqueezeRiskScore,
			ReversalRiskScore:  sig.ReversalRiskScore,
			TrendSupportScore:  sig.TrendSupportScore,
			ModelVersion:       params.ModelVersion,
		}

		if err := j.signalRepo.UpsertSignalSnapshot(ctx, snapshot); err != nil {
			slog.Warn("failed to store signal snapshot", "commodity", c.Slug, "error", err)
			continue
		}

		slog.Info("computed commodity signal",
			"commodity", c.Slug,
			"net_mm", sig.NetManagedMoney,
			"crowding", fmt.Sprintf("%.2f", sig.CrowdingScore),
			"squeeze", fmt.Sprintf("%.2f", sig.SqueezeRiskScore))

		count++
	}

	return count, nil
}

func (j *SnapshotJob) generateAlerts(ctx context.Context, asOf time.Time) (int, error) {
	params, err := j.configRepo.LoadModelParams(ctx)
	if err != nil {
		return 0, fmt.Errorf("loading model params: %w", err)
	}

	// Get the current regime snapshot
	regime, err := j.snapshotRepo.GetLatestSnapshot(ctx, params.ModelVersion)
	if err != nil || regime == nil {
		return 0, fmt.Errorf("no regime snapshot available for alert generation")
	}

	commodities, err := j.cotRepo.ListActiveCommodities(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing commodities: %w", err)
	}

	// Deactivate older alerts before generating new ones
	if err := j.alertRepo.DeactivateOlderAlerts(ctx, asOf); err != nil {
		slog.Warn("failed to deactivate old alerts", "error", err)
	}

	lookbackStart := asOf.AddDate(-2, 0, 0)
	count := 0

	for _, c := range commodities {
		positions, err := j.signalRepo.ListCOTPositions(ctx, c.ID, lookbackStart, asOf)
		if err != nil || len(positions) == 0 {
			continue
		}

		sig := analytics.ComputePositionSignal(positions, params)
		if sig == nil {
			continue
		}

		input := analytics.AlertInput{
			Commodity:        c,
			Signal:           sig,
			RegimeLabel:      regime.RegimeLabel,
			RegimeConfidence: regime.Confidence,
			IsTransitioning:  regime.IsTransitioning,
		}

		alerts := analytics.GenerateAlerts(input, params)
		for _, a := range alerts {
			alert := &domain.Alert{
				CommodityID:      a.CommodityID,
				AsOfDate:         asOf,
				Severity:         a.Severity,
				AlertType:        a.AlertType,
				Headline:         a.Headline,
				Summary:          a.Summary,
				ExplanationJSON:  a.ExplanationJSON,
				RegimeLabel:      a.RegimeLabel,
				RegimeConfidence: a.RegimeConfidence,
				FinalAlertScore:  a.FinalAlertScore,
				IsActive:         true,
			}

			if _, err := j.alertRepo.InsertAlert(ctx, alert); err != nil {
				slog.Warn("failed to insert alert", "commodity", c.Slug, "error", err)
				continue
			}

			slog.Info("generated alert",
				"commodity", c.Slug,
				"severity", a.Severity,
				"type", a.AlertType,
				"score", fmt.Sprintf("%.2f", a.FinalAlertScore))
			count++
		}
	}

	slog.Info("alert generation complete", "count", count)
	return count, nil
}
