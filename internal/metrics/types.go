package metrics

import "context"

// Client defines the ability of fetching a list of active ingresses
// taking in account a maxIdle period
type Client interface {
	ListActiveIngresses(maxIdle string) (map[string]bool, error)
	MetricExists(metricName string) (bool, error)
}

type collector interface {
	getIngresses(ctx context.Context, query string) (ingresses ingressResults, err error)
	getMetric(query string) (bool, error)
}

type ingressResults map[string]bool
