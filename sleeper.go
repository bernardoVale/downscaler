package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/go-redis/redis"
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

			for k, v := range candidates {
				i := Ingress(k)
				deployments := strings.Split(v, ",")
				for _, deployment := range deployments {
					app := strings.Split(deployment, "/")
					namespace := app[0]
					deploy := app[1]
					queue := fmt.Sprintf("sleeping:%s:%s", namespace, deploy)
					status, err := backend.Retrieve(queue)
					if err != nil {
						if err != redis.Nil {
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
					err = backend.Post(fmt.Sprintf("sleeping:%s:%s", namespace, deploy), "sleeping", 0)
					if err != nil {
						logrus.Errorf("Could not write sleep signal on backend. Error:%v", err)
						break
					}
					logrus.Debugf("Putting app %s to sleep", v)
					exists := kube.CheckDeployment(ctx, deploy, namespace)
					if exists {
						logrus.Debugf("Will put %v to sleep", i)
						err := kube.PatchDeployment(ctx, deployment, SleepAction)
						if err != nil {
							logrus.Errorf("Could not put app %v to sleep. Error %v", i, err)
							break
						}
						logrus.Infof("App %v is now sleeping :)", i)
					}
				}
			}
		case <-ctx.Done():
			logrus.Info("Shutting down sleeper")
			return
		}
	}
}

func sleepCandidates(active map[string]int, all map[string]string) map[string]string {
	for k := range active {
		delete(all, k)
	}
	// candidates := make([]string, 0)
	// for k := range all {
	// 	candidates = append(candidates, k)
	// }
	logrus.Infof("sleepCandidates total:%d", len(all))
	return all
}
