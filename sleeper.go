package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/go-redis/redis"
	"github.com/rusenask/k8s-kv/kv"
	"github.com/sirupsen/logrus"
)

func sleeper(ctx context.Context, backend backend.PosterRetriever, collector IngressCollector, kube checkPatchReceiver) {
	logrus.Infoln("Starting sleeper process")
	tick := time.NewTicker(time.Hour * 2)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			logrus.Info("Running sleeper process")
			activeIngress, err := checkPrometheusMetrics(ctx, collector)
			if err != nil {
				logrus.Errorf("Could not retrieve the list of active ingresses. Err: %v", err)
				break
			}
			allIngresses := kube.RetrieveIngresses(ctx)

			candidates := sleepCandidates(activeIngress, allIngresses)

			for _, v := range candidates {
				i := Ingress(v)
				queue := fmt.Sprintf("sleeping:%s", i.AsQueue())

				status, err := backend.Retrieve(queue)
				if err != nil {
					if err != redis.Nil && err != kv.ErrNotFound {
						logrus.Infof("Could not check the status of backend key. Err:%v", err)
						break
					}
				}

				if status == "waking_up" {
					logrus.Infof("Skipping app %v with status waking_up", i)
					break
				}
				// should check if app is waking_up before trying to put it to sleep
				// Notify backend that sleeper will put a new app to sleep
				err = backend.Post(fmt.Sprintf("sleeping:%s", i.AsQueue()), "sleeping", 0)
				if err != nil {
					logrus.Errorf("Could not write sleep signal on backend. Error:%v", err)
					break
				}
				logrus.Debugf("Putting app %s to sleep", v)
				exists := kube.CheckDeployment(ctx, i.GetName(), i.GetNamespace())
				if exists {
					logrus.Debugf("Will put %v to sleep", i)
					err := kube.PatchDeployment(i.String(), SleepAction)
					if err != nil {
						logrus.Errorf("Could not put app %v to sleep. Error %v", i, err)
						break
					}
					logrus.Infof("App %v is now sleeping :)", i)
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
