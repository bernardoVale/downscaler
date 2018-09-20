package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func appsToSleep(active map[string]int, actual map[string]int) []string {
	return []string{"default:foo"}
}

func sleeper(ctx context.Context, poster backend.Poster) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			logrus.Info("Running sleeper again")
			// activeIngress := checkPrometheusMetrics(ctx)
			// actualIngress := checkK8sRegistry(ctx)

			dif := appsToSleep(nil, nil)

			for _, v := range dif {
				// Notify backend that sleeper will put a new app to sleep
				err := poster.Post(fmt.Sprintf("sleeping:%s", v), "sleeping")
				if err != nil {
					logrus.Errorf("Could not write sleep signal on backend. Error:%v", err)
				}
				logrus.Infof("Putting app %s to sleep", v)
				// go putToSleepOnK8s(ctx, v)
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {

	ctx := context.Background()
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient("127.0.0.1:6379", "npCYPR7uAt")

	// try to retrieve some vals
	_, err := redis.Retrieve("sleeping:default:grafana")
	must(err)

	logrus.Infoln("Starting sleeper process")
	go sleeper(ctx, redis)

	ctx.Done()
	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
