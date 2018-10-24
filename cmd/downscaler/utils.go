package main

import (
	"errors"
	"time"

	"github.com/bernardoVale/downscaler/internal/metrics"
	"github.com/sirupsen/logrus"
)

const (
	// awake key status time to live
	// The maximum amount of time we should wait
	// for the default backend to redirect
	// enqueued requests to the original
	// backend
	awakeTTL = time.Minute * 1
	// waking-up key status time to live.
	// The maximum amount of time we should wait
	// until give up
	wakingUpTTL = time.Minute * 20

	// sleeping status time to live.
	// Apps can sleep forever ;)
	sleepingTTL = 0
)

func waitForPrometheusMetric(client metrics.Client) error {
	logger := logrus.WithField("method", "waitForPrometheusMetric")
	logger.Info("Waiting for prometheus signal before registering sleeper and wakeuper")

	tick := time.Tick(time.Second * 5)
	timeout := time.After(15 * time.Minute)

	for {
		select {
		case <-tick:
			exists, err := client.MetricExists("nginx_ingress_controller_requests")
			if err != nil {
				logger.WithError(err).Error("Could not check if metric exists")
				break
			}
			if exists {
				return nil
			}
			logger.Info("Prometheus metric nginx_ingress_controller_requests is not there. Waiting for more 5 seconds")
		case <-timeout:
			return errors.New("Timeout while waiting for prometheus metric")
		}
	}
}
