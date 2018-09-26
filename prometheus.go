package main

import (
	"context"
	"fmt"
	"time"

	api "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
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

type ingressResults map[string]int

// type ingressResult struct {
// Ingress  string
// Requests model.SampleValue
// }

func (c PrometheusClient) getIngresses(ctx context.Context, query string) (ingresses ingressResults, err error) {
	results := make(map[string]int)

	val, err := c.client.Query(ctx, query, time.Now())
	if err != nil {
		return results, err
	}

	for _, sample := range val.(model.Vector) {
		if ingress, ok := sample.Metric["ingress"]; ok {
			if namespace, ok := sample.Metric["exported_namespace"]; ok {
				value := float64(sample.Value)
				if value > 0 {
					results[fmt.Sprintf("%s/%s", namespace, ingress)] = 1
				}
			}
		}
	}
	return results, nil
}

func checkPrometheusMetrics(ctx context.Context, collector IngressCollector) (map[string]int, error) {
	//"rate(nginx_ingress_controller_requests{status=\"200\"}[12h])"
	results, err := collector.getIngresses(ctx, "sum(rate(nginx_ingress_controller_requests{status=\"200\"}[12h])) by (ingress,exported_namespace)")
	if err != nil {
		logrus.Errorf("Could not check prometheus metrics:%v", err)
		return results, err
	}
	logrus.Infof("ActiveIngresses total:%d", len(results))
	return results, nil
}
