package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const fredBaseURL = "https://api.stlouisfed.org/fred"

// FREDClient fetches data from the FRED API.
type FREDClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewFREDClient(apiKey string) *FREDClient {
	return &FREDClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FREDObservation is a single data point from FRED.
type FREDObservation struct {
	RealtimeStart string `json:"realtime_start"`
	RealtimeEnd   string `json:"realtime_end"`
	Date          string `json:"date"`
	Value         string `json:"value"`
}

// FREDObservationsResponse is the FRED API response for series observations.
type FREDObservationsResponse struct {
	RealtimeStart    string             `json:"realtime_start"`
	RealtimeEnd      string             `json:"realtime_end"`
	ObservationStart string             `json:"observation_start"`
	ObservationEnd   string             `json:"observation_end"`
	Units            string             `json:"units"`
	OutputType       int                `json:"output_type"`
	FileType         string             `json:"file_type"`
	OrderBy          string             `json:"order_by"`
	SortOrder        string             `json:"sort_order"`
	Count            int                `json:"count"`
	Offset           int                `json:"offset"`
	Limit            int                `json:"limit"`
	Observations     []FREDObservation  `json:"observations"`
}

// FetchObservations retrieves observations for a FRED series within a date range.
func (c *FREDClient) FetchObservations(ctx context.Context, seriesID string, startDate, endDate time.Time) (*FREDObservationsResponse, []byte, error) {
	params := url.Values{}
	params.Set("series_id", seriesID)
	params.Set("api_key", c.apiKey)
	params.Set("file_type", "json")
	params.Set("observation_start", startDate.Format("2006-01-02"))
	params.Set("observation_end", endDate.Format("2006-01-02"))
	params.Set("sort_order", "asc")

	reqURL := fmt.Sprintf("%s/series/observations?%s", fredBaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}

	slog.Info("fetching FRED observations", "series_id", seriesID, "start", startDate.Format("2006-01-02"), "end", endDate.Format("2006-01-02"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("FRED API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result FREDObservationsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, fmt.Errorf("parsing response: %w", err)
	}

	slog.Info("fetched FRED observations", "series_id", seriesID, "count", result.Count)

	return &result, body, nil
}

// FetchALFREDObservations retrieves vintage observations from ALFRED for point-in-time reconstruction.
func (c *FREDClient) FetchALFREDObservations(ctx context.Context, seriesID string, startDate, endDate time.Time, realtimeStart, realtimeEnd time.Time) (*FREDObservationsResponse, []byte, error) {
	params := url.Values{}
	params.Set("series_id", seriesID)
	params.Set("api_key", c.apiKey)
	params.Set("file_type", "json")
	params.Set("observation_start", startDate.Format("2006-01-02"))
	params.Set("observation_end", endDate.Format("2006-01-02"))
	params.Set("realtime_start", realtimeStart.Format("2006-01-02"))
	params.Set("realtime_end", realtimeEnd.Format("2006-01-02"))
	params.Set("sort_order", "asc")

	reqURL := fmt.Sprintf("%s/series/observations?%s", fredBaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}

	slog.Info("fetching ALFRED observations", "series_id", seriesID,
		"realtime_start", realtimeStart.Format("2006-01-02"),
		"realtime_end", realtimeEnd.Format("2006-01-02"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("ALFRED API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result FREDObservationsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, fmt.Errorf("parsing response: %w", err)
	}

	slog.Info("fetched ALFRED observations", "series_id", seriesID, "count", result.Count)

	return &result, body, nil
}
