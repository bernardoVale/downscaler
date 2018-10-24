package metrics

import (
	"context"
	"fmt"
	"time"

	api "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type (
	prometheusClient struct {
		ctx       context.Context
		collector collector
	}

	prometheusCollector struct {
		client prometheusQuery
	}

	prometheusQuery interface {
		Query(ctx context.Context, query string, ts time.Time) (model.Value, error)
	}

	queryClient struct {
		client prometheus.API
	}
)

func (q queryClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	return q.client.Query(ctx, query, ts)
}

// NewPrometheusClient creates a MetricsClient for a Prometheus backend
func NewPrometheusClient(ctx context.Context, host string) Client {
	baseClient, err := api.NewClient(api.Config{Address: host})
	if err != nil {
		panic(err)
	}
	return prometheusClient{ctx: ctx, collector: prometheusCollector{prometheus.NewAPI(baseClient)}}
}

func (c prometheusClient) MetricExists(metricName string) (bool, error) {
	query := fmt.Sprintf("absent(%s)", metricName)
	return c.collector.getMetric(query)
}

func (p prometheusCollector) getMetric(query string) (bool, error) {
	results, err := p.client.Query(context.Background(), query, time.Now())
	if err != nil {
		logrus.WithError(err).Error("Could not query prometheus")
		return false, err
	}
	if len(results.(model.Vector)) > 0 {
		return false, err
	}
	return true, nil
}

func (c prometheusClient) ListActiveIngresses(maxIdle string) (map[string]bool, error) {
	l := logrus.WithField("method", "metrics:prometheusClient:ListActiveIngresses")
	// namespace or exported_namespace depends on your version
	exp := fmt.Sprintf("sum(rate(nginx_ingress_controller_requests{status=~\".+\"}[%s])) by (ingress, exported_namespace,namespace)", maxIdle)
	l.Infof("Looking for apps that has been idle for %s", maxIdle)
	results, err := c.collector.getIngresses(c.ctx, exp)
	if err != nil {
		l.Errorf("Could not check prometheus metrics:%v", err)
		return results, err
	}
	l.Infof("ActiveIngresses total:%d", len(results))
	return results, nil
}

func (p prometheusCollector) getIngresses(ctx context.Context, query string) (ingresses ingressResults, err error) {
	results := make(map[string]bool)

	val, err := p.client.Query(ctx, query, time.Now())
	if err != nil {
		return results, err
	}

	for _, sample := range val.(model.Vector) {
		namespace := ""
		if ns, ok := sample.Metric["exported_namespace"]; ok {
			namespace = string(ns)
		} else if ns, ok := sample.Metric["namespace"]; ok {
			namespace = string(ns)
		}
		if namespace != "" {
			if ingress, ok := sample.Metric["ingress"]; ok {
				value := float64(sample.Value)
				if value > 0 {
					// A 0 value should be considered active as well.
					// Only the absence of an app means that the its
					// inactive.
					results[fmt.Sprintf("%s/%s", namespace, ingress)] = true
				}
			}
		}
	}
	return results, nil
}
