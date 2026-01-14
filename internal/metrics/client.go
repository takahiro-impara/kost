package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client wraps a Prometheus API client
type Client struct {
	api     v1.API
	timeout time.Duration
}

// NewClient creates a new Prometheus client
func NewClient(prometheusURL string, timeoutSeconds int) (*Client, error) {
	apiClient, err := api.NewClient(api.Config{
		Address: prometheusURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	return &Client{
		api:     v1.NewAPI(apiClient),
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

// Query executes a PromQL query at a single point in time
func (c *Client) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	result, warnings, err := c.api.Query(ctx, query, ts)
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}

	if len(warnings) > 0 {
		// Log warnings but don't fail
		fmt.Printf("Prometheus query warnings: %v\n", warnings)
	}

	return result, nil
}

// QueryRange executes a PromQL query over a time range
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (model.Value, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	result, warnings, err := c.api.QueryRange(ctx, query, r)
	if err != nil {
		return nil, fmt.Errorf("prometheus range query failed: %w", err)
	}

	if len(warnings) > 0 {
		// Log warnings but don't fail
		fmt.Printf("Prometheus range query warnings: %v\n", warnings)
	}

	return result, nil
}
