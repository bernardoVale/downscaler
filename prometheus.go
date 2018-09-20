package main

import (
	"context"
	"time"

	api "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type PrometheusClient struct {
	client prometheus.API
}

func NewPrometheusClient() PrometheusClient {
	baseClient, err := api.NewClient(api.Config{Address: "http://localhost:9090"})
	must(err)
	return PrometheusClient{client: prometheus.NewAPI(baseClient)}
}

func (c PrometheusClient) query(ctx context.Context, query string) {
	val, err := c.client.Query(ctx, query, time.Now())
	must(err)
	logrus.Info("type: %s", val.Type().String())

	vector := val.(model.Vector)

	for _, sample := range vector {
		logrus.Info("Sample:", sample.Metric.String())
		// sample.Metric.
		// for i, v := range sample.Metric.(model.LabelSet) {
		// 	logrus.Info("key=%v val=%v", i, v)
		// }
		logrus.Info("Value:", sample.Value)
	}
	// logrus.Info("data: %s", vector.String())
	// for foo, bar := range val {
	// 	logrus.Info("Foo: %v Bar: %v", foo, bar)
	// }

	// ioutil.WriteFile("query.json", data, 0640)

	// if ()

	// val.(model.Vector).
	// vector, ok := val.(model.Vector).
	//
	// vector.
	// if ok {
	// logrus.Info("Vector: %v", vector)
	// }
}
