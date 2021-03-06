package main

import (
	"context"
	"sync"

	"github.com/bernardoVale/downscaler/internal/kube"
	"github.com/bernardoVale/downscaler/internal/storage"
	"github.com/bernardoVale/downscaler/internal/types"
	"github.com/sirupsen/logrus"
)

func sleepingState(searcher storage.KeySearcher) {
	keys, err := searcher.KeysByValue("downscaler:*:*", "sleeping")
	if err != nil {
		logrus.WithError(err).Error("Could not read downscaler keys")
		return
	}
	sleepingGauge.Set(float64(len(keys)))
}

func reconciliate(ctx context.Context, backend storage.PostSearcher, kube kube.GetScaler) {
	var wg sync.WaitGroup
	logger := logrus.WithField("method", "reconciliator")
	logger.Info("Starting reconciliator")

	go sleepingState(backend)
	keys, err := backend.KeysByValue("downscaler:*:*", "waking_up")

	logger.Infof("%d apps to reconciliate", len(keys))
	if err != nil {
		logger.WithError(err).Panicf("Could not get waking_up keys")
		panic(err)
	}
	wg.Add(len(keys))
	for _, key := range keys {
		go func(key string) {
			defer wg.Done()
			app, err := types.NewApp(key)
			if err != nil {
				reconciliatorErrCounter.Inc()
				logger.WithError(err).Errorf("Could not create an App representation")
				return
			}
			err = kube.Scale(app.Namespace(), app.Name(), types.ScaleUp)
			if err != nil {
				reconciliatorErrCounter.Inc()
				logger.Errorf("Failed to scale deployment. Err: %v", err)
				return
			}
			reconciliatorCounter.Inc()
			awakeWatcher(ctx, backend, kube, app)
		}(key)
	}
	logger.Info("Waiting for all apps to reconciliate")
	wg.Wait()
	logger.Info("We're good to go!")
}
