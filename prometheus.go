package main

import (
	"context"
	"fmt"
	"time"

	api "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type IngressCollector interface {
	getIngresses(ctx context.Context, query string) (ingresses ingressResults, err error)
}

type PrometheusClient struct {
	client prometheus.API
}

func NewPrometheusClient() PrometheusClient {
	baseClient, err := api.NewClient(api.Config{Address: "http://localhost:9090"})
	must(err)
	return PrometheusClient{client: prometheus.NewAPI(baseClient)}
}

type ingressResults map[string]float64

// type ingressResult struct {
// Ingress  string
// Requests model.SampleValue
// }

func (c PrometheusClient) getIngresses(ctx context.Context, query string) (ingresses ingressResults, err error) {
	results := make(map[string]float64)

	val, err := c.client.Query(ctx, query, time.Now())
	if err != nil {
		return results, err
	}

	for _, sample := range val.(model.Vector) {
		if ingress, ok := sample.Metric["ingress"]; ok {
			if namespace, ok := sample.Metric["exported_namespace"]; ok {
				results[fmt.Sprintf("%s/%s", namespace, ingress)] = float64(sample.Value)
			}
		}
	}
	return results, nil
}
