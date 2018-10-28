package main

import (
	"time"

	"github.com/bernardoVale/downscaler/internal/kube"
	"github.com/bernardoVale/downscaler/internal/metrics"
	"github.com/bernardoVale/downscaler/internal/storage"
	"github.com/bernardoVale/downscaler/internal/types"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

func sleeper(c sleeperConfig, backend storage.PosterRetriever, metrics metrics.Client,
	kube kube.PatchLister) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "sleeper",
	})

	logger.Infof("Starting sleeper process. Interval [%v]", c.sleeperInterval)
	tick := time.NewTicker(c.sleeperInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			logger.Info("Running sleeper process")
			activeIngress, err := metrics.ListActiveIngresses(c.sleepAfter)
			logger.Infof("Active Ingresses: %d", len(activeIngress))
			if err != nil {
				logger.Errorf("Could not retrieve the list of active ingresses. Err: %v", err)
				break
			}
			allIngresses := kube.List()

			candidates := sleepCandidates(activeIngress, allIngresses)

			for _, v := range candidates {
				app := types.App(v)

				status, err := backend.Retrieve(app.Key())
				if err != nil {
					if err != redis.Nil {
						logger.WithError(err).Error("Could not check the status of backend key")
						break
					}
				}
				switch status {
				case "waking_up":
					logger.WithField("app", app).Info("Skipping app with status waking_up")
					continue
				case "sleeping":
					logger.WithField("app", app).Info("Skipping app with status sleeping")
					continue
				}
				// should check if app is waking_up before trying to put it to sleep
				// Notify backend that sleeper will put a new app to sleep
				err = backend.Post(app.Key(), "sleeping", sleepingTTL)
				if err != nil {
					logger.WithError(err).Error("Could not write sleep signal on backend.")
					sleepingErr.Inc()
					break
				}
				err = kube.Scale(app.Namespace(), app.Name(), types.ScaleDown)
				if err != nil {
					logger.WithError(err).WithField("app", app).Error("Could not put app to sleep")
					sleepingErr.Inc()
					break
				}
				logger.WithField("app", app).Info("App is now sleeping :)")
				sleepingGauge.Inc()
				sleepingCounter.Inc()
			}
		}
	}
}

func sleepCandidates(active map[string]bool, all map[string]bool) []string {
	for k := range active {
		delete(all, k)
	}
	candidates := make([]string, 0)
	for k := range all {
		candidates = append(candidates, k)
	}
	logrus.WithField("method", "sleepCandidates").Infof("sleepCandidates total:%d", len(candidates))
	return candidates
}
