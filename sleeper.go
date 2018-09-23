package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func sleeper(ctx context.Context, poster backend.Poster, collector IngressCollector, kube checkReceiver) {
	tick := time.NewTicker(time.Minute * 2)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			logrus.Info("Running sleeper process")
			activeIngress := checkPrometheusMetrics(ctx, collector)
			allIngresses := kube.RetrieveIngresses(ctx)

			candidates := sleepCandidates(activeIngress, allIngresses)

			for _, v := range candidates {
				// Notify backend that sleeper will put a new app to sleep
				err := poster.Post(fmt.Sprintf("sleeping:%s", v), "sleeping")
				if err != nil {
					logrus.Errorf("Could not write sleep signal on backend. Error:%v", err)
				}
				logrus.Debugf("Putting app %s to sleep", v)
				app := strings.Split(v, "/")
				namespace := app[0]
				ingress := app[1]
				exists := kube.CheckDeployment(ctx, ingress, namespace)
				if exists {
					logrus.Infof("Will put %s/%s to sleep", namespace, ingress)
					// go putToSleepOnK8s(ctx, k)
				}
			}
		case <-ctx.Done():
			logrus.Info("Shutting down sleeper")
			return
		}
	}
}

func sleepCandidates(active map[string]int, all map[string]bool) []string {
	for k := range active {
		delete(all, k)
	}
	candidates := make([]string, 0)
	for k := range all {
		candidates = append(candidates, k)
	}
	logrus.Infof("sleepCandidates total:%d", len(candidates))
	return candidates
}
